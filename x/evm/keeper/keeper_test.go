package keeper

import (
	"errors"
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/palomachain/paloma/x/evm/types/mocks"
	schedulertypes "github.com/palomachain/paloma/x/scheduler/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type validatorChainInfo struct {
	chainType        string
	chainReferenceID string
}

func getValidators(num int, chains []validatorChainInfo) []valsettypes.Validator {
	validators := make([]valsettypes.Validator, num)
	for i := 0; i < num; i++ {
		chainInfos := make([]*valsettypes.ExternalChainInfo, len(chains))
		for i, chain := range chains {
			chainInfos[i] = &valsettypes.ExternalChainInfo{
				ChainType:        chain.chainType,
				ChainReferenceID: chain.chainReferenceID,
			}
		}
		validators[i] = valsettypes.Validator{
			State:              valsettypes.ValidatorState_ACTIVE,
			ShareCount:         sdkmath.NewInt(25000),
			ExternalChainInfos: chainInfos,
		}
	}
	return validators
}

func buildKeeper(t *testing.T) (*Keeper, sdk.Context, mockedServices) {
	k, mockServices, ctx := NewEvmKeeper(t)

	unpublishedSnapshot := &valsettypes.Snapshot{
		Id:          1,
		TotalShares: sdkmath.NewInt(75000),
		Validators: getValidators(
			3,
			[]validatorChainInfo{
				{
					chainType:        "evm",
					chainReferenceID: "test-chain",
				},
			},
		),
	}
	// test-chain mocks
	mockServices.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(unpublishedSnapshot, nil)
	mockServices.ConsensusKeeper.On("PutMessageInQueue", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uint64(0), nil)
	mockServices.GravityKeeper.On("GetLastObservedEventNonce", mock.Anything).Return(uint64(100), nil)

	// invalid-test-chain mocks
	mockServices.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(unpublishedSnapshot, nil)
	mockServices.ConsensusKeeper.On("PutMessageInQueue", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uint64(0), nil)
	mockServices.GravityKeeper.On("GetLastObservedEventNonce", mock.Anything).Return(uint64(100), nil)

	// Add 2 new chains for our tests to use
	err := k.AddSupportForNewChain(
		ctx,
		"test-chain",
		1,
		uint64(123),
		"0x1234",
		big.NewInt(55),
	)
	require.NoError(t, err)

	sc, err := k.SaveNewSmartContract(ctx, contractAbi, common.FromHex(contractBytecodeStr))
	require.NoError(t, err)
	err = k.SetAsCompassContract(ctx, sc)
	require.NoError(t, err)

	dep, _ := k.getSmartContractDeploymentByContractID(ctx, sc.GetId(), "test-chain")
	require.NotNil(t, dep)

	err = k.ActivateChainReferenceID(
		ctx,
		"test-chain",
		sc,
		"0x1234",
		dep.GetUniqueID(),
	)
	require.NoError(t, err)

	err = k.AddSupportForNewChain(
		ctx,
		"inactive-test-chain",
		2,
		uint64(123),
		"",
		big.NewInt(55),
	)
	require.NoError(t, err)

	sc, err = k.SaveNewSmartContract(ctx, contractAbi, common.FromHex(contractBytecodeStr))
	require.NoError(t, err)
	err = k.SetAsCompassContract(ctx, sc)
	require.NoError(t, err)

	dep, _ = k.getSmartContractDeploymentByContractID(ctx, sc.GetId(), "test-chain")
	require.NotNil(t, dep)

	err = k.ActivateChainReferenceID(
		ctx,
		"test-chain",
		sc,
		"0x1234",
		dep.GetUniqueID(),
	)
	require.NoError(t, err)

	return k, ctx, mockServices
}

