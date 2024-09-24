package keeper

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	"github.com/palomachain/paloma/x/evm/types"
	metrixtypes "github.com/palomachain/paloma/x/metrix/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBuildValidatorsInfos(t *testing.T) {
	k, ms, ctx := NewEvmKeeper(t)
	mAssigner := newMsgAssigner(
		ms.ValsetKeeper,
		ms.MetrixKeeper,
		ms.TreasuryKeeper,
		k.Logger,
	)

	testcases := []struct {
		name     string
		input    []valsettypes.Validator
		expected map[string]ValidatorInfo
		setup    func()
	}{
		{
			name: "returns base weights for all validators",
			input: []valsettypes.Validator{
				{
					Address: sdk.ValAddress("testvalidator1"),
				},
				{
					Address: sdk.ValAddress("testvalidator2"),
				},
				{
					Address: sdk.ValAddress("testvalidator3"),
				},
			},
			expected: map[string]ValidatorInfo{
				sdk.ValAddress("testvalidator1").String(): {
					ExecutionTime: math.LegacyMustNewDecFromStr("5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
					Uptime:        math.LegacyMustNewDecFromStr("1.0"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				sdk.ValAddress("testvalidator2").String(): {
					ExecutionTime: math.LegacyMustNewDecFromStr("5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
					Uptime:        math.LegacyMustNewDecFromStr("1.0"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				sdk.ValAddress("testvalidator3").String(): {
					ExecutionTime: math.LegacyMustNewDecFromStr("5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
					Uptime:        math.LegacyMustNewDecFromStr("1.0"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
			},
			setup: func() {
				ms.MetrixKeeper.On("Validators", mock.Anything, mock.Anything).
					Return(&metrixtypes.QueryValidatorsResponse{
						ValMetrics: []metrixtypes.ValidatorMetrics{
							{
								ValAddress:    sdk.ValAddress("testvalidator1").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
							{
								ValAddress:    sdk.ValAddress("testvalidator2").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
							{
								ValAddress:    sdk.ValAddress("testvalidator3").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
						},
					}, nil)

				ms.TreasuryKeeper.On("GetRelayerFeesByChainReferenceID",
					mock.Anything, mock.Anything).
					Return(map[string]math.LegacyDec{
						sdk.ValAddress("testvalidator1").String(): math.LegacyMustNewDecFromStr("0.05"),
						sdk.ValAddress("testvalidator2").String(): math.LegacyMustNewDecFromStr("0.05"),
						sdk.ValAddress("testvalidator3").String(): math.LegacyMustNewDecFromStr("0.05"),
					}, nil)
			},
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			actual, err := mAssigner.buildValidatorsInfos(ctx, "test-main", tt.input)
			asserter.NoError(err)

			asserter.Equal(tt.expected, actual)
		})
	}
}

func TestScoreValue(t *testing.T) {
	testcases := []struct {
		name     string
		min      math.LegacyDec
		max      math.LegacyDec
		value    math.LegacyDec
		reverse  bool
		expected math.LegacyDec
	}{
		{
			name:     "scores the percentile of a value compared to the mean.  lower numbers are better",
			min:      math.LegacyMustNewDecFromStr("0.1"),
			max:      math.LegacyMustNewDecFromStr("0.6"),
			value:    math.LegacyMustNewDecFromStr("0.2"),
			reverse:  true,
			expected: math.LegacyMustNewDecFromStr("0.2"),
		},
		{
			name:     "scores the percentile of a value compared to the mean.  higher numbers are better",
			min:      math.LegacyMustNewDecFromStr("0.1"),
			max:      math.LegacyMustNewDecFromStr("0.6"),
			value:    math.LegacyMustNewDecFromStr("0.2"),
			reverse:  false,
			expected: math.LegacyMustNewDecFromStr("0.8"),
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual := scoreValue(tt.min, tt.max, tt.value, tt.reverse)

			asserter.Equal(tt.expected, actual)
		})
	}
}

func TestRankValidators(t *testing.T) {
	_, _, ctx := NewEvmKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	testcases := []struct {
		name            string
		validatorsInfos map[string]ValidatorInfo
		relayWeights    types.RelayWeightDec
		expected        scoreSnapshot
		expectedErr     error
	}{
		{
			name: "When all are equal, all are ranked the same",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator2": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator3": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
			},
			relayWeights: types.RelayWeightDec{
				ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
				SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
				Uptime:        math.LegacyMustNewDecFromStr("0.5"),
				Fee:           math.LegacyMustNewDecFromStr("0.5"),
				FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
			},
			expected: scoreSnapshot{
				blockHeight: sdkCtx.BlockHeight(),
				scores: []validatorScore{
					{address: "testvalidator1", score: math.LegacyNewDec(0)},
					{address: "testvalidator2", score: math.LegacyNewDec(0)},
					{address: "testvalidator3", score: math.LegacyNewDec(0)},
				},
			},
		},
		{
			name: "Validator with a lower fee is ranked higher",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.04"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator2": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator3": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
			},
			relayWeights: types.RelayWeightDec{
				ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
				SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
				Uptime:        math.LegacyMustNewDecFromStr("0.5"),
				Fee:           math.LegacyMustNewDecFromStr("0.5"),
				FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
			},
			expected: scoreSnapshot{
				blockHeight: sdkCtx.BlockHeight(),
				scores: []validatorScore{
					{address: "testvalidator1", score: math.LegacyMustNewDecFromStr("0.5")},
					{address: "testvalidator2", score: math.LegacyNewDec(0)},
					{address: "testvalidator3", score: math.LegacyNewDec(0)},
				},
			},
		},
		{
			name: "Validator with a lower executionTime is ranked higher",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator2": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.4"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator3": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
			},
			relayWeights: types.RelayWeightDec{
				ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
				SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
				Uptime:        math.LegacyMustNewDecFromStr("0.5"),
				Fee:           math.LegacyMustNewDecFromStr("0.5"),
				FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
			},
			expected: scoreSnapshot{
				blockHeight: sdkCtx.BlockHeight(),
				scores: []validatorScore{
					{address: "testvalidator2", score: math.LegacyMustNewDecFromStr("0.5")},
					{address: "testvalidator1", score: math.LegacyNewDec(0)},
					{address: "testvalidator3", score: math.LegacyNewDec(0)},
				},
			},
		},
		{
			name: "Validator with a higher uptime is ranked higher",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator2": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator3": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.9"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
			},
			relayWeights: types.RelayWeightDec{
				ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
				SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
				Uptime:        math.LegacyMustNewDecFromStr("0.5"),
				Fee:           math.LegacyMustNewDecFromStr("0.5"),
				FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
			},
			expected: scoreSnapshot{
				blockHeight: sdkCtx.BlockHeight(),
				scores: []validatorScore{
					{address: "testvalidator3", score: math.LegacyMustNewDecFromStr("0.5")},
					{address: "testvalidator1", score: math.LegacyNewDec(0)},
					{address: "testvalidator2", score: math.LegacyNewDec(0)},
				},
			},
		},
		{
			name: "Validator with a higher success rate is ranked higher",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator2": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.8"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator3": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
			},
			relayWeights: types.RelayWeightDec{
				ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
				SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
				Uptime:        math.LegacyMustNewDecFromStr("0.5"),
				Fee:           math.LegacyMustNewDecFromStr("0.5"),
				FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
			},
			expected: scoreSnapshot{
				blockHeight: sdkCtx.BlockHeight(),
				scores: []validatorScore{
					{address: "testvalidator2", score: math.LegacyMustNewDecFromStr("0.5")},
					{address: "testvalidator1", score: math.LegacyNewDec(0)},
					{address: "testvalidator3", score: math.LegacyNewDec(0)},
				},
			},
		},
		{
			name: "Validator with a higher feature set count is ranked higher",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator2": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.8"),
				},
				"testvalidator3": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
			},
			relayWeights: types.RelayWeightDec{
				ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
				SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
				Uptime:        math.LegacyMustNewDecFromStr("0.5"),
				Fee:           math.LegacyMustNewDecFromStr("0.5"),
				FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
			},
			expected: scoreSnapshot{
				blockHeight: sdkCtx.BlockHeight(),
				scores: []validatorScore{
					{address: "testvalidator2", score: math.LegacyMustNewDecFromStr("0.5")},
					{address: "testvalidator1", score: math.LegacyNewDec(0)},
					{address: "testvalidator3", score: math.LegacyNewDec(0)},
				},
			},
		},
		{
			name: "higher success rate offset by lower success weight.  lower fee adjusted with higher fee weight",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator2": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.8"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator3": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.04"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
			},
			relayWeights: types.RelayWeightDec{
				ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
				SuccessRate:   math.LegacyMustNewDecFromStr("0.25"),
				Uptime:        math.LegacyMustNewDecFromStr("0.5"),
				Fee:           math.LegacyMustNewDecFromStr("0.75"),
				FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
			},
			expected: scoreSnapshot{
				blockHeight: sdkCtx.BlockHeight(),
				scores: []validatorScore{
					{address: "testvalidator3", score: math.LegacyMustNewDecFromStr("0.75")},
					{address: "testvalidator2", score: math.LegacyMustNewDecFromStr("0.25")},
					{address: "testvalidator1", score: math.LegacyNewDec(0)},
				},
			},
		},
		{
			name: "Full mixed bag example",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.4"), // Faster executionTime
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.8"), // Higher uptime
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator2": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
					SuccessRate:   math.LegacyMustNewDecFromStr("0.8"), // Higher success rate
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.05"),
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
				"testvalidator3": {
					ExecutionTime: math.LegacyMustNewDecFromStr("0.6"), // Slower executionTime
					SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
					Uptime:        math.LegacyMustNewDecFromStr("0.7"),
					Fee:           math.LegacyMustNewDecFromStr("0.04"), // Lower fee
					FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
				},
			},
			relayWeights: types.RelayWeightDec{
				ExecutionTime: math.LegacyMustNewDecFromStr("0.5"),
				SuccessRate:   math.LegacyMustNewDecFromStr("0.5"),
				Uptime:        math.LegacyMustNewDecFromStr("0.5"),
				Fee:           math.LegacyMustNewDecFromStr("0.5"),
				FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
			},
			expected: scoreSnapshot{
				blockHeight: sdkCtx.BlockHeight(),
				scores: []validatorScore{
					{address: "testvalidator1", score: math.LegacyMustNewDecFromStr("1")},
					{address: "testvalidator2", score: math.LegacyMustNewDecFromStr("0.75")},
					{address: "testvalidator3", score: math.LegacyMustNewDecFromStr("0.5")},
				},
			},
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := rankValidators(ctx, tt.validatorsInfos, tt.relayWeights)
			asserter.NoError(err)

			asserter.Equal(tt.expected, actual)
		})
	}
}

