package types

import "strconv"

type RelayWeightsFloat64 struct {
	Fee           float64
	Uptime        float64
	SuccessRate   float64
	ExecutionTime float64
	FeatureSet    float64
}

// Float64Values returns the float64Values of our RelayWeight strings.  On error, we use 0
func (m *RelayWeights) Float64Values() RelayWeightsFloat64 {
	fee, _ := strconv.ParseFloat(m.Fee, 64)
	uptime, _ := strconv.ParseFloat(m.Uptime, 64)
	successRate, _ := strconv.ParseFloat(m.SuccessRate, 64)
	executionTime, _ := strconv.ParseFloat(m.ExecutionTime, 64)
	featureSet, _ := strconv.ParseFloat(m.ExecutionTime, 64)
	return RelayWeightsFloat64{
		Fee:           fee,
		Uptime:        uptime,
		SuccessRate:   successRate,
		ExecutionTime: executionTime,
		FeatureSet:    featureSet,
	}
}
