package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
)

type RelayWeightDec struct {
	Fee           math.LegacyDec
	Uptime        math.LegacyDec
	SuccessRate   math.LegacyDec
	ExecutionTime math.LegacyDec
	FeatureSet    math.LegacyDec
}

// DecValues returns the math.LegacyDec of our RelayWeight strings.  On error, we use 0
func (m *RelayWeights) DecValues() (w RelayWeightDec, err error) {
	w.Fee, err = math.LegacyNewDecFromStr(m.Fee)
	if err != nil {
		return w, fmt.Errorf("error converting Fee '%s' to Dec: %w", m.Fee, err)
	}
	w.Uptime, err = math.LegacyNewDecFromStr(m.Uptime)
	if err != nil {
		return w, fmt.Errorf("error converting Uptime '%s' to Dec: %w", m.Uptime, err)
	}
	w.SuccessRate, err = math.LegacyNewDecFromStr(m.SuccessRate)
	if err != nil {
		return w, fmt.Errorf("error converting SuccessRate '%s' to Dec: %w", m.SuccessRate, err)
	}
	w.ExecutionTime, err = math.LegacyNewDecFromStr(m.ExecutionTime)
	if err != nil {
		return w, fmt.Errorf("error converting ExecutionTime '%s' to Dec: %w", m.ExecutionTime, err)
	}
	w.FeatureSet, err = math.LegacyNewDecFromStr(m.FeatureSet)
	if err != nil {
		return w, fmt.Errorf("error converting FeatureSet '%s' to Dec: %w", m.FeatureSet, err)
	}

	return w, nil
}

func (m *RelayWeights) ValueOrDefault() *RelayWeights {
	if m == nil {
		return &RelayWeights{
			Fee:           "1.0",
			Uptime:        "1.0",
			SuccessRate:   "1.0",
			ExecutionTime: "1.0",
			FeatureSet:    "1.0",
		}
	}
	return m
}
