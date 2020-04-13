package physical

import (
	"time"
)

const defaultSyncInterval = 5

type Parameters struct {
	SyncInterval time.Duration `json:"syncInterval,omitempty"`
}

func (p *Parameters) Validate() error {
	return nil
}

func DefaultParameters() Parameters {
	return Parameters{
		SyncInterval: defaultSyncInterval,
	}
}
