// +build test

package cluster

import (
	"fmt"
)

// ProvisionerType specifies the type of testing cluster provisioner.
type ProvisionerType string

// Provisioner specifies the provisioner of testing cluster.
type Provisioner interface {
	fmt.Stringer

	// Startup creates the testing cluster.
	Startup() error

	// Cleanup deletes the testing cluster.
	Cleanup() error

	// AddWorker adds the given worker to the testing cluster.
	AddWorker(name string) error

	// IsLocalCluster indicates the testing cluster is local or not.
	IsLocalCluster() bool
}
