package physical

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/pkg/errors"

	"github.com/rancher/octopus/adaptors/opcua/api/v1alpha1"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/mqtt"
	"github.com/rancher/octopus/pkg/util/critical"
	"github.com/rancher/octopus/pkg/util/object"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Device interface {
	Configure(references api.ReferencesHandler, obj v1alpha1.OPCUADevice) error
	Shutdown()
}

func NewDevice(log logr.Logger, name types.NamespacedName, handler DataHandler) Device {
	return &device{
		log:     log,
		name:    name,
		handler: handler,
	}
}

type device struct {
	sync.Mutex

	stop chan struct{}

	log     logr.Logger
	name    types.NamespacedName
	handler DataHandler

	spec   v1alpha1.OPCUADeviceSpec
	status v1alpha1.OPCUADeviceStatus

	client     *opcua.Client
	mqttClient mqtt.Client
}

func (d *device) Configure(references api.ReferencesHandler, obj v1alpha1.OPCUADevice) error {
	deviceSpec := d.spec
	d.spec = obj.Spec

	// configure protocol config and parameters
	if !reflect.DeepEqual(d.spec.ProtocolConfig, deviceSpec.ProtocolConfig) || !reflect.DeepEqual(d.spec.Parameters, deviceSpec.Parameters) {
		if err := d.on(); err != nil {
			return err
		}
	}

	if !reflect.DeepEqual(d.spec.Extension, deviceSpec.Extension) {
		if d.mqttClient != nil {
			d.mqttClient.Disconnect()
			d.mqttClient = nil
		}

		if d.spec.Extension.MQTT != nil {
			var cli, err = mqtt.NewClient(*d.spec.Extension.MQTT, object.GetControlledOwnerObjectReference(&obj), references)
			if err != nil {
				return errors.Wrap(err, "failed to create MQTT client")
			}

			err = cli.Connect()
			if err != nil {
				return errors.Wrap(err, "failed to connect MQTT broker")
			}
			d.mqttClient = cli
		}
	}

	// configure device properties
	properties := d.spec.Properties
	for _, property := range properties {
		if property.ReadOnly {
			continue
		}
		if err := d.writeProperty(property.DataType, property.Visitor, property.Value); err != nil {
			d.log.Error(err, "Error write property", "property", property)
			continue
		}
		d.log.Info("Write property", "property", property)
	}
	return nil
}

// write data of a property to the corresponding opc-ua node
func (d *device) writeProperty(dataType v1alpha1.PropertyDataType, visitor v1alpha1.PropertyVisitor, value string) error {
	data, err := StringToVariant(dataType, value)
	if err != nil {
		d.log.Error(err, "Error converting writing data", "data", value)
		return err
	}
	id, err := ua.ParseNodeID(visitor.NodeID)
	if err != nil {
		return err
	}
	req := &ua.WriteRequest{
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
	resp, err := d.client.Write(req)
	if err != nil {
		d.log.Error(err, "Write failed")
		return err
	}
	d.log.Info("Writing success", "response", resp.Results[0])
	return nil
}

func (d *device) on() error {
	if d.stop != nil {
		close(d.stop)
	}
	d.stop = make(chan struct{})

	// close old client
	if d.client != nil {
		if err := d.client.Close(); err != nil {
			d.log.Error(err, "Fail to close old opc-ua client")
		}
	}

	// create client
	var err error
	spec := d.spec
	d.client, err = newClient(spec.ProtocolConfig, spec.Parameters.Timeout.Duration)
	if err != nil {
		d.log.Error(err, "Fail to create opc-ua client")
		return err
	}

	// connect to device
	ctx := critical.Context(d.stop)
	if err := d.client.Connect(ctx); err != nil {
		d.log.Error(err, "Error connecting to device")
		return err
	}

	d.subscribe(ctx, d.spec)
	return nil
}

func (d *device) subscribe(ctx context.Context, spec v1alpha1.OPCUADeviceSpec) {
	notifyCh := make(chan *opcua.PublishNotificationData)

	sub, err := d.client.Subscribe(&opcua.SubscriptionParameters{Interval: d.spec.Parameters.SyncInterval.Duration}, notifyCh)
	if err != nil {
		d.log.Error(err, "Subscription error")
	}
	d.log.Info("Created subscription", "id", sub.SubscriptionID)

	go sub.Run(ctx) // start Publish loop

	properties := spec.Properties
	for i, property := range properties {
		d.monitorProperty(i, property, sub)
	}

	go d.receiveNotification(ctx, notifyCh, properties)
}

func (d *device) Shutdown() {
	d.Lock()
	defer d.Unlock()

	if d.stop != nil {
		close(d.stop)
	}

	// close OPC-UA client
	if d.client != nil {
		if err := d.client.Close(); err != nil {
			d.log.Error(err, "Error closing connection")
		}
	}

	// close MQTT connection
	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
	}

	d.log.Info("Closed connection")
}