func TestFilterValidatorsForJob(t *testing.T) {
	chainID := "test-chain"
	scores := scoreSnapshot{
		scores: []validatorScore{
			{
				address: sdk.ValAddress("validator-1").String(),
				score:   math.LegacyMustNewDecFromStr("0.5"),
			},
			{
				address: sdk.ValAddress("validator-2").String(),
				score:   math.LegacyMustNewDecFromStr("0.5"),
			},
			{
				address: sdk.ValAddress("validator-3").String(),
				score:   math.LegacyMustNewDecFromStr("0.5"),
			},
		},
	}
	defaultValidators := map[string]valsettypes.Validator{
		sdk.ValAddress("validator-1").String(): {
			Address: sdk.ValAddress("validator-1"),
			ExternalChainInfos: []*valsettypes.ExternalChainInfo{
				{
					ChainReferenceID: "test-chain",
				},
			},
		},
		sdk.ValAddress("validator-2").String(): {
			Address: sdk.ValAddress("validator-2"),
			ExternalChainInfos: []*valsettypes.ExternalChainInfo{
				{
					ChainReferenceID: "test-chain",
				},
			},
		},
		sdk.ValAddress("validator-3").String(): {
			Address: sdk.ValAddress("validator-3"),
			ExternalChainInfos: []*valsettypes.ExternalChainInfo{
				{
					ChainReferenceID: "test-chain",
				},
			},
		},
	}
	tests := []struct {
		name         string
		validators   map[string]valsettypes.Validator
		chainID      string
		requirements *xchain.JobRequirements
		expected     []validatorScore
		expectedStr  string
	}{
		{
			name:         "with empty validator slice",
			expectedStr:  "should return empty slice",
			validators:   map[string]valsettypes.Validator{},
			expected:     nil,
			requirements: nil,
			chainID:      chainID,
		},
		{
			name:         "with nil requirement",
			expectedStr:  "should return input slice",
			validators:   defaultValidators,
			expected:     scores.scores,
			requirements: nil,
			chainID:      chainID,
		},
		{
			name:         "with no set requirement",
			expectedStr:  "should return input slice",
			validators:   defaultValidators,
			expected:     scores.scores,
			requirements: &xchain.JobRequirements{},
			chainID:      chainID,
		},
		{
			name:        "with MEV set as requirement",
			expectedStr: "should return only validators with MEV trait",
			validators: map[string]valsettypes.Validator{
				sdk.ValAddress("validator-1").String(): {
					Address: sdk.ValAddress("validator-1"),
				},
				sdk.ValAddress("validator-2").String(): {
					Address: sdk.ValAddress("validator-2"),
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: chainID,
							Traits:           []string{valsettypes.PIGEON_TRAIT_MEV},
						},
					},
				},
				sdk.ValAddress("validator-3").String(): {
					Address: sdk.ValAddress("validator-3"),
				},
			},
			expected:     []validatorScore{scores.scores[1]},
			requirements: &xchain.JobRequirements{EnforceMEVRelay: true},
			chainID:      chainID,
		},
		{
			name:        "with MEV set as requirement and multiple chains configured - REGRESSION",
			expectedStr: "should return only validators with MEV trait",
			validators: map[string]valsettypes.Validator{
				sdk.ValAddress("validator-1").String(): {
					Address: sdk.ValAddress("validator-1"),
				},
				sdk.ValAddress("validator-2").String(): {
					Address: sdk.ValAddress("validator-2"),
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: "other-chain",
							Traits:           []string{valsettypes.PIGEON_TRAIT_MEV},
						},
						{
							ChainReferenceID: chainID,
							Traits:           []string{valsettypes.PIGEON_TRAIT_MEV},
						},
					},
				},
				sdk.ValAddress("validator-3").String(): {
					Address: sdk.ValAddress("validator-3"),
				},
			},
			expected:     []validatorScore{scores.scores[1]},
			requirements: &xchain.JobRequirements{EnforceMEVRelay: true},
			chainID:      chainID,
		},
		{
			name:        "with validators not supporting required chain ID",
			expectedStr: "should remove validators without full chain support from result set",
			validators: map[string]valsettypes.Validator{
				sdk.ValAddress("validator-1").String(): {
					Address: sdk.ValAddress("validator-1"),
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: "other-chain",
						},
					},
				},
				sdk.ValAddress("validator-2").String(): {
					Address: sdk.ValAddress("validator-2"),
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: "other-chain",
						},
						{
							ChainReferenceID: chainID,
						},
					},
				},
				sdk.ValAddress("validator-3").String(): {
					Address: sdk.ValAddress("validator-3"),
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: "other-chain",
						},
					},
				},
			},
			expected:     []validatorScore{scores.scores[1]},
			requirements: nil,
			chainID:      chainID,
		},
	}

	for k, v := range tests {
		t.Run(fmt.Sprintf("%d. %s", k, v.name), func(t *testing.T) {
			r := filterValidatorsForJob(scores, v.validators, v.chainID, v.requirements)
			require.Equal(t, v.expected, r, v.expectedStr)
		})
	}
}

