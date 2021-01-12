package physical

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	stdlog "log"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/ua"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/rancher/octopus/adaptors/opcua/api/v1alpha1"
	"github.com/rancher/octopus/adaptors/opcua/pkg/metadata"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/log"
	"github.com/rancher/octopus/pkg/adaptor/socket/handler"
	"github.com/rancher/octopus/pkg/util/converter"
	"github.com/rancher/octopus/pkg/util/critical"
	"github.com/rancher/octopus/pkg/util/log/logflag"
)

// OPCUADataValueNotificationHandler specifies to handle the changed from a UA DataValue subscription.
type OPCUADataValueNotificationHandler func(*ua.DataValue)

type subscription struct {
	handler  OPCUADataValueNotificationHandler
	uaNodeID *ua.NodeID
	nodeID   string
}

// OPCUAClient is a wrapper to handle the real opcua.Client.
type OPCUAClient struct {
	log                   logr.Logger
	cli                   *opcua.Client
	subs                  map[uint32]subscription
	subPublishingInterval time.Duration

	subCli *opcua.Subscription
}

// WriteDataValue writes UA data value to given UA node ID.
func (c *OPCUAClient) WriteDataValue(nodeID string, value *ua.DataValue) error {
	var uaNodeID, err = ua.ParseNodeID(nodeID)
	if err != nil {
		return errors.Wrapf(err, "failed to parse UA node ID %s", nodeID)
	}
	var writeReq = &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      uaNodeID,
				AttributeID: ua.AttributeIDValue,
				Value:       value,
			},
		},
	}

	writeResp, err := c.cli.Write(writeReq)
	if err != nil {
		return errors.Wrapf(err, "failed to write UA node %s", nodeID)
	}
	if len(writeResp.Results) == 0 {
		return errors.Errorf("failed to write UA node %s as the result is empty", nodeID)
	}
	if statusCode := writeResp.Results[0]; statusCode != ua.StatusOK {
		return errors.Errorf("failed to write UA node %s as the response status is %s", nodeID, statusCode.Error())
	}

	c.log.V(4).Info("Write UA data value", "nodeID", nodeID)
	return nil
}

// ReadDataValue reads UA data value from given UA node ID.
func (c *OPCUAClient) ReadDataValue(nodeID string) (*ua.DataValue, error) {
	var uaNodeID, err = ua.ParseNodeID(nodeID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse UA node ID %s", nodeID)
	}
	var readReq = &ua.ReadRequest{
		NodesToRead: []*ua.ReadValueID{
			{
				NodeID: uaNodeID,
			},
		},
	}

	readResp, err := c.cli.Read(readReq)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read UA node %s", nodeID)
	}
	if len(readResp.Results) == 0 {
		return nil, errors.Errorf("failed to read UA node %s as the result is empty", nodeID)
	}
	var result = readResp.Results[0]
	if statusCode := result.Status; statusCode != ua.StatusOK {
		return nil, errors.Errorf("failed to read UA node %s as the response status is %s", nodeID, statusCode.Error())
	}

	c.log.V(4).Info("Read UA data value", "nodeID", nodeID)
	return result, nil
}

// RegisterDataValueSubscription registers a subscriptions of the UA data value matched given UA node ID.
func (c *OPCUAClient) RegisterDataValueSubscription(nodeID string, handler OPCUADataValueNotificationHandler) error {
	var uaNodeID, err = ua.ParseNodeID(nodeID)
	if err != nil {
		return errors.Wrapf(err, "failed to parse UA node ID %s", nodeID)
	}

	// records notification
	c.subs[uint32(len(c.subs))] = subscription{
		nodeID:   nodeID,
		uaNodeID: uaNodeID,
		handler:  handler,
	}

	c.log.V(4).Info("Register to subscribe UA data value", "nodeID", nodeID)
	return nil
}

