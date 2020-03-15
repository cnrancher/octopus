package physical

import (
	"time"
)

const (
	defaultSyncInterval = 5
	defaultTimeout      = 30
)

type Parameters struct {
	SyncInterval time.Duration `json:"syncInterval,omitempty"`
	Timeout      time.Duration `json:"timeout,omitempty"`
}

func (p *Parameters) Validate() error {
	// set default values
	if p == nil || p.SyncInterval == 0 {
		p.SyncInterval = defaultSyncInterval
	}
	if p == nil || p.Timeout == 0 {
		p.Timeout = defaultTimeout
	}
	return nil
}
