package keeper

import (
	"errors"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/palomachain/paloma/x/evm/types/mocks"
	schedulertypes "github.com/palomachain/paloma/x/scheduler/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildKeeper(t *testing.T) (*Keeper, sdk.Context) {
	k, mockServices, ctx := NewEvmKeeper(t)

	unpublishedSnapshot := &valsettypes.Snapshot{
		Id:          1,
		TotalShares: sdk.NewInt(75000),
		Validators: []valsettypes.Validator{
			{
				State:      valsettypes.ValidatorState_ACTIVE,
				ShareCount: sdk.NewInt(25000),
				ExternalChainInfos: []*valsettypes.ExternalChainInfo{
					{
						ChainType:        "evm",
						ChainReferenceID: "test-chain",
					},
				},
			},
			{
				State:      valsettypes.ValidatorState_ACTIVE,
				ShareCount: sdk.NewInt(25000),
				ExternalChainInfos: []*valsettypes.ExternalChainInfo{
					{
						ChainType:        "evm",
						ChainReferenceID: "test-chain",
					},
				},
			},
			{
				State:      valsettypes.ValidatorState_ACTIVE,
				ShareCount: sdk.NewInt(25000),
				ExternalChainInfos: []*valsettypes.ExternalChainInfo{
					{
						ChainType:        "evm",
						ChainReferenceID: "test-chain",
					},
				},
			},
		},
	}
	// test-chain mocks
	mockServices.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(unpublishedSnapshot, nil)
	mockServices.ConsensusKeeper.On("PutMessageInQueue", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// invalid-test-chain mocks
	mockServices.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(unpublishedSnapshot, nil)
	mockServices.ConsensusKeeper.On("PutMessageInQueue", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

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

	dep, _ := k.getSmartContractDeploying(ctx, sc.GetId(), "test-chain")
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

	dep, _ = k.getSmartContractDeploying(ctx, sc.GetId(), "test-chain")
	require.NotNil(t, dep)

	err = k.ActivateChainReferenceID(
		ctx,
		"test-chain",
		sc,
		"0x1234",
		dep.GetUniqueID(),
	)
	require.NoError(t, err)

	return k, ctx
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

				unpublishedSnapshot := &valsettypes.Snapshot{
					Id:          1,
					TotalShares: sdk.NewInt(75000),
					Validators: []valsettypes.Validator{
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdk.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdk.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdk.NewInt(25000),
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

				publishedSnapshot := &valsettypes.Snapshot{
					Id:          1,
					Chains:      []string{"test-chain"},
					TotalShares: sdk.NewInt(75000),
					Validators: []valsettypes.Validator{
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdk.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdk.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdk.NewInt(25000),
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
				// Success is indicated by returning nil before calling msgSender.SendValsetMsgForChain
			},
			expectedError: nil,
		},
		{
			name:             "inactive chain.  do nothing",
			chainReferenceID: "inactive-test-chain",
			setupMocks: func(ctx sdk.Context, k *Keeper) {
				valsetKeeperMock := mocks.NewValsetKeeper(t)

				unpublishedSnapshot := &valsettypes.Snapshot{
					Id:          1,
					TotalShares: sdk.NewInt(75000),
					Validators: []valsettypes.Validator{
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdk.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "inactive-test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdk.NewInt(25000),
							ExternalChainInfos: []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "inactive-test-chain",
								},
							},
						},
						{
							State:      valsettypes.ValidatorState_ACTIVE,
							ShareCount: sdk.NewInt(25000),
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
				// Success is indicated by returning nil before calling msgSender.SendValsetMsgForChain
			},
			expectedError: nil,
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			// ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			k, ctx := buildKeeper(t)
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