func (c *OPCUAClient) StartSubscriptions(stopChan chan struct{}) (berr error) {
	if len(c.subs) == 0 {
		return nil
	}

	var subNotificationChan = make(chan *opcua.PublishNotificationData)
	var subCli, err = c.cli.Subscribe(
		&opcua.SubscriptionParameters{
			Interval: c.subPublishingInterval,
		},
		subNotificationChan,
	)
	if err != nil {
		return errors.Wrap(err, "failed to start UA subscription")
	}
	defer func() {
		// just stop the subscription if error occurs
		if berr != nil {
			_ = subCli.Cancel()
		}
	}()
	var logger = c.log.WithValues("subscriptionID", subCli.SubscriptionID)
	logger.Info("Created a UA subscription", "publishingInterval", c.subPublishingInterval)

	for clientHandler, sub := range c.subs {
		var monitorReq = opcua.NewMonitoredItemCreateRequestWithDefaults(sub.uaNodeID, ua.AttributeIDValue, clientHandler)
		var monitorResp, err = subCli.Monitor(ua.TimestampsToReturnNeither, monitorReq)
		if err != nil {
			return errors.Wrapf(err, "failed to monitor UA node %s", sub.nodeID)
		}

		if len(monitorResp.Results) == 0 {
			return errors.Wrapf(err, "failed to subscribe UA node %s as the result is empty", sub.nodeID)
		}

		var result = monitorResp.Results[0]
		if statusCode := result.StatusCode; statusCode != ua.StatusOK {
			return errors.Errorf("failed to subscribe UA node %s as the response status is %s", sub.nodeID, statusCode.Error())
		}
	}

	// receives subscription
	go func() {
		defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

		logger.V(4).Info("Start UA subscription receiving ")
		subCli.Run(critical.Context(stopChan))
		logger.V(4).Info("Finished UA subscription receiving ")
	}()
	// processes subscription
	go func() {
		defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

		logger.V(4).Info("Start UA subscription processing")
		defer func() {
			logger.V(4).Info("Finished UA subscription processing")
		}()

		for {
			select {
			case <-stopChan:
				return
			case resp, ok := <-subNotificationChan:
				if !ok {
					return
				}
				if err := resp.Error; err != nil {
					// TODO give a way to feedback this to limb.
					logger.Error(err, "Received error from UA subscription")
					continue
				}

				switch val := resp.Value.(type) {
				case *ua.DataChangeNotification:
					for _, item := range val.MonitoredItems {
						if sub, exist := c.subs[item.ClientHandle]; exist {
							if h := sub.handler; h != nil {
								h(item.Value)
							}
						}
					}
				}
			}
		}
	}()

	c.subCli = subCli
	return
}

func (c *OPCUAClient) StopSubscriptions() {
	if len(c.subs) == 0 || c.subCli == nil {
		return
	}
	var logger = c.log.WithValues("subscriptionID", c.subCli.SubscriptionID)

	// cancels
	var err = c.subCli.Cancel()
	if err != nil {
		logger.Error(err, "Failed to stop UA subscription")
	}
	c.subCli = nil

	// cleans
	for clientHandler := range c.subs {
		delete(c.subs, clientHandler)
	}

	return
}

func (c *OPCUAClient) Close() error {
	var err = c.cli.Close()
	c.log.V(4).Info("Closed")
	return err
}

