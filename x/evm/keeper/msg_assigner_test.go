package keeper

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/palomachain/paloma/x/evm/types/mocks"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBuildValidatorsInfos(t *testing.T) {
	testcases := []struct {
		name     string
		input    []valsettypes.Validator
		expected map[string]ValidatorInfo
	}{
		{
			name: "returns our example base weights for all validators.  Will change once it's smarter",
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
					ExecutionTime: 0.5,
					SuccessRate:   1.0,
					Uptime:        1.0,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				sdk.ValAddress("testvalidator2").String(): {
					ExecutionTime: 0.5,
					SuccessRate:   1.0,
					Uptime:        1.0,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				sdk.ValAddress("testvalidator3").String(): {
					ExecutionTime: 0.5,
					SuccessRate:   1.0,
					Uptime:        1.0,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
			},
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildValidatorsInfos(tt.input)

			asserter.Equal(tt.expected, actual)
		})
	}
}

func TestScoreValue(t *testing.T) {
	testcases := []struct {
		name     string
		min      float64
		max      float64
		value    float64
		reverse  bool
		expected float64
	}{
		{
			name:     "scores the percentile of a value compared to the mean.  lower numbers are better",
			min:      0.1,
			max:      0.6,
			value:    0.2,
			reverse:  true,
			expected: 0.2,
		},
		{
			name:     "scores the percentile of a value compared to the mean.  higher numbers are better",
			min:      0.1,
			max:      0.6,
			value:    0.2,
			reverse:  false,
			expected: 0.8,
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual := scoreValue(tt.min, tt.max, tt.value, tt.reverse)

			asserter.Equal(tt.expected, math.Round(actual*100)/100)
		})
	}
}

func TestRankValidators(t *testing.T) {
	testcases := []struct {
		name            string
		validatorsInfos map[string]ValidatorInfo
		relayWeights    types.RelayWeightsFloat64
		expected        map[string]int
		expectedErr     error
	}{
		{
			name: "When all are equal, all are ranked the same",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
				FeatureSet:    0.5,
			},
			expected: map[string]int{
				"testvalidator1": 0,
				"testvalidator2": 0,
				"testvalidator3": 0,
			},
		},
		{
			name: "Validator with a lower fee is ranked higher",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.04,
					FeatureSet:    0.5,
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
				FeatureSet:    0.5,
			},
			expected: map[string]int{
				"testvalidator1": 5,
				"testvalidator2": 0,
				"testvalidator3": 0,
			},
		},
		{
			name: "Validator with a lower executionTime is ranked higher",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator2": {
					ExecutionTime: 0.4,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
				FeatureSet:    0.5,
			},
			expected: map[string]int{
				"testvalidator1": 0,
				"testvalidator2": 5,
				"testvalidator3": 0,
			},
		},
		{
			name: "Validator with a higher uptime is ranked higher",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.9,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
				FeatureSet:    0.5,
			},
			expected: map[string]int{
				"testvalidator1": 0,
				"testvalidator2": 0,
				"testvalidator3": 5,
			},
		},
		{
			name: "Validator with a higher success rate is ranked higher",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.8,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
				FeatureSet:    0.5,
			},
			expected: map[string]int{
				"testvalidator1": 0,
				"testvalidator2": 5,
				"testvalidator3": 0,
			},
		},
		{
			name: "Validator with a higher feature set count is ranked higher",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.8,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
				FeatureSet:    0.5,
			},
			expected: map[string]int{
				"testvalidator1": 0,
				"testvalidator2": 5,
				"testvalidator3": 0,
			},
		},
		{
			name: "higher success rate offset by lower success weight.  lower fee adjusted with higher fee weight",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.8,
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.04,
					FeatureSet:    0.5,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.25,
				Uptime:        0.5,
				Fee:           0.75,
				FeatureSet:    0.5,
			},
			expected: map[string]int{
				"testvalidator1": 0,
				"testvalidator2": 3,
				"testvalidator3": 8,
			},
		},
		{
			name: "Full mixed bag example",
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: 0.4, // Faster executionTime
					SuccessRate:   0.5,
					Uptime:        0.8, // Higher uptime
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.8, // Higher success rate
					Uptime:        0.7,
					Fee:           0.05,
					FeatureSet:    0.5,
				},
				"testvalidator3": {
					ExecutionTime: 0.6, // Slower executionTime
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.04, // Lower fee
					FeatureSet:    0.5,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
				FeatureSet:    0.5,
			},
			expected: map[string]int{
				"testvalidator1": 10,
				"testvalidator2": 8,
				"testvalidator3": 5,
			},
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual := rankValidators(tt.validatorsInfos, tt.relayWeights)

			asserter.Equal(tt.expected, actual)
		})
	}
}

