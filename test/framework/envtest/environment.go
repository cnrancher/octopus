// +build test

package envtest

import (
	"k8s.io/client-go/rest"

	"github.com/rancher/octopus/test/framework/envtest/cluster"
)

// Environment is similar to sigs.k8s.io/controller-runtime/pkg/envtest.Environment,
// but only keep the inject resource.
type Environment struct {
	// CRDInstallOptions are the options for installing CRDs.
	CRDInstallOptions CRDInstallOptions

	// CRDInstallOptions are the options for installing webhooks.
	WebhookInstallOptions WebhookInstallOptions

	config      *rest.Config
	provisioner cluster.Provisioner
}

// Start starts a local Kubernetes cluster if needed.
func (te *Environment) Start() (*rest.Config, error) {
	if !IsUsingExistingCluster() {
		var err error
		te.provisioner, err = GetProvisioner()
		if err != nil {
			return nil, err
		}

		if err := te.provisioner.Startup(); err != nil {
			return nil, err
		}
	} else {
		log.V(1).Info("using existing cluster")
	}

	var clientConfig, err = GetConfig()
	if err != nil {
		return nil, err
	}
	te.config, err = clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	log.V(1).Info("installing CRDs")
	err = te.CRDInstallOptions.Install(te.config)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("installing webhooks")
	err = te.WebhookInstallOptions.Install(te.config)
	if err != nil {
		return nil, err
	}

	return te.config, nil
}

// Stop stops the local Kubernetes cluster.
// Previously installed CRDs, as listed in CRDInstallOptions.CRDs, will be uninstalled
// if CRDInstallOptions.CleanUpAfterUse are set to true.
func (te *Environment) Stop() error {
	log.V(1).Info("cleaning up webhooks")
	var err = te.WebhookInstallOptions.Cleanup()
	if err != nil {
		return err
	}

	log.V(1).Info("cleaning up CRDs")
	err = te.CRDInstallOptions.Cleanup(te.config)
	if err != nil {
		return err
	}

	if !IsUsingExistingCluster() && te.provisioner != nil {
		return te.provisioner.Cleanup()
	}
	return nil
}

func (te *Environment) AddWorker(name string) error {
	if !IsUsingExistingCluster() && te.provisioner != nil {
		return te.provisioner.AddWorker(name)
	}
	return nil
}
