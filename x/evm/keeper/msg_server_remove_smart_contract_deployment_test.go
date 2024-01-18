package keeper

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/palomachain/paloma/x/evm/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func addDeploymentToKeeper(t *testing.T, ctx sdk.Context, k *Keeper, mockServices mockedServices) {
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
	mockServices.ConsensusKeeper.On("PutMessageInQueue", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uint64(10), nil)
	mockServices.GravityKeeper.On("GetLastObservedEventNonce", mock.Anything).Return(uint64(100), nil)

	// Add a new chains for our test to use
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
}

func TestKeeper_RemoveSmartContractDeployment(t *testing.T) {
	t.Run("removes a smart contract deployment", func(t *testing.T) {
		k, mockServices, ctx := NewEvmKeeper(t)
		addDeploymentToKeeper(t, ctx, k, mockServices)

		deployments, err := k.AllSmartContractsDeployments(ctx)
		require.Len(t, deployments, 1)
		require.NoError(t, err)

		_, err = k.RemoveSmartContractDeployment(ctx, &types.MsgRemoveSmartContractDeploymentRequest{
			SmartContractID:  1,
			ChainReferenceID: "test-chain",
		})
		require.NoError(t, err)

		deployments, err = k.AllSmartContractsDeployments(ctx)
		require.Len(t, deployments, 0)
		require.NoError(t, err)
	})
}
