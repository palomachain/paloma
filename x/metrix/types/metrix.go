package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/palomath"
)

func (m *ValidatorMetrics) ValueOrDefault(valAddr sdk.ValAddress) *ValidatorMetrics {
	if m != nil {
		return m
	}

	return &ValidatorMetrics{
		ValAddress:    valAddr.String(),
		Uptime:        palomath.LegacyDecFromFloat64(1.0),
		SuccessRate:   palomath.LegacyDecFromFloat64(0.5),
		ExecutionTime: math.ZeroInt(),
		Fee:           math.ZeroInt(),
		FeatureSet:    math.LegacyZeroDec(),
	}
}
