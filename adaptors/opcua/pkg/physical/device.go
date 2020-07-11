package physical

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rancher/octopus/adaptors/opcua/api/v1alpha1"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/mqtt"
	"github.com/rancher/octopus/pkg/util/critical"
	"github.com/rancher/octopus/pkg/util/object"
)

// Device is an interface for device operations set.
type Device interface {
	// Shutdown uses to close the connection between adaptor and real(physical) device.
	Shutdown()
	// Configure uses to set up the device.
	Configure(references api.ReferencesHandler, obj *v1alpha1.OPCUADevice) error
}

// NewDevice creates a Device.
func NewDevice(log logr.Logger, meta metav1.ObjectMeta, toLimb OPCUADeviceLimbSyncer) Device {
	log.Info("Created ")
	return &opcuaDevice{
		log: log,
		instance: &v1alpha1.OPCUADevice{
			ObjectMeta: meta,
		},
		toLimb: toLimb,
	}
}

type opcuaDevice struct {
	sync.Mutex

	log         logr.Logger
	instance    *v1alpha1.OPCUADevice
	toLimb      OPCUADeviceLimbSyncer
	stop        chan struct{}
	opcuaClient *opcua.Client

	mqttClient mqtt.Client
}

func (d *opcuaDevice) Configure(references api.ReferencesHandler, device *v1alpha1.OPCUADevice) error {
	d.Lock()
	defer d.Unlock()

	var newSpec = device.Spec

	// configures MQTT opcuaClient if needed
	var staleExtension, newExtension v1alpha1.OPCUADeviceExtension
	if d.instance.Spec.Extension != nil {
		staleExtension = *d.instance.Spec.Extension
	}
	if newSpec.Extension != nil {
		newExtension = *newSpec.Extension
	}
	if !reflect.DeepEqual(staleExtension.MQTT, newExtension.MQTT) {
		if d.mqttClient != nil {
			d.mqttClient.Disconnect()
			d.mqttClient = nil
		}

		if newExtension.MQTT != nil {
			var cli, err = mqtt.NewClient(*newExtension.MQTT, object.GetControlledOwnerObjectReference(device), references)
			if err != nil {
				return errors.Wrap(err, "failed to create MQTT opcuaClient")
			}

			err = cli.Connect()
			if err != nil {
				return errors.Wrap(err, "failed to connect MQTT broker")
			}
			d.mqttClient = cli
		}
	}

	// configures OPC-UA client
	if !reflect.DeepEqual(d.instance.Spec.Protocol, newSpec.Protocol) || !reflect.DeepEqual(d.instance.Spec.Parameters, newSpec.Parameters) {
		if d.opcuaClient != nil {
			if err := d.opcuaClient.Close(); err != nil {
				d.log.Error(err, "Error closing OPC-UA connection")
			}
			d.opcuaClient = nil
		}

		var client, err = NewOPCUAClient(newSpec.Protocol, newSpec.Parameters.Timeout.Duration, references)
		if err != nil {
			return errors.Wrap(err, "failed to create OPC-UA client")
		}
		d.opcuaClient = client
	}

	return d.refresh(newSpec)
}

func (d *opcuaDevice) Shutdown() {
	d.Lock()
	defer d.Unlock()

	d.stopSubscribe()
	if d.opcuaClient != nil {
		if err := d.opcuaClient.Close(); err != nil {
			d.log.Error(err, "Error closing OPC-UA connection")
		}
		d.opcuaClient = nil
	}
	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
	}
	d.log.Info("Shutdown")
}

// refresh refreshes the status with new spec.
func (d *opcuaDevice) refresh(newSpec v1alpha1.OPCUADeviceSpec) error {
	var status = d.instance.Status
	var staleSpec = d.instance.Spec
	if !reflect.DeepEqual(staleSpec.Properties, newSpec.Properties) {
		d.stopSubscribe()

		// configures properties
		var specProps = newSpec.Properties
		var statusProps = make([]v1alpha1.OPCUADeviceStatusProperty, 0, len(specProps))
		for _, prop := range specProps {
			var value string
			if prop.ReadOnly {
				// TODO need to read property at first?
			} else {
				// the written value should be consistent with the read value,
				// so there is no need to fetch it once more.
				if err := d.writeProperty(prop.Type, prop.Visitor, prop.Value); err != nil {
					return errors.Wrapf(err, "failed to write property %s", prop.Name)
				}
				value = prop.Value
				d.log.V(4).Info("Write property", "property", prop.Name, "type", prop.Type)
			}
			statusProps = append(statusProps, v1alpha1.OPCUADeviceStatusProperty{
				Name:      prop.Name,
				Value:     value,
				Type:      prop.Type,
				UpdatedAt: now(),
			})
		}
		status = v1alpha1.OPCUADeviceStatus{Properties: statusProps}
	}

	// subscribed in backend
	if err := d.startSubscribe(newSpec.Parameters.SyncInterval.Duration, newSpec.Properties); err != nil {
		return errors.Wrap(err, "failed to subscribing")
	}

	// records
	d.instance.Spec = newSpec
	d.instance.Status = status
	return d.sync()
}