func TestKeeper_PreJobExecution(t *testing.T) {
	testcases := []struct {
		name             string
		chainReferenceID string
		setupMocks       func(sdk.Context, *Keeper)
		expectedError    error
	}{
		{
			name:             "publishes most recent valset",
			chainReferenceID: "test-chain",
			setupMocks: func(ctx sdk.Context, k *Keeper) {
				valsetKeeperMock := mocks.NewValsetKeeper(t)
				msgSenderMock := mocks.NewMsgSender(t)
				gravityKeeperMock := mocks.NewGravityKeeper(t)

				unpublishedSnapshot := &valsettypes.Snapshot{
					Id:          1,
					TotalShares: sdkmath.NewInt(75000),
					Validators: []valsettypes.Validator{
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdkmath.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdkmath.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdkmath.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "test-chain",
								},
							},
						},
					},
				}
				valsetKeeperMock.On("GetCurrentSnapshot", mock.Anything).Return(unpublishedSnapshot, nil)

				publishedSnapshot := &valsettypes.Snapshot{
					Id:     3,
					Chains: []string{"test-chain"},
				}
				valsetKeeperMock.On("GetLatestSnapshotOnChain", mock.Anything, mock.Anything).Return(publishedSnapshot, nil)

				msgSenderMock.On(
					"SendValsetMsgForChain",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				k.Valset = valsetKeeperMock
				k.msgSender = msgSenderMock
				k.Gravity = gravityKeeperMock
			},
			expectedError: nil,
		},
		{
			name:             "no snapshot exists yet, return an error",
			chainReferenceID: "test-chain",
			setupMocks: func(ctx sdk.Context, k *Keeper) {
				valsetKeeperMock := mocks.NewValsetKeeper(t)
				valsetKeeperMock.On("GetCurrentSnapshot", mock.Anything).Return(nil, nil)
				k.Valset = valsetKeeperMock
			},
			expectedError: errors.New("nil, nil returned from Valset.GetCurrentSnapshot"),
		},
		{
			name:             "already using most recent published snapshot.  do nothing",
			chainReferenceID: "test-chain",
			setupMocks: func(ctx sdk.Context, k *Keeper) {
				valsetKeeperMock := mocks.NewValsetKeeper(t)
				gravityKeeperMock := mocks.NewGravityKeeper(t)

				publishedSnapshot := &valsettypes.Snapshot{
					Id:          1,
					Chains:      []string{"test-chain"},
					TotalShares: sdkmath.NewInt(75000),
					Validators: []valsettypes.Validator{
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdkmath.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdkmath.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdkmath.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "test-chain",
								},
							},
						},
					},
				}
				valsetKeeperMock.On("GetCurrentSnapshot", mock.Anything).Return(publishedSnapshot, nil)

				valsetKeeperMock.On("GetLatestSnapshotOnChain", mock.Anything, mock.Anything).Return(publishedSnapshot, nil)

				k.Valset = valsetKeeperMock
				k.Gravity = gravityKeeperMock
				// Success is indicated by returning nil before calling msgSender.SendValsetMsgForChain
			},
			expectedError: nil,
		},
		{
			name:             "inactive chain.  do nothing",
			chainReferenceID: "inactive-test-chain",
			setupMocks: func(ctx sdk.Context, k *Keeper) {
				valsetKeeperMock := mocks.NewValsetKeeper(t)
				gravityKeeperMock := mocks.NewGravityKeeper(t)

				unpublishedSnapshot := &valsettypes.Snapshot{
					Id:          1,
					TotalShares: sdkmath.NewInt(75000),
					Validators: []valsettypes.Validator{
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdkmath.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "inactive-test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdkmath.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "inactive-test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdkmath.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "inactive-test-chain",
								},
							},
						},
					},
				}
				valsetKeeperMock.On("GetCurrentSnapshot", mock.Anything).Return(unpublishedSnapshot, nil)

				publishedSnapshot := &valsettypes.Snapshot{
					Id:     3,
					Chains: []string{"inactive-test-chain"},
				}
				valsetKeeperMock.On("GetLatestSnapshotOnChain", mock.Anything, mock.Anything).Return(publishedSnapshot, nil)

				k.Valset = valsetKeeperMock
				k.Gravity = gravityKeeperMock
				// Success is indicated by returning nil before calling msgSender.SendValsetMsgForChain
			},
			expectedError: nil,
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx, _ := buildKeeper(t)
			tt.setupMocks(ctx, k)
			job := &schedulertypes.Job{
				ID: "test_job_1",
				Routing: schedulertypes.Routing{
					ChainType:        "evm",
					ChainReferenceID: tt.chainReferenceID,
				},
				IsPayloadModifiable: false,
			}

			actualErr := k.PreJobExecution(ctx, job)

			// The real assertions we're making is that the mocks call the correct functions for the path
			asserter.Equal(tt.expectedError, actualErr)
		})
	}
}

