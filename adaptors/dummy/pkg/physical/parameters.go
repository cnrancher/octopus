package physical

import (
	"github.com/pkg/errors"
)

type Parameters struct {
	IP string `json:"ip"`
}

func (p *Parameters) Validate() error {
	if p == nil {
		return errors.New("parameter instance is nil")
	}

	if p.IP == "" {
		return errors.New("ip is required")
	}

	return nil
}
