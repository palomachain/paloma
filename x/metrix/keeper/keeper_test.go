package keeper

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/palomachain/paloma/x/metrix/types"
	"github.com/stretchr/testify/require"
)

func TestCalculateUptime(t *testing.T) {
	tests := []struct {
		window int64
		missed int64
		want   math.LegacyDec
	}{
		{100, 10, math.LegacyMustNewDecFromStr("0.9")},
		{200, 50, math.LegacyMustNewDecFromStr("0.75")},
		{300, 0, math.LegacyMustNewDecFromStr("1.0")},
		{0, 0, math.LegacyNewDec(0)},
	}

	for _, test := range tests {
		got := calculateUptime(test.window, test.missed)
		if !got.Equal(test.want) {
			t.Errorf("calculateUptime(%d, %d) = %f; want %f", test.window, test.missed, got, test.want)
		}
	}
}

func TestApply(t *testing.T) {
	for i, v := range []struct {
		name   string
		patch  recordPatch
		assert func(recordPatch, *types.ValidatorMetrics) bool
	}{
		{
			name: "with uptime",
			patch: recordPatch{
				uptime: float64Ptr(0.95),
			},
			assert: func(p recordPatch, v *types.ValidatorMetrics) bool {
				return v.Uptime.Equal(*p.uptime)
			},
		},
		{
			name: "with success rate",
			patch: recordPatch{
				successRate: float64Ptr(0.25),
			},
			assert: func(p recordPatch, v *types.ValidatorMetrics) bool {
				return v.SuccessRate.Equal(*p.successRate)
			},
		},
		{
			name: "with execution time",
			patch: recordPatch{
				executionTime: intPtr(10),
			},
			assert: func(p recordPatch, v *types.ValidatorMetrics) bool {
				return v.ExecutionTime.Equal(*p.executionTime)
			},
		},
		{
			name: "with fee",
			patch: recordPatch{
				fee: intPtr(250),
			},
			assert: func(p recordPatch, v *types.ValidatorMetrics) bool {
				return v.Fee.Equal(*p.fee)
			},
		},
		{
			name: "with feature set",
			patch: recordPatch{
				featureSet: float64Ptr(1.0),
			},
			assert: func(p recordPatch, v *types.ValidatorMetrics) bool {
				return v.FeatureSet.Equal(*p.featureSet)
			},
		},
	} {
		t.Run(fmt.Sprintf("%d. %s", i, v.name), func(t *testing.T) {
			r := require.New(t)
			record := types.ValidatorMetrics{
				ValAddress:    "validator-1",
				Uptime:        math.LegacyZeroDec(),
				SuccessRate:   math.LegacyZeroDec(),
				ExecutionTime: math.ZeroInt(),
				Fee:           math.ZeroInt(),
				FeatureSet:    math.LegacyZeroDec(),
			}

			patched := v.patch.apply(record)
			r.Equal(record.ValAddress, patched.ValAddress)
			r.True(v.assert(v.patch, patched))
		})
	}

	t.Run("with multiple changed fields", func(t *testing.T) {
		r := require.New(t)
		record := types.ValidatorMetrics{
			ValAddress:    "validator-1",
			Uptime:        *float64Ptr(0.8),
			SuccessRate:   *float64Ptr(0.75),
			ExecutionTime: *intPtr(50),
			Fee:           *intPtr(250),
			FeatureSet:    *float64Ptr(1.0),
		}

		patch := recordPatch{
			uptime:     float64Ptr(0.9),
			fee:        intPtr(10),
			featureSet: float64Ptr(0),
		}

		patched := patch.apply(record)
		r.Equal(record.ValAddress, patched.ValAddress)
		r.Equal(record.SuccessRate, patched.SuccessRate)
		r.Equal(record.ExecutionTime, patched.ExecutionTime)
		r.Equal(*patch.uptime, patched.Uptime)
		r.Equal(*patch.fee, patched.Fee)
		r.Equal(*patch.featureSet, patched.FeatureSet)
	})
}

// Helper functions to create pointers to values
func float64Ptr(value float64) *math.LegacyDec {
	v, err := math.LegacyNewDecFromStr(fmt.Sprintf("%f", value))
	if err != nil {
		panic(err)
	}

	return &v
}

func intPtr(value int) *math.Int {
	v := math.NewIntFromUint64(uint64(value))
	return &v
}