func NewOPCUAClient(protocol v1alpha1.OPCUADeviceProtocol, references api.ReferencesHandler) (*OPCUAClient, error) {
	if logflag.GetLogVerbosity() > 4 {
		// setup opcua debug log
		debug.Enable = true
		debug.Logger = stdlog.New(os.Stdout, "opcua.underlay.client ", stdlog.LstdFlags)
	}

	var logger = log.WithName("opcua.client").WithValues("endpoint", protocol.Endpoint)

	// discovers the endpoint which matches the security policy and security mode.
	logger.V(2).Info("Discovering OPC-UA server")
	var endpoint, err = func() (*ua.EndpointDescription, error) {
		var connectCtx, cancelCtx = context.WithTimeout(context.Background(), protocol.GetConnectTimeout())
		defer cancelCtx()
		var cli = opcua.NewClient(protocol.Endpoint)
		if err := cli.Dial(connectCtx); err != nil {
			return nil, errors.Wrap(err, "failed to dial OPC-UA endpoint")
		}
		defer cli.Close()

		var availableEntrances, err = cli.GetEndpoints()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get available entrances of OPC-UA endpoint")
		}
		logger.V(2).Info("Get available OPC-UA endpoints")

		var ep = opcua.SelectEndpoint(
			availableEntrances.Endpoints,
			string(protocol.SecurityPolicy),
			ua.MessageSecurityModeFromString(string(protocol.SecurityMode)),
		)
		if ep == nil {
			return nil, errors.New("failed to get matched entrance of OPC-UA endpoint")
		}
		logger.V(2).Info("Select matched OPC-UA endpoint")
		return ep, nil
	}()
	if err != nil {
		return nil, err
	}

	var options = []opcua.Option{
		opcua.RequestTimeout(protocol.GetRequestTimeout()),
		opcua.Lifetime(protocol.GetLifetime()),
	}

	// configures authentication
	if protocol.BasicAuth != nil {
		var basicAuthSpec = protocol.BasicAuth

		var username string
		if basicAuthSpec.Username != "" {
			username = basicAuthSpec.Username
		} else if ref := basicAuthSpec.UsernameRef; ref != nil {
			if references == nil {
				return nil, errors.New("references handler is nil")
			}
			username = converter.UnsafeBytesToString(references.GetData(ref.Name, ref.Item))
		} else {
			return nil, errors.New("illegal basic authentication as the username is blank")
		}

		var password string
		if basicAuthSpec.Password != "" {
			password = basicAuthSpec.Password
		} else if ref := basicAuthSpec.PasswordRef; ref != nil {
			if references == nil {
				return nil, errors.New("references handler is nil")
			}
			password = converter.UnsafeBytesToString(references.GetData(ref.Name, ref.Item))
		} else {
			return nil, errors.New("illegal basic authentication as the password is blank")
		}

		options = append(options,
			opcua.AuthUsername(username, password),
			opcua.SecurityFromEndpoint(endpoint, ua.UserTokenTypeUserName),
		)
		logger.V(3).Info("Use basic client authentication")
	} else if protocol.CertificateAuth != nil {
		var certAuthSpec = protocol.CertificateAuth

		var certEncodedPEM []byte
		if certAuthSpec.CertFilePEM != "" {
			certEncodedPEM = converter.UnsafeStringToBytes(certAuthSpec.CertFilePEM)
		} else if ref := certAuthSpec.CertFilePEMRef; ref != nil {
			if references == nil {
				return nil, errors.New("references handler is nil")
			}
			certEncodedPEM = references.GetData(ref.Name, ref.Item)
		} else {
			return nil, errors.New("illegal certificate authentication as the certificate is blank")
		}

		var certPEM, err = decodeCertificatePEM(certEncodedPEM)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get certificate from cert PEM content")
		}
		options = append(options,
			opcua.AuthCertificate(certPEM),
			opcua.SecurityFromEndpoint(endpoint, ua.UserTokenTypeCertificate),
		)
		logger.V(3).Info("Use certificate client authentication")
	} else if protocol.IssuedTokenAuth != nil {
		var issuedTokenSpec = protocol.IssuedTokenAuth

		var token []byte
		if issuedTokenSpec.Token != "" {
			token = converter.UnsafeStringToBytes(issuedTokenSpec.Token)
		} else if ref := issuedTokenSpec.TokenRef; ref != nil {
			if references == nil {
				return nil, errors.New("references handler is nil")
			}
			token = references.GetData(ref.Name, ref.Item)
		} else {
			return nil, errors.New("illegal issued token authentication as the token is blank")
		}

		options = append(options,
			opcua.AuthIssuedToken(token),
			opcua.SecurityFromEndpoint(endpoint, ua.UserTokenTypeIssuedToken),
		)
		logger.V(3).Info("Use issued token client authentication")
	} else {
		options = append(options,
			opcua.AuthAnonymous(),
			opcua.SecurityFromEndpoint(endpoint, ua.UserTokenTypeAnonymous),
		)
		logger.V(3).Info("Use as anonymous")
	}

	// configures connection security
	if protocol.TLSConfig != nil {
		var tlsConfigSpec = protocol.TLSConfig

		var certEncodedPEM []byte
		if tlsConfigSpec.CertFilePEM != "" {
			certEncodedPEM = converter.UnsafeStringToBytes(tlsConfigSpec.CertFilePEM)
		} else if ref := tlsConfigSpec.CertFilePEMRef; ref != nil {
			if references == nil {
				return nil, errors.New("references handler is nil")
			}
			certEncodedPEM = references.GetData(ref.Name, ref.Item)
		} else {
			return nil, errors.New("illegal TLS configuration as the certificate is blank")
		}
		var certPEM, err = decodeCertificatePEM(certEncodedPEM)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get certificate from cert PEM content")
		}
		options = append(options,
			opcua.Certificate(certPEM),
		)

		var keyEncodedPEM []byte
		if tlsConfigSpec.KeyFilePEM != "" {
			keyEncodedPEM = converter.UnsafeStringToBytes(tlsConfigSpec.KeyFilePEM)
		} else if ref := tlsConfigSpec.KeyFilePEMRef; ref != nil {
			if references == nil {
				return nil, errors.New("references handler is nil")
			}
			keyEncodedPEM = references.GetData(ref.Name, ref.Item)
		} else {
			return nil, errors.New("illegal TLS configuration as the private key is blank")
		}
		key, err := decodeKeyPEM(keyEncodedPEM)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get private key from key PEM content")
		}
		options = append(options,
			opcua.PrivateKey(key),
		)

		logger.V(3).Info("Use TLS/SSL configuration")
	}

	// connects
	logger.Info("Connecting matched OPC-UA endpoint")
	var connectCtx, cancelCtx = context.WithTimeout(context.Background(), protocol.GetConnectTimeout())
	defer cancelCtx()
	var client = opcua.NewClient(protocol.Endpoint, options...)
	if err := client.Connect(connectCtx); err != nil {
		return nil, errors.Wrap(err, "failed to connect to OPC-UA endpoint")
	}

	return &OPCUAClient{
		log:                   logger,
		cli:                   client,
		subs:                  make(map[uint32]subscription),
		subPublishingInterval: protocol.GetPublishingInterval(),
	}, nil
}

func decodeCertificatePEM(encodedPEM []byte) ([]byte, error) {
	var block, _ = pem.Decode(encodedPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.Errorf("failed to decode PEM block with certificate")
	}
	return block.Bytes, nil
}

func decodeKeyPEM(encodedPEM []byte) (*rsa.PrivateKey, error) {
	var block, _ = pem.Decode(encodedPEM)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.Errorf("failed to decode PEM block with private key")
	}
	var pk, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse private key")
	}
	return pk, nil
}
