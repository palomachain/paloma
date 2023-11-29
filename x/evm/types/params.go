package types

// ParamKeyTable the param key table for launch module
func ParamKeyTable() KeyTable {
	return KeyTable{}
}

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() ParamSetPairs {
	return ParamSetPairs{}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}
func DefaultParams() Params {
	return NewParams()
}
