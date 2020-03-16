package physical

import (
	"time"
)

const (
	defaultSyncInterval = 30
	defaultTimeout      = 60
)

type Parameters struct {
	syncInterval time.Duration
	timeout      time.Duration
}

func (p *Parameters) Validate() error {
	if p == nil || p.timeout == 0 {
		p.timeout = defaultTimeout
	}
	if p == nil || p.syncInterval == 0 {
		p.syncInterval = defaultSyncInterval
	}
	return nil
}
