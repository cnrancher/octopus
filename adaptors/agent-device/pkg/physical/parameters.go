package physical

type Parameters struct {
}

func (p *Parameters) Validate() error {
	// nothing to do

	return nil
}

func DefaultParameters() Parameters {
	return Parameters{}
}
