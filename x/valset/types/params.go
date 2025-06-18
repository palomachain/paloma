package types

import (
	fmt "fmt"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var (
	KeyMissedAttestationJailDuration                     = []byte("MissedAttestationJailDuration")
	MissedBridgeClaimJailDuration                        = []byte("MissedBridgeClaimJailDuration")
	_                                paramtypes.ParamSet = (*Params)(nil)
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		MissedAttestationJailDuration: time.Hour * 12,
		MissedBridgeClaimJailDuration: time.Hour * 24,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMissedAttestationJailDuration, &p.MissedAttestationJailDuration, validateMissedAttestationJailDuration),
		paramtypes.NewParamSetPair(MissedBridgeClaimJailDuration, &p.MissedBridgeClaimJailDuration, validateMissedBridgeClaimJailDuration),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	switch {
	case p.MissedAttestationJailDuration == 0:
		return fmt.Errorf("missed attestation jail duration must be greater than zero")
	case p.MissedBridgeClaimJailDuration == 0:
		return fmt.Errorf("missed bridge claim jail duration must be greater than zero")
	default:
		return nil
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateMissedAttestationJailDuration(v interface{}) error {
	d, ok := v.(time.Duration)
	if !ok {
		return fmt.Errorf("expected int64 type for MissedAttestationJailDuration, got %T", v)
	}
	if d <= 0 {
		return fmt.Errorf("MissedAttestationJailDuration must be greater than zero, got %d", d)
	}
	return nil
}

func validateMissedBridgeClaimJailDuration(v interface{}) error {
	d, ok := v.(time.Duration)
	if !ok {
		return fmt.Errorf("expected int64 type for MissedBridgeClaimJailDuration, got %T", v)
	}
	if d <= 0 {
		return fmt.Errorf("MissedBridgeClaimJailDuration must be greater than zero, got %d", d)
	}
	return nil
}