func TestPickValidator(t *testing.T) {
	testcases := []struct {
		name            string
		weights         types.RelayWeightsFloat64
		validatorsInfos map[string]ValidatorInfo
		expected        string
	}{
		{
			name: "returns a consistent pseudo-random validator in the happy path with equal weights",
			weights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
			},
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.5,
					Fee:           0.5,
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.5,
					Fee:           0.5,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.5,
					Fee:           0.5,
				},
			},
			expected: "testvalidator2",
		},
		{
			name: "returns the validator with the higher score",
			weights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
			},
			validatorsInfos: map[string]ValidatorInfo{
				"testvalidator1": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.5,
					Fee:           0.5,
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.5,
					Fee:           0.5,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.5,
					Fee:           0.4,
				},
			},
			expected: "testvalidator3",
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.NewContext(
				nil,
				tmproto.Header{
					Height: 4,
				},
				false,
				log.NewNopLogger(),
			)

			actual := pickValidator(ctx, tt.validatorsInfos, tt.weights)

			asserter.Equal(tt.expected, actual)
		})
	}
}

func TestPickValidatorForMessage(t *testing.T) {
	testcases := []struct {
		name         string
		setup        func() MsgAssigner
		weights      *types.RelayWeights
		chainID      string
		requirements *xchain.JobRequirements
		expected     string
		expectedErr  error
	}{
		{
			name: "assigns a consistent pseudo-random validator in the happy path with equal weights",
			setup: func() MsgAssigner {
				msgAssigner := MsgAssigner{}

				valsetKeeperMock := mocks.NewValsetKeeper(t)

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
				valsetKeeperMock.On("GetCurrentSnapshot", mock.Anything).Return(snapshot, nil)

				msgAssigner.ValsetKeeper = valsetKeeperMock

				return msgAssigner
			},
			weights: &types.RelayWeights{
				ExecutionTime: "0.5",
				SuccessRate:   "0.5",
				Uptime:        "0.5",
				Fee:           "0.5",
			},
			chainID:  "test-chain",
			expected: sdk.ValAddress("testvalidator1").String(),
		},
		{
			name: "returns error when error getting snapshot",
			setup: func() MsgAssigner {
				msgAssigner := MsgAssigner{}

				valsetKeeperMock := mocks.NewValsetKeeper(t)

				valsetKeeperMock.On("GetCurrentSnapshot", mock.Anything).Return(nil, errors.New("example-error"))

				msgAssigner.ValsetKeeper = valsetKeeperMock

				return msgAssigner
			},
			expectedErr: errors.New("example-error"),
		},
		{
			name: "returns error when no snapshot found",
			setup: func() MsgAssigner {
				msgAssigner := MsgAssigner{}

				valsetKeeperMock := mocks.NewValsetKeeper(t)

				valsetKeeperMock.On("GetCurrentSnapshot", mock.Anything).Return(nil, nil)

				msgAssigner.ValsetKeeper = valsetKeeperMock

				return msgAssigner
			},
			expectedErr: errors.New("no snapshot found"),
		},
		{
			chainID:  "test-chain",
			expected: sdk.ValAddress("testvalidator1").String(),
			name:     "assigns a consistent pseudo-random validator in the happy path when no weights exist (cold start)",
			setup: func() MsgAssigner {
				msgAssigner := MsgAssigner{}

				valsetKeeperMock := mocks.NewValsetKeeper(t)

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
				valsetKeeperMock.On("GetCurrentSnapshot", mock.Anything).Return(snapshot, nil)

				msgAssigner.ValsetKeeper = valsetKeeperMock

				return msgAssigner
			},
		},
		{
			name: "applies validator filtering based on job requirements",
			setup: func() MsgAssigner {
				msgAssigner := MsgAssigner{}

				valsetKeeperMock := mocks.NewValsetKeeper(t)

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
				valsetKeeperMock.On("GetCurrentSnapshot", mock.Anything).Return(snapshot, nil)

				msgAssigner.ValsetKeeper = valsetKeeperMock

				return msgAssigner
			},
			expected:     sdk.ValAddress("testvalidator2").String(),
			chainID:      "test-chain",
			requirements: &xchain.JobRequirements{EnforceMEVRelay: true},
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.NewContext(
				nil,
				tmproto.Header{
					Height: 4,
				},
				false,
				log.NewNopLogger(),
			)

			messageAssigner := tt.setup()

			actual, actualErr := messageAssigner.PickValidatorForMessage(ctx, tt.weights, tt.chainID, tt.requirements)

			asserter.Equal(tt.expected, actual, tt.name)
			asserter.Equal(tt.expectedErr, actualErr, tt.name)
		})
	}
}

