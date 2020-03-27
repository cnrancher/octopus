package physical

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/rancher/octopus/adaptors/opcua/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/critical"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Device interface {
	Configure(spec v1alpha1.OPCUADeviceSpec)
	On(spec v1alpha1.OPCUADeviceSpec)
	Shutdown()
}

func NewDevice(log logr.Logger, name types.NamespacedName, handler DataHandler, syncInterval time.Duration, config *v1alpha1.OPCUAProtocolConfig) Device {
	return &device{
		log:          log,
		name:         name,
		handler:      handler,
		syncInterval: syncInterval,
		config:       config,
	}
}

type device struct {
	sync.Mutex

	stop chan struct{}

	log     logr.Logger
	name    types.NamespacedName
	handler DataHandler

	status       v1alpha1.OPCUADeviceStatus
	syncInterval time.Duration
	config       *v1alpha1.OPCUAProtocolConfig
	client       *opcua.Client
}

func (d *device) Configure(spec v1alpha1.OPCUADeviceSpec) {
	properties := spec.Properties
	for _, property := range properties {
		if property.ReadOnly {
			continue
		}
		if err := d.writeProperty(property.DataType, property.Visitor, property.Value); err != nil {
			d.log.Error(err, "Error write property", "property", property.Name)
		}
	}
}

// write data of a property to coil register or holding register
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

func (d *device) On(spec v1alpha1.OPCUADeviceSpec) {
	if d.stop != nil {
		close(d.stop)
	}
	d.stop = make(chan struct{})

	ctx := critical.Context(d.stop)

	d.newClient(d.config)
	if err := d.client.Connect(ctx); err != nil {
		d.log.Error(err, "Error connecting to device")
	}
	d.subscribe(ctx, spec)
}

func (d *device) subscribe(ctx context.Context, spec v1alpha1.OPCUADeviceSpec) {
	notifyCh := make(chan *opcua.PublishNotificationData)

	sub, err := d.client.Subscribe(&opcua.SubscriptionParameters{
		Interval: d.syncInterval,
	}, notifyCh)
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
	if err := d.client.Close(); err != nil {
		d.log.Error(err, "Error closing connection")
	}
	d.log.Info("Closed connection")
}

// read data of a property from its corresponding register
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
	}
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
					d.log.Info("MonitoredItem with client", "handle", item.ClientHandle, "value", data)
				}
				d.handler(d.name, d.status)
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

func (d *device) newClient(config *v1alpha1.OPCUAProtocolConfig) {
	url := config.URL
	endpoints, err := opcua.GetEndpoints(url)
	if err != nil {
		d.log.Error(err, "Error get endpoints")
	}
	policy := config.SecurityPolicy
	mode := config.SecurityMode
	ep := opcua.SelectEndpoint(endpoints, policy, ua.MessageSecurityModeFromString(mode))

	opts := []opcua.Option{
		opcua.SecurityPolicy(policy),
		opcua.SecurityModeString(mode),
		// TODO read CA file from the container
		opcua.CertificateFile(config.CertificateFile),
		opcua.PrivateKeyFile(config.PrivateKeyFile),
		opcua.AuthAnonymous(),
		opcua.AuthUsername(config.UserName, config.Password),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}
	d.client = opcua.NewClient(url, opts...)
}