// writeProperty writes data of a property to the corresponding OPC-UA node.
func (d *opcuaDevice) writeProperty(dataType v1alpha1.OPCUADevicePropertyType, visitor v1alpha1.OPCUADevicePropertyVisitor, value string) error {
	var data, err = StringToVariant(dataType, value)
	if err != nil {
		return errors.Wrapf(err, "failed to convert %s string to %s variant", value, dataType)
	}

	id, err := ua.ParseNodeID(visitor.NodeID)
	if err != nil {
		return errors.Wrapf(err, "failed to parse node ID %s", visitor.NodeID)
	}
	var req = &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      id,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask: ua.DataValueValue,
					Value:        data,
				},
			},
		},
	}
	_, err = d.opcuaClient.Write(req)
	if err != nil {
		return errors.Wrapf(err, "failed to write")
	}
	return nil
}

// subscribe is blocked, it is used to watch the notification from OPC-UA server
// and update the opcua device status.
func (d *opcuaDevice) subscribe(ctx context.Context, notifyCh chan *opcua.PublishNotificationData) {
	d.log.Info("Subscribing")
	defer func() {
		d.log.Info("Finished subscription")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case res := <-notifyCh:
			if res.Error != nil {
				// TODO give a way to feedback this to limb.
				d.log.Error(res.Error, "Received error from subscription")
				continue
			}

			switch v := res.Value.(type) {
			case *ua.DataChangeNotification:
				d.Lock()
				func() {
					defer d.Unlock()

					var statusProps = d.instance.Status.Properties
					for _, item := range v.MonitoredItems {
						var idx = int(item.ClientHandle)
						if idx >= len(statusProps) {
							continue
						}
						var prop = statusProps[idx]
						var variant = item.Value.Value
						var value = VariantToString(variant.Type(), variant)
						var propType = typeMap[variant.Type()]
						d.log.V(4).Info("Received property", "property", item.ClientHandle, "type", propType)
						statusProps[idx] = v1alpha1.OPCUADeviceStatusProperty{
							Name:      prop.Name,
							Value:     value,
							Type:      propType,
							UpdatedAt: now(),
						}
					}
					d.instance.Status.Properties = statusProps
					if err := d.sync(); err != nil {
						d.log.Error(err, "failed to sync")
					}
				}()
			default:
				d.log.V(4).Info(fmt.Sprintf("Received unknown property %+v", res.Value))
			}
		}
	}
}

func (d *opcuaDevice) stopSubscribe() {
	if d.stop != nil {
		close(d.stop)
		d.stop = nil
	}
}

func (d *opcuaDevice) startSubscribe(subscribeInterval time.Duration, properties []v1alpha1.OPCUADeviceProperty) error {
	if d.stop == nil {
		d.stop = make(chan struct{})

		// creates subscription
		var notifyCh = make(chan *opcua.PublishNotificationData)
		var sub, err = d.opcuaClient.Subscribe(&opcua.SubscriptionParameters{Interval: subscribeInterval}, notifyCh)
		if err != nil {
			return errors.Wrap(err, "failed to create subscription")
		}
		d.log.Info("Created subscription", "id", sub.SubscriptionID)

		// creates monitoring request for all properties
		for idx, prop := range properties {
			var id, err = ua.ParseNodeID(prop.Visitor.NodeID)
			if err != nil {
				return errors.Wrapf(err, "failed to parse node ID %s", prop.Visitor.NodeID)
			}

			var handle = uint32(idx)
			var miCreateRequest = opcua.NewMonitoredItemCreateRequestWithDefaults(id, ua.AttributeIDValue, handle)
			res, err := sub.Monitor(ua.TimestampsToReturnBoth, miCreateRequest)
			if err != nil {
				return errors.Wrapf(err, "error monitoring property %s", prop.Name)
			}
			if res.Results[0].StatusCode != ua.StatusOK {
				return errors.Errorf("failed to monitor property %s", prop.Name)
			}
			d.log.V(4).Info("Monitored property", "property", prop.Name)
		}

		// subscribes
		var ctx = critical.Context(d.stop, func() {
			var err = sub.Cancel()
			if err != nil {
				d.log.Error(err, "Failed to cancel subscription")
			}
		})
		go sub.Run(ctx)
		d.log.Info("Running subscription", "id", sub.SubscriptionID)

		go d.subscribe(ctx, notifyCh)
	}

	return nil
}

// sync combines all synchronization operations.
func (d *opcuaDevice) sync() error {
	if err := d.toLimb(d.instance); err != nil {
		return err
	}
	if d.mqttClient != nil {
		if err := d.mqttClient.Publish(mqtt.PublishMessage{Payload: d.instance.Status}); err != nil {
			return err
		}
	}
	d.log.V(1).Info("Synced")
	return nil
}

func now() *metav1.Time {
	var ret = metav1.Now()
	return &ret
}
