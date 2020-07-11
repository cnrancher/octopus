package physical

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/ua"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rancher/octopus/adaptors/opcua/api/v1alpha1"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/converter"
	"github.com/rancher/octopus/pkg/util/log/logflag"
)

type DataHandler func(name types.NamespacedName, status v1alpha1.OPCUADeviceStatus)

// OPCUADeviceLimbSyncer is used to sync opcua opcuaDevice to limb.
type OPCUADeviceLimbSyncer func(in *v1alpha1.OPCUADevice) error

// NewOPCUAClient creates a opcua.Client
func NewOPCUAClient(protocol v1alpha1.OPCUADeviceProtocol, timeout time.Duration, references api.ReferencesHandler) (*opcua.Client, error) {
	if logflag.GetLogVerbosity() > 4 {
		// setup opcua debug log
		debug.Enable = true
		debug.Logger = log.New(os.Stdout, "opcua.client ", log.LstdFlags)
	}

	// discoveries the available endpoint for the server
	var endpoints, err = func(endpoint string) ([]*ua.EndpointDescription, error) {
		var c = opcua.NewClient(endpoint)
		var ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := c.Dial(ctx); err != nil {
			return nil, err
		}
		defer c.Close()
		var res, err = c.GetEndpoints()
		if err != nil {
			return nil, err
		}
		return res.Endpoints, nil
	}(protocol.Endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get OPC-UA endpoint")
	}

	var policy = string(protocol.SecurityPolicy)
	var mode = string(protocol.SecurityMode)
	var ep = opcua.SelectEndpoint(endpoints, policy, ua.MessageSecurityModeFromString(mode))
	var options = []opcua.Option{
		opcua.RequestTimeout(timeout),
		opcua.SecurityPolicy(policy),
		opcua.SecurityModeString(mode),
	}

	if protocol.BasicAuth != nil {
		var basicAuthSpec = protocol.BasicAuth

		var username string
		if basicAuthSpec.Username != "" {
			username = basicAuthSpec.Username
		} else if ref := basicAuthSpec.UsernameRef; ref != nil {
			if references == nil {
				return nil, errors.Errorf("references handler is nil")
			}
			username = converter.UnsafeBytesToString(references.GetData(ref.Name, ref.Item))
		}

		var password string
		if basicAuthSpec.Password != "" {
			password = basicAuthSpec.Password
		} else if ref := basicAuthSpec.PasswordRef; ref != nil {
			if references == nil {
				return nil, errors.Errorf("references handler is nil")
			}
			password = converter.UnsafeBytesToString(references.GetData(ref.Name, ref.Item))
		}

		if username == "" || password == "" {
			return nil, errors.Errorf("illegal basic auth account as blank username or password")
		}
		options = append(options,
			opcua.AuthUsername(username, password),
		)
	} else {
		options = append(options,
			opcua.AuthAnonymous(),
			opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
		)
	}

	if protocol.TLSConfig != nil {
		var tlsConfigSpec = protocol.TLSConfig

		var certEncodedPEM []byte
		if tlsConfigSpec.CertFilePEM != "" {
			certEncodedPEM = converter.UnsafeStringToBytes(tlsConfigSpec.CertFilePEM)
		} else if ref := tlsConfigSpec.CertFilePEMRef; ref != nil {
			if references == nil {
				return nil, errors.Errorf("references handler is nil")
			}
			certEncodedPEM = references.GetData(ref.Name, ref.Item)
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
				return nil, errors.Errorf("references handler is nil")
			}
			keyEncodedPEM = references.GetData(ref.Name, ref.Item)
		}
		key, err := decodeKeyPEM(keyEncodedPEM)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get private key from key PEM content")
		}
		options = append(options,
			opcua.PrivateKey(key),
		)
	}

	var ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var client = opcua.NewClient(protocol.Endpoint, options...)
	if err := client.Connect(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to connect to OPC-UA endpoint")
	}
	return client, nil
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