func TestFilterAssignableValidators(t *testing.T) {
	chainID := "test-chain"
	defaultValidators := []valsettypes.Validator{
		{
			Address: sdk.ValAddress("validator-1"),
			ExternalChainInfos: []*valsettypes.ExternalChainInfo{
				{
					ChainReferenceID: "test-chain",
				},
			},
		},
		{
			Address: sdk.ValAddress("validator-2"),
			ExternalChainInfos: []*valsettypes.ExternalChainInfo{
				{
					ChainReferenceID: "test-chain",
				},
			},
		},
		{
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
		validators   []valsettypes.Validator
		chainID      string
		requirements *xchain.JobRequirements
		expected     []valsettypes.Validator
		expectedStr  string
	}{
		{
			name:         "with empty validator slice",
			expectedStr:  "should return empty slice",
			validators:   []valsettypes.Validator{},
			expected:     []valsettypes.Validator(nil),
			requirements: nil,
			chainID:      chainID,
		},
		{
			name:         "with nil requirement",
			expectedStr:  "should return input slice",
			validators:   defaultValidators,
			expected:     defaultValidators,
			requirements: nil,
			chainID:      chainID,
		},
		{
			name:         "with no set requirement",
			expectedStr:  "should return input slice",
			validators:   defaultValidators,
			expected:     defaultValidators,
			requirements: &xchain.JobRequirements{},
			chainID:      chainID,
		},
		{
			name:        "with MEV set as requirement",
			expectedStr: "should return only validators with MEV trait",
			validators: []valsettypes.Validator{
				{
					Address: sdk.ValAddress("validator-1"),
				},
				{
					Address: sdk.ValAddress("validator-2"),
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: chainID,
							Traits:           []string{valsettypes.PIGEON_TRAIT_MEV},
						},
					},
				},
				{
					Address: sdk.ValAddress("validator-3"),
				},
			},
			expected: []valsettypes.Validator{
				{
					Address: sdk.ValAddress("validator-2"),
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: chainID,
							Traits:           []string{valsettypes.PIGEON_TRAIT_MEV},
						},
					},
				},
			},
			requirements: &xchain.JobRequirements{EnforceMEVRelay: true},
			chainID:      chainID,
		},
		{
			name:        "with MEV set as requirement and multiple chains configured - REGRESSION",
			expectedStr: "should return only validators with MEV trait",
			validators: []valsettypes.Validator{
				{
					Address: sdk.ValAddress("validator-1"),
				},
				{
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
				{
					Address: sdk.ValAddress("validator-3"),
				},
			},
			expected: []valsettypes.Validator{
				{
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
			},
			requirements: &xchain.JobRequirements{EnforceMEVRelay: true},
			chainID:      chainID,
		},
		{
			name:        "with validators not supporting required chain ID",
			expectedStr: "should remove validators without full chain support from result set",
			validators: []valsettypes.Validator{
				{
					Address: sdk.ValAddress("validator-1"),
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: "other-chain",
						},
					},
				},
				{
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
				{
					Address: sdk.ValAddress("validator-3"),
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: "other-chain",
						},
					},
				},
			},
			expected: []valsettypes.Validator{
				{
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
			},
			requirements: nil,
			chainID:      chainID,
		},
	}

	for k, v := range tests {
		t.Run(fmt.Sprintf("%d. %s", k, v.name), func(t *testing.T) {
			r := filterAssignableValidators(v.validators, v.chainID, v.requirements)
			require.Equal(t, v.expected, r, v.expectedStr)
		})
	}
}
