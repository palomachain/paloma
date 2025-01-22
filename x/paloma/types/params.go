package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var KeyGasExcemptAddresses = []byte("GasExemptAddresses")

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyGasExcemptAddresses, &p.GasExemptAddresses, validateGasExemptAddresses),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	lkup := make(map[string]struct{})
	for _, v := range p.GasExemptAddresses {
		if _, ok := lkup[v]; ok {
			return fmt.Errorf("duplicate gas exempt address %s", v)
		}

		if _, err := sdk.AccAddressFromBech32(v); err != nil {
			return fmt.Errorf("invalid address: %w", err)
		}

		lkup[v] = struct{}{}
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateGasExemptAddresses(i interface{}) error {
	addr, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	lkup := make(map[string]struct{})
	for _, v := range addr {
		if _, ok := lkup[v]; ok {
			return fmt.Errorf("duplicate gas exempt address %s", v)
		}

		if _, err := sdk.AccAddressFromBech32(v); err != nil {
			return fmt.Errorf("invalid gas exempt address: %w", err)
		}

		lkup[v] = struct{}{}
	}

	return nil
}