// monitor data of a property from its corresponding opc-ua node
func (d *device) monitorProperty(idx int, property v1alpha1.DeviceProperty, sub *opcua.Subscription) {
	node := property.Visitor.NodeID

	id, err := ua.ParseNodeID(node)
	if err != nil {
		d.log.Error(err, "Error parsing nodeID")
	}

	// index of the array is the client handle for the monitoring item
	handle := uint32(idx)
	miCreateRequest := opcua.NewMonitoredItemCreateRequestWithDefaults(id, ua.AttributeIDValue, handle)
	res, err := sub.Monitor(ua.TimestampsToReturnBoth, miCreateRequest)
	if err != nil || res.Results[0].StatusCode != ua.StatusOK {
		d.log.Error(err, "")
		return
	}
	d.log.Info("Monitoring property", "property", property)
}

// update the properties from physical device to status
func (d *device) receiveNotification(ctx context.Context, notifyCh chan *opcua.PublishNotificationData, properties []v1alpha1.DeviceProperty) {
	// read from subscription's notification channel until ctx is cancelled
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-notifyCh:
			if res.Error != nil {
				d.log.Error(res.Error, "")
				continue
			}

			switch x := res.Value.(type) {
			case *ua.DataChangeNotification:
				for _, item := range x.MonitoredItems {
					property := properties[item.ClientHandle]
					value := item.Value.Value
					typeID := value.Type()
					data := VariantToString(typeID, value)
					property.DataType = typeMap[typeID]
					d.updateDeviceStatus(&property, data)
					d.log.V(6).Info("MonitoredItem with client", "handle", item.ClientHandle, "value", data)
				}
				d.handler(d.name, d.status)
				d.log.Info("Sync opc-ua device status", "properties", d.status.Properties)

				// pub updated status to the MQTT broker
				if d.mqttClient != nil {
					var status = d.status.DeepCopy()
					if err := d.mqttClient.Publish(mqtt.PublishMessage{Payload: status}); err != nil {
						d.log.Error(err, "Failed to publish MQTT message")
					}
				}
				d.log.V(2).Info("Success pub device status to the MQTT Broker", d.status.Properties)
			default:
				d.log.Info("what's this publish result? ", "value", res.Value)
			}
		}
	}
}

func (d *device) updateDeviceStatus(property *v1alpha1.DeviceProperty, data string) {
	d.Lock()
	defer d.Unlock()
	newProperty := v1alpha1.StatusProperties{
		Name:      property.Name,
		DataType:  property.DataType,
		Value:     data,
		UpdatedAt: metav1.Time{Time: time.Now()},
	}
	found := false
	for i, p := range d.status.Properties {
		if p.Name == newProperty.Name {
			d.status.Properties[i] = newProperty
			found = true
			break
		}
	}
	if !found {
		d.status.Properties = append(d.status.Properties, newProperty)
	}
}

func newClient(config *v1alpha1.OPCUAProtocolConfig, timeout time.Duration) (*opcua.Client, error) {
	url := config.URL
	endpoints, err := opcua.GetEndpoints(url)
	if err != nil {
		return nil, err
	}
	policy := config.SecurityPolicy
	mode := config.SecurityMode
	ep := opcua.SelectEndpoint(endpoints, policy, ua.MessageSecurityModeFromString(mode))

	opts := []opcua.Option{
		opcua.RequestTimeout(timeout),
		opcua.SecurityPolicy(policy),
		opcua.SecurityModeString(mode),
		// TODO read CA file from the req.References
		opcua.CertificateFile(config.CertificateFile),
		opcua.PrivateKeyFile(config.PrivateKeyFile),
		opcua.AuthAnonymous(),
		opcua.AuthUsername(config.UserName, config.Password),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}
	return opcua.NewClient(url, opts...), nil
}