func TestKeeper_MissingChains(t *testing.T) {
	testcases := []struct {
		name                   string
		inputChainReferenceIDs []string
		setup                  func(sdk.Context, *Keeper)
		expected               []string
		expectedError          error
	}{
		{
			name: "Returns a list of chains that are missing - 2 chain missing, inactive chain ignored",
			inputChainReferenceIDs: []string{
				"test-chain",
			},
			setup: func(ctx sdk.Context, k *Keeper) {
				for i, chainId := range []string{"test-chain-2", "test-chain-3"} {
					err := k.AddSupportForNewChain(
						ctx,
						chainId,
						uint64(i+3), // 2 chains already set up by keeper
						uint64(i+100),
						"",
						big.NewInt(55),
					)
					require.NoError(t, err)

					// Activate chain
					chainInfo, err := k.GetChainInfo(ctx, chainId)
					require.NoError(t, err)

					chainInfo.SmartContractAddr = "0x1234"

					err = k.updateChainInfo(ctx, chainInfo)
					require.NoError(t, err)
				}
			},
			expected: []string{
				"test-chain-2",
				"test-chain-3",
			},
		},
		{
			name: "Returns a list of chains that are missing - extra chain in input ignored",
			inputChainReferenceIDs: []string{
				"test-chain",
				"extra-chain",
				"test-chain-2",
			},
			setup: func(ctx sdk.Context, k *Keeper) {
				for i, chainId := range []string{"test-chain-2", "test-chain-3"} {
					err := k.AddSupportForNewChain(
						ctx,
						chainId,
						uint64(i+3), // 2 chains already set up by keeper
						uint64(i+100),
						"",
						big.NewInt(55),
					)
					require.NoError(t, err)

					// Activate chain
					chainInfo, err := k.GetChainInfo(ctx, chainId)
					require.NoError(t, err)

					chainInfo.SmartContractAddr = "0x1234"

					err = k.updateChainInfo(ctx, chainInfo)
					require.NoError(t, err)
				}
			},
			expected: []string{
				"test-chain-3",
			},
		},
		{
			name: "Returns a list of chains that are missing - nil slice when matching",
			inputChainReferenceIDs: []string{
				"test-chain",
				"test-chain-2",
				"test-chain-3",
			},
			setup: func(ctx sdk.Context, k *Keeper) {
				for i, chainId := range []string{"test-chain-2", "test-chain-3"} {
					err := k.AddSupportForNewChain(
						ctx,
						chainId,
						uint64(i+3), // 2 chains already set up by keeper
						uint64(i+100),
						"",
						big.NewInt(55),
					)
					require.NoError(t, err)

					// Activate chain
					chainInfo, err := k.GetChainInfo(ctx, chainId)
					require.NoError(t, err)

					chainInfo.SmartContractAddr = "0x1234"

					err = k.updateChainInfo(ctx, chainInfo)
					require.NoError(t, err)
				}
			},
			expected: []string(nil),
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx, _ := buildKeeper(t)
			tt.setup(ctx, k)

			actual, actualErr := k.MissingChains(ctx, tt.inputChainReferenceIDs)
			asserter.Equal(tt.expected, actual)
			asserter.Equal(len(tt.expected), len(actual))
			asserter.Equal(tt.expectedError, actualErr)
		})
	}
}

func TestKeeper_PublishSnapshotToAllChains(t *testing.T) {
	testcases := []struct {
		name          string
		setup         func(sdk.Context, *Keeper, mockedServices)
		forcePublish  bool
		expectedError error
	}{
		{
			name: "Publishes when no previous snapshot on the chain",
			setup: func(ctx sdk.Context, k *Keeper, ms mockedServices) {
				ms.ValsetKeeper.On("GetLatestSnapshotOnChain", mock.Anything, mock.Anything).Return(nil, nil)
				// SendValsetMsgForChain indicates a publish
				ms.MsgSender.On("SendValsetMsgForChain", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "Doesn't publish when latest valset not over a month old and not forcePublish",
			setup: func(ctx sdk.Context, k *Keeper, ms mockedServices) {
				validators := getValidators(
					3,
					[]validatorChainInfo{
						{
							chainType:        "evm",
							chainReferenceID: "test-chain",
						},
						{
							chainType:        "evm",
							chainReferenceID: "test-chain",
						},
					},
				)
				publishedSnapshot := &valsettypes.Snapshot{
					Id:          1,
					Chains:      []string{"test-chain"},
					TotalShares: sdkmath.NewInt(75000),
					Validators:  validators,
					CreatedAt:   time.Now().Add(time.Duration(-28*24) * time.Hour), // 28 days ago
				}

				ms.ValsetKeeper.On("GetLatestSnapshotOnChain", mock.Anything, mock.Anything).Return(publishedSnapshot, nil)
				// Lack of a call to SendValsetMsgForChain indicates no publish
			},
		},
		{
			name: "Publishes regardless of age when force publish requested",
			setup: func(ctx sdk.Context, k *Keeper, ms mockedServices) {
				validators := getValidators(
					3,
					[]validatorChainInfo{
						{
							chainType:        "evm",
							chainReferenceID: "test-chain",
						},
						{
							chainType:        "evm",
							chainReferenceID: "test-chain",
						},
					},
				)
				publishedSnapshot := &valsettypes.Snapshot{
					Id:          1,
					Chains:      []string{"test-chain"},
					TotalShares: sdkmath.NewInt(75000),
					Validators:  validators,
					CreatedAt:   time.Now().Add(time.Duration(-28*24) * time.Hour), // 28 days ago
				}

				ms.ValsetKeeper.On("GetLatestSnapshotOnChain", mock.Anything, mock.Anything).Return(publishedSnapshot, nil)
				// SendValsetMsgForChain indicates a publish
				ms.MsgSender.On("SendValsetMsgForChain", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			forcePublish: true,
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx, mockServices := buildKeeper(t)
			tt.setup(ctx, k, mockServices)

			ctx = ctx.WithBlockTime(time.Now())
			newSnapshot := &valsettypes.Snapshot{
				Id:          2,
				TotalShares: sdkmath.NewInt(75000),
				Validators: getValidators(
					3,
					[]validatorChainInfo{
						{
							chainType:        "evm",
							chainReferenceID: "test-chain",
						},
					},
				),
			}

			actualErr := k.PublishSnapshotToAllChains(ctx, newSnapshot, tt.forcePublish)
			asserter.Equal(tt.expectedError, actualErr)
		})
	}
}
