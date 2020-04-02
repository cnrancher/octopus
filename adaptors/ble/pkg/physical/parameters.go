package physical

import (
	"time"
)

const (
	defaultSyncInterval = 30
	defaultTimeout      = 60
)

type Parameters struct {
	SyncInterval time.Duration `json:"syncInterval,omitempty"`
	Timeout      time.Duration `json:"timeout,omitempty"`
}

func (p *Parameters) Validate() error {
	// nothing to do

	return nil
}

func DefaultParameters() Parameters {
	return Parameters{
		SyncInterval: defaultSyncInterval,
		Timeout:      defaultTimeout,
	}
}
