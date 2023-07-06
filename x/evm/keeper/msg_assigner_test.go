package keeper

import (
	"errors"
	"math"
	"testing"

	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/palomachain/paloma/x/evm/types/mocks"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
				},
				sdk.ValAddress("testvalidator2").String(): {
					ExecutionTime: 0.5,
					SuccessRate:   1.0,
					Uptime:        1.0,
					Fee:           0.05,
				},
				sdk.ValAddress("testvalidator3").String(): {
					ExecutionTime: 0.5,
					SuccessRate:   1.0,
					Uptime:        1.0,
					Fee:           0.05,
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
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
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
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
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
				},
				"testvalidator2": {
					ExecutionTime: 0.4,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
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
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.9,
					Fee:           0.05,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
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
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.8,
					Uptime:        0.7,
					Fee:           0.05,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.05,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
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
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.8,
					Uptime:        0.7,
					Fee:           0.05,
				},
				"testvalidator3": {
					ExecutionTime: 0.5,
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.04,
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.25,
				Uptime:        0.5,
				Fee:           0.75,
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
				},
				"testvalidator2": {
					ExecutionTime: 0.5,
					SuccessRate:   0.8, // Higher success rate
					Uptime:        0.7,
					Fee:           0.05,
				},
				"testvalidator3": {
					ExecutionTime: 0.6, // Slower executionTime
					SuccessRate:   0.5,
					Uptime:        0.7,
					Fee:           0.04, // Lower fee
				},
			},
			relayWeights: types.RelayWeightsFloat64{
				ExecutionTime: 0.5,
				SuccessRate:   0.5,
				Uptime:        0.5,
				Fee:           0.5,
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
		name        string
		setup       func() MsgAssigner
		weights     *types.RelayWeights
		expected    string
		expectedErr error
	}{
		{
			name: "assigns a consistent pseudo-random validator in the happy path with equal weights",
			setup: func() MsgAssigner {
				msgAssigner := MsgAssigner{}

				valsetKeeperMock := mocks.NewValsetKeeper(t)

				snapshot := &valsettypes.Snapshot{
					Id:          1,
					Chains:      []string{"test-chain"},
					TotalShares: sdk.NewInt(75000),
					Validators: []valsettypes.Validator{
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
			name: "assigns a consistent pseudo-random validator in the happy path when no weights exist (cold start)",
			setup: func() MsgAssigner {
				msgAssigner := MsgAssigner{}

				valsetKeeperMock := mocks.NewValsetKeeper(t)

				snapshot := &valsettypes.Snapshot{
					Id:          1,
					Chains:      []string{"test-chain"},
					TotalShares: sdk.NewInt(75000),
					Validators: []valsettypes.Validator{
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
				}
				valsetKeeperMock.On("GetCurrentSnapshot", mock.Anything).Return(snapshot, nil)

				msgAssigner.ValsetKeeper = valsetKeeperMock

				return msgAssigner
			},
			expected: sdk.ValAddress("testvalidator1").String(),
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

			actual, actualErr := messageAssigner.PickValidatorForMessage(ctx, tt.weights)

			asserter.Equal(tt.expected, actual)
			asserter.Equal(tt.expectedErr, actualErr)
		})
	}
}
