package types

import (
	"bytes"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

func NewParams() Params {
	return Params{}
}

// DefaultParams returns a copy of the default params
func DefaultParams() *Params {
	return &Params{}
}

// ValidateBasic checks that the parameters have valid values.
func (p Params) ValidateBasic() error {
	return nil
}

// ParamKeyTable for auth module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}
