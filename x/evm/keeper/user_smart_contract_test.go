package keeper

import (
	"testing"

	"github.com/palomachain/paloma/testutil/sample"
	"github.com/palomachain/paloma/util/blocks"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestUserSmartContracts(t *testing.T) {
	k, ctx, _ := buildKeeper(t)
	valAddr1 := "cosmosvaloper1pzf9apnk8yw7pjw3v9vtmxvn6guhkslanh8r07"
	valAddr2 := "cosmosvaloper1jxx2ym4pgk2yw4jkffxjc0tgddd8qqqhwu2tpm"
	var id uint64

	t.Run("Return empty list when not initialized", func(t *testing.T) {
		contracts, err := k.UserSmartContracts(ctx, valAddr1)
		require.NoError(t, err)
		require.Empty(t, contracts)
	})

	t.Run("Save a contract with invalid ABI", func(t *testing.T) {
		contract := &types.UserSmartContract{
			Author:           valAddr1,
			Title:            "Test Contract",
			Bytecode:         "0x01",
			ConstructorInput: "0x00",
			AbiJson:          `[{}]`,
		}

		_, err := k.SaveUserSmartContract(ctx, valAddr1, contract)
		require.Error(t, err)
	})

	t.Run("Save a contract with invalid bytecode", func(t *testing.T) {
		contract := &types.UserSmartContract{
			Author:           valAddr1,
			Title:            "Test Contract",
			Bytecode:         "invalid",
			ConstructorInput: "0x00",
			AbiJson:          sample.SimpleABI,
		}

		_, err := k.SaveUserSmartContract(ctx, valAddr1, contract)
		require.Error(t, err)
	})

	t.Run("Save a contract with empty title", func(t *testing.T) {
		contract := &types.UserSmartContract{
			Author:           valAddr1,
			Title:            "",
			Bytecode:         "0x01",
			ConstructorInput: "0x00",
			AbiJson:          sample.SimpleABI,
		}

		_, err := k.SaveUserSmartContract(ctx, valAddr1, contract)
		require.Error(t, err)
	})

	t.Run("Save the contract for a user", func(t *testing.T) {
		contract := &types.UserSmartContract{
			Author:           valAddr1,
			Title:            "Test Contract",
			Bytecode:         "0x01",
			ConstructorInput: "0x00",
			AbiJson:          sample.SimpleABI,
		}

		var err error
		id, err = k.SaveUserSmartContract(ctx, valAddr1, contract)
		require.NoError(t, err)
		require.NotZero(t, id)

		expected := &types.UserSmartContract{
			Id:               id,
			Author:           contract.Author,
			Title:            contract.Title,
			Bytecode:         contract.Bytecode,
			ConstructorInput: contract.ConstructorInput,
			AbiJson:          contract.AbiJson,
		}

		actual, err := k.UserSmartContracts(ctx, valAddr1)
		require.NoError(t, err)
		require.Len(t, actual, 1)
		require.Equal(t, expected, actual[0])
	})

	t.Run("Return empty list for a different address", func(t *testing.T) {
		contracts, err := k.UserSmartContracts(ctx, valAddr2)
		require.NoError(t, err)
		require.Empty(t, contracts)
	})

	t.Run("Create a deployment", func(t *testing.T) {
		_, err := k.CreateUserSmartContractDeployment(ctx, valAddr1,
			id, "test-chain")
		require.NoError(t, err)

		contracts, err := k.UserSmartContracts(ctx, valAddr1)
		require.NoError(t, err)
		require.Len(t, contracts[0].Deployments, 1)

		expected := &types.UserSmartContract_Deployment{
			ChainReferenceId:     "test-chain",
			Status:               types.DeploymentStatus_IN_FLIGHT,
			Address:              "",
			CreatedAtBlockHeight: ctx.BlockHeight(),
			UpdatedAtBlockHeight: ctx.BlockHeight(),
		}

		require.Equal(t, expected, contracts[0].Deployments[0])
	})

	t.Run("Create a deployment for an unsupported chain", func(t *testing.T) {
		_, err := k.CreateUserSmartContractDeployment(ctx, valAddr1,
			id, "invalid-chain")
		require.Error(t, err)
	})

	t.Run("Set deployment to active", func(t *testing.T) {
		err := k.SetUserSmartContractDeploymentActive(ctx, valAddr1,
			id, ctx.BlockHeight(), "test-chain", "contract_addr")
		require.NoError(t, err)

		contracts, err := k.UserSmartContracts(ctx, valAddr1)
		require.NoError(t, err)
		require.Len(t, contracts[0].Deployments, 1)

		expected := &types.UserSmartContract_Deployment{
			ChainReferenceId:     "test-chain",
			Status:               types.DeploymentStatus_ACTIVE,
			Address:              "contract_addr",
			CreatedAtBlockHeight: ctx.BlockHeight(),
			UpdatedAtBlockHeight: ctx.BlockHeight(),
		}

		require.Equal(t, expected, contracts[0].Deployments[0])
	})

	t.Run("Set deployment to error", func(t *testing.T) {
		err := k.SetUserSmartContractDeploymentError(ctx, valAddr1,
			id, ctx.BlockHeight(), "test-chain")
		require.NoError(t, err)

		contracts, err := k.UserSmartContracts(ctx, valAddr1)
		require.NoError(t, err)
		require.Len(t, contracts[0].Deployments, 1)

		expected := &types.UserSmartContract_Deployment{
			ChainReferenceId:     "test-chain",
			Status:               types.DeploymentStatus_ERROR,
			Address:              "",
			CreatedAtBlockHeight: ctx.BlockHeight(),
			UpdatedAtBlockHeight: ctx.BlockHeight(),
		}

		require.Equal(t, expected, contracts[0].Deployments[0])
	})

	t.Run("Create a new deployment for the same chain", func(t *testing.T) {
		newCtx := ctx.WithBlockHeight(ctx.BlockHeight() + 1000)
		_, err := k.CreateUserSmartContractDeployment(newCtx, valAddr1,
			id, "test-chain")
		require.NoError(t, err)

		contracts, err := k.UserSmartContracts(ctx, valAddr1)
		require.NoError(t, err)
		require.Len(t, contracts[0].Deployments, 2)

		// Should now contain both
		expected := []*types.UserSmartContract_Deployment{{
			ChainReferenceId:     "test-chain",
			Status:               types.DeploymentStatus_ERROR,
			Address:              "",
			CreatedAtBlockHeight: ctx.BlockHeight(),
			UpdatedAtBlockHeight: ctx.BlockHeight(),
		}, {
			ChainReferenceId:     "test-chain",
			Status:               types.DeploymentStatus_IN_FLIGHT,
			Address:              "",
			CreatedAtBlockHeight: newCtx.BlockHeight(),
			UpdatedAtBlockHeight: newCtx.BlockHeight(),
		}}

		require.Equal(t, expected, contracts[0].Deployments)
	})

	t.Run("Update the right deployment", func(t *testing.T) {
		newCtx := ctx.WithBlockHeight(ctx.BlockHeight() + 1000)
		err := k.SetUserSmartContractDeploymentActive(newCtx, valAddr1,
			id, newCtx.BlockHeight(), "test-chain", "contract_addr")
		require.NoError(t, err)

		contracts, err := k.UserSmartContracts(ctx, valAddr1)
		require.NoError(t, err)
		require.Len(t, contracts[0].Deployments, 2)

		// Should now contain both
		expected := []*types.UserSmartContract_Deployment{{
			ChainReferenceId:     "test-chain",
			Status:               types.DeploymentStatus_ERROR,
			Address:              "",
			CreatedAtBlockHeight: ctx.BlockHeight(),
			UpdatedAtBlockHeight: ctx.BlockHeight(),
		}, {
			ChainReferenceId:     "test-chain",
			Status:               types.DeploymentStatus_ACTIVE,
			Address:              "contract_addr",
			CreatedAtBlockHeight: newCtx.BlockHeight(),
			UpdatedAtBlockHeight: newCtx.BlockHeight(),
		}}

		require.Equal(t, expected, contracts[0].Deployments)
	})

	t.Run("Delete the contract", func(t *testing.T) {
		err := k.DeleteUserSmartContract(ctx, valAddr1, id)
		require.NoError(t, err)

		contracts, err := k.UserSmartContracts(ctx, valAddr1)
		require.NoError(t, err)
		require.Empty(t, contracts)
	})
}

func TestPurgeUserSmartContracts(t *testing.T) {
	k, ctx, _ := buildKeeper(t)
	valAddr1 := "palomavaloper1tsu8nthuspe4zlkejtj3v27rtq8qz7q6983zt2"

	ctx = ctx.WithBlockHeight(10000)

	contract := &types.UserSmartContract{
		Author:           valAddr1,
		Title:            "Test Contract",
		Bytecode:         "0x01",
		ConstructorInput: "0x00",
		AbiJson:          sample.SimpleABI,
	}

	_, err := k.SaveUserSmartContract(ctx, valAddr1, contract)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + blocks.MonthlyHeight)

	_, err = k.SaveUserSmartContract(ctx, valAddr1, contract)
	require.NoError(t, err)

	contracts, err := k.UserSmartContracts(ctx, valAddr1)
	require.NoError(t, err)
	require.Len(t, contracts, 2)

	t.Run("Should remove only one contract", func(t *testing.T) {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 100)

		err = k.PurgeStaleUserSmartContracts(ctx)
		require.NoError(t, err)

		contracts, err = k.UserSmartContracts(ctx, valAddr1)
		require.NoError(t, err)
		require.Len(t, contracts, 1)
	})

	t.Run("Should remove the remaining contract", func(t *testing.T) {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + blocks.MonthlyHeight)

		err = k.PurgeStaleUserSmartContracts(ctx)
		require.NoError(t, err)

		contracts, err = k.UserSmartContracts(ctx, valAddr1)
		require.NoError(t, err)
		require.Empty(t, contracts)
	})
}