func TestPickValidatorForMessage(t *testing.T) {
	testcases := []struct {
		name         string
		setup        func() (msgAssigner, sdk.Context)
		weights      *types.RelayWeights
		chainID      string
		requirements *xchain.JobRequirements
		expected     string
		expectedErr  error
	}{
		{
			name: "assigns a consistent validator in the happy path with equal weights",
			setup: func() (msgAssigner, sdk.Context) {
				k, ms, ctx := NewEvmKeeper(t)
				mAssigner := newMsgAssigner(
					ms.ValsetKeeper,
					ms.MetrixKeeper,
					ms.TreasuryKeeper,
					k.Logger,
				)

				snapshot := &valsettypes.Snapshot{
					Id:          1,
					Chains:      []string{"test-chain"},
					TotalShares: sdkmath.NewInt(75000),
					Validators: []valsettypes.Validator{
						{
							Address: sdk.ValAddress("testvalidator1"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							Address: sdk.ValAddress("testvalidator2"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							Address: sdk.ValAddress("testvalidator3"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainReferenceID: "test-chain",
								},
							},
						},
					},
				}
				ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(snapshot, nil)
				ms.MetrixKeeper.On("Validators", mock.Anything, mock.Anything).
					Return(&metrixtypes.QueryValidatorsResponse{
						ValMetrics: []metrixtypes.ValidatorMetrics{
							{
								ValAddress:    sdk.ValAddress("testvalidator1").String(),
								ExecutionTime: math.NewInt(1),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
							{
								ValAddress:    sdk.ValAddress("testvalidator2").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
							{
								ValAddress:    sdk.ValAddress("testvalidator3").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
						},
					}, nil)
				ms.TreasuryKeeper.On("GetRelayerFeesByChainReferenceID",
					mock.Anything, mock.Anything).
					Return(map[string]math.LegacyDec{
						sdk.ValAddress("testvalidator1").String(): math.LegacyMustNewDecFromStr("0.05"),
						sdk.ValAddress("testvalidator2").String(): math.LegacyMustNewDecFromStr("0.05"),
						sdk.ValAddress("testvalidator3").String(): math.LegacyMustNewDecFromStr("0.05"),
					}, nil)

				return mAssigner, ctx
			},
			weights: &types.RelayWeights{
				FeatureSet:    "1",
				ExecutionTime: "5",
				SuccessRate:   "0.5",
				Uptime:        "0.5",
				Fee:           "0.5",
			},
			chainID:  "test-chain",
			expected: sdk.ValAddress("testvalidator1").String(),
		},
		{
			name: "returns error when error getting snapshot",
			setup: func() (msgAssigner, sdk.Context) {
				k, ms, ctx := NewEvmKeeper(t)
				mAssigner := newMsgAssigner(
					ms.ValsetKeeper,
					ms.MetrixKeeper,
					ms.TreasuryKeeper,
					k.Logger,
				)

				ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(nil, errors.New("example-error"))

				return mAssigner, ctx
			},
			expectedErr: errors.New("example-error"),
		},
		{
			name: "returns error when no snapshot found",
			setup: func() (msgAssigner, sdk.Context) {
				k, ms, ctx := NewEvmKeeper(t)
				mAssigner := newMsgAssigner(
					ms.ValsetKeeper,
					ms.MetrixKeeper,
					ms.TreasuryKeeper,
					k.Logger,
				)

				ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(nil, nil)

				return mAssigner, ctx
			},
			expectedErr: errors.New("no snapshot found"),
		},
		{
			chainID:  "test-chain",
			expected: sdk.ValAddress("testvalidator3").String(),
			name:     "assigns the first validator in lexicographical order when scores are equal",
			setup: func() (msgAssigner, sdk.Context) {
				k, ms, ctx := NewEvmKeeper(t)
				mAssigner := newMsgAssigner(
					ms.ValsetKeeper,
					ms.MetrixKeeper,
					ms.TreasuryKeeper,
					k.Logger,
				)

				snapshot := &valsettypes.Snapshot{
					Id:          1,
					Chains:      []string{"test-chain"},
					TotalShares: sdkmath.NewInt(75000),
					Validators: []valsettypes.Validator{
						{
							Address: sdk.ValAddress("testvalidator1"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							Address: sdk.ValAddress("testvalidator2"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							Address: sdk.ValAddress("testvalidator3"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainReferenceID: "test-chain",
								},
							},
						},
					},
				}
				ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(snapshot, nil)
				ms.MetrixKeeper.On("Validators", mock.Anything, mock.Anything).
					Return(&metrixtypes.QueryValidatorsResponse{
						ValMetrics: []metrixtypes.ValidatorMetrics{
							{
								ValAddress:    sdk.ValAddress("testvalidator1").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
							{
								ValAddress:    sdk.ValAddress("testvalidator2").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
							{
								ValAddress:    sdk.ValAddress("testvalidator3").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
						},
					}, nil)
				ms.TreasuryKeeper.On("GetRelayerFeesByChainReferenceID",
					mock.Anything, mock.Anything).
					Return(map[string]math.LegacyDec{
						sdk.ValAddress("testvalidator1").String(): math.LegacyMustNewDecFromStr("0.05"),
						sdk.ValAddress("testvalidator2").String(): math.LegacyMustNewDecFromStr("0.05"),
						sdk.ValAddress("testvalidator3").String(): math.LegacyMustNewDecFromStr("0.05"),
					}, nil)

				return mAssigner, ctx
			},
		},
		{
			chainID:  "test-chain",
			expected: sdk.ValAddress("testvalidator1").String(),
			name:     "assigns the second validator in lexicographical order when scores are equal and block timestamp end in 1",
			setup: func() (msgAssigner, sdk.Context) {
				k, ms, ctx := NewEvmKeeper(t)
				mAssigner := newMsgAssigner(
					ms.ValsetKeeper,
					ms.MetrixKeeper,
					ms.TreasuryKeeper,
					k.Logger,
				)

				ctx = ctx.WithBlockTime(time.Now().Round(time.Minute).Add(time.Second))

				snapshot := &valsettypes.Snapshot{
					Id:          1,
					Chains:      []string{"test-chain"},
					TotalShares: sdkmath.NewInt(75000),
					Validators: []valsettypes.Validator{
						{
							Address: sdk.ValAddress("testvalidator1"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							Address: sdk.ValAddress("testvalidator2"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							Address: sdk.ValAddress("testvalidator3"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainReferenceID: "test-chain",
								},
							},
						},
					},
				}
				ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(snapshot, nil)
				ms.MetrixKeeper.On("Validators", mock.Anything, mock.Anything).
					Return(&metrixtypes.QueryValidatorsResponse{
						ValMetrics: []metrixtypes.ValidatorMetrics{
							{
								ValAddress:    sdk.ValAddress("testvalidator1").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
							{
								ValAddress:    sdk.ValAddress("testvalidator2").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
							{
								ValAddress:    sdk.ValAddress("testvalidator3").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
						},
					}, nil)
				ms.TreasuryKeeper.On("GetRelayerFeesByChainReferenceID",
					mock.Anything, mock.Anything).
					Return(map[string]math.LegacyDec{
						sdk.ValAddress("testvalidator1").String(): math.LegacyMustNewDecFromStr("0.05"),
						sdk.ValAddress("testvalidator2").String(): math.LegacyMustNewDecFromStr("0.05"),
						sdk.ValAddress("testvalidator3").String(): math.LegacyMustNewDecFromStr("0.05"),
					}, nil)

				return mAssigner, ctx
			},
		},
		{
			name: "applies validator filtering based on job requirements",
			setup: func() (msgAssigner, sdk.Context) {
				k, ms, ctx := NewEvmKeeper(t)
				mAssigner := newMsgAssigner(
					ms.ValsetKeeper,
					ms.MetrixKeeper,
					ms.TreasuryKeeper,
					k.Logger,
				)

				snapshot := &valsettypes.Snapshot{
					Id:          1,
					Chains:      []string{"test-chain"},
					TotalShares: sdkmath.NewInt(75000),
					Validators: []valsettypes.Validator{
						{
							Address: sdk.ValAddress("testvalidator1"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							Address: sdk.ValAddress("testvalidator2"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{ChainReferenceID: "test-chain", Traits: []string{valsettypes.PIGEON_TRAIT_MEV}},
							},
						},
						{
							Address: sdk.ValAddress("testvalidator3"),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainReferenceID: "test-chain",
								},
							},
						},
					},
				}
				ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(snapshot, nil)
				ms.MetrixKeeper.On("Validators", mock.Anything, mock.Anything).
					Return(&metrixtypes.QueryValidatorsResponse{
						ValMetrics: []metrixtypes.ValidatorMetrics{
							{
								ValAddress:    sdk.ValAddress("testvalidator1").String(),
								ExecutionTime: math.NewInt(1), // Lowest
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
							{
								ValAddress:    sdk.ValAddress("testvalidator2").String(),
								ExecutionTime: math.NewInt(5),
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
							{
								ValAddress:    sdk.ValAddress("testvalidator3").String(),
								ExecutionTime: math.NewInt(1), // Lowest
								SuccessRate:   math.LegacyMustNewDecFromStr("1.0"),
								Uptime:        math.LegacyMustNewDecFromStr("1.0"),
								FeatureSet:    math.LegacyMustNewDecFromStr("0.5"),
							},
						},
					}, nil)
				ms.TreasuryKeeper.On("GetRelayerFeesByChainReferenceID",
					mock.Anything, mock.Anything).
					Return(map[string]math.LegacyDec{
						sdk.ValAddress("testvalidator1").String(): math.LegacyMustNewDecFromStr("0.05"),
						sdk.ValAddress("testvalidator2").String(): math.LegacyMustNewDecFromStr("0.05"),
						sdk.ValAddress("testvalidator3").String(): math.LegacyMustNewDecFromStr("0.05"),
					}, nil)

				return mAssigner, ctx
			},
			expected:     sdk.ValAddress("testvalidator2").String(),
			chainID:      "test-chain",
			requirements: &xchain.JobRequirements{EnforceMEVRelay: true},
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			messageAssigner, ctx := tt.setup()

			actual, actualErr := messageAssigner.PickValidatorForMessage(ctx,
				tt.weights, tt.chainID, tt.requirements)

			asserter.Equal(tt.expected, actual, tt.name)
			asserter.Equal(tt.expectedErr, actualErr, tt.name)
		})
	}
}
