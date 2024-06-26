package keeper

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k Keeper) userSmartContractStore(
	ctx context.Context,
	addr string,
) storetypes.KVStore {
	kvstore := runtime.KVStoreAdapter(k.storeKey.OpenKVStore(ctx))
	return prefix.NewStore(kvstore, types.UserSmartContractStoreKey(addr))
}

func (k Keeper) UserSmartContracts(
	ctx context.Context,
	addr string,
) ([]*types.UserSmartContract, error) {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger())
	logger.WithFields("val_address", addr).Debug("list user smart contracts")

	st := k.userSmartContractStore(ctx, addr)
	_, all, err := keeperutil.IterAll[*types.UserSmartContract](st, k.cdc)
	return all, err
}

func (k Keeper) SaveUserSmartContract(
	ctx context.Context,
	addr string,
	c *types.UserSmartContract,
) (uint64, error) {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger())
	logger.WithFields("val_address", addr).Debug("save user smart contract")

	if err := c.Validate(); err != nil {
		logger.WithError(err).Warn("save with invalid smart contract")
		return 0, err
	}

	stKey := types.UserSmartContractStoreKey(addr)

	// Create a new contract to make sure fields are properly initialized
	contract := &types.UserSmartContract{
		ValAddress:       addr,
		Id:               k.ider.IncrementNextID(ctx, string(stKey)),
		Title:            c.Title,
		AbiJson:          c.AbiJson,
		Bytecode:         c.Bytecode,
		ConstructorInput: c.ConstructorInput,
	}

	key := types.UserSmartContractKey(contract.Id)

	st := k.userSmartContractStore(ctx, addr)
	return contract.Id, keeperutil.Save(st, k.cdc, key, contract)
}

func (k Keeper) DeleteUserSmartContract(
	ctx context.Context,
	addr string,
	id uint64,
) error {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger())
	logger.WithFields("val_address", addr, "id", id).
		Debug("delete user smart contract")

	key := types.UserSmartContractKey(id)

	st := k.userSmartContractStore(ctx, addr)

	if !st.Has(key) {
		return fmt.Errorf("contract not found %v", id)
	}

	st.Delete(key)
	return nil
}

func (k Keeper) CreateUserSmartContractDeployment(
	ctx context.Context,
	addr string,
	id uint64,
	targetChain string,
) (uint64, error) {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger())
	logger.WithFields("val_address", addr, "id", id, "xchain", targetChain).
		Debug("create user smart contract deployment")

	key := types.UserSmartContractKey(id)
	st := k.userSmartContractStore(ctx, addr)

	contract, err := keeperutil.Load[*types.UserSmartContract](st, k.cdc, key)
	if err != nil {
		return 0, err
	}

	deployment := &types.UserSmartContract_Deployment{
		ChainReferenceId: targetChain,
		Status:           types.DeploymentStatus_IN_FLIGHT,
	}

	found := false
	for i, dep := range contract.Deployments {
		if dep.ChainReferenceId == deployment.ChainReferenceId {
			// If we already have a deployment for this chain, we overwrite it
			contract.Deployments[i] = deployment
			found = true
			break
		}
	}

	if !found {
		// If we don't have it yet, we append a new one
		contract.Deployments = append(contract.Deployments, deployment)
	}

	if err := keeperutil.Save(st, k.cdc, key, contract); err != nil {
		return 0, err
	}

	userSmartContract := &types.UploadUserSmartContract{
		Bytecode:         common.FromHex(contract.Bytecode),
		Abi:              contract.AbiJson,
		ConstructorInput: common.FromHex(contract.ConstructorInput),
		Id:               id,
		ValAddress:       addr,
	}

	return k.AddUploadUserSmartContractToConsensus(ctx, targetChain, userSmartContract)
}

func (k Keeper) AddUploadUserSmartContractToConsensus(
	ctx context.Context,
	chainReferenceID string,
	userSmartContract *types.UploadUserSmartContract,
) (uint64, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	assignee, err := k.PickValidatorForMessage(ctx, chainReferenceID, nil)
	if err != nil {
		return 0, err
	}

	return k.ConsensusKeeper.PutMessageInQueue(
		ctx,
		consensustypes.Queue(
			types.ConsensusTurnstoneMessage,
			xchainType,
			chainReferenceID,
		),
		&types.Message{
			ChainReferenceID: chainReferenceID,
			Action: &types.Message_UploadUserSmartContract{
				UploadUserSmartContract: userSmartContract,
			},
			Assignee:              assignee,
			AssignedAtBlockHeight: sdkmath.NewInt(sdkCtx.BlockHeight()),
		}, nil)
}

func (k Keeper) SetUserSmartContractDeploymentActive(
	ctx context.Context,
	addr string,
	id uint64,
	targetChain string,
	contractAddr string,
) error {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger())
	logger.WithFields(
		"val_address", addr,
		"id", id,
		"xchain", targetChain,
		"contract_addr", contractAddr,
	).Debug("user smart contract deployment success")

	return k.finishUserSmartContractDeployment(ctx, addr, id, targetChain,
		contractAddr, types.DeploymentStatus_ACTIVE)
}

func (k Keeper) SetUserSmartContractDeploymentError(
	ctx context.Context,
	addr string,
	id uint64,
	targetChain string,
) error {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger())
	logger.WithFields(
		"val_address", addr,
		"id", id,
		"xchain", targetChain,
	).Debug("user smart contract deployment failed")

	return k.finishUserSmartContractDeployment(ctx, addr, id, targetChain,
		"", types.DeploymentStatus_ERROR)
}

func (k Keeper) finishUserSmartContractDeployment(
	ctx context.Context,
	addr string,
	id uint64,
	targetChain string,
	contractAddr string,
	status types.DeploymentStatus,
) error {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger())
	logger.WithFields("val_address", addr, "id", id, "xchain", targetChain).
		Debug("finish user smart contract deployment")

	key := types.UserSmartContractKey(id)
	st := k.userSmartContractStore(ctx, addr)

	contract, err := keeperutil.Load[*types.UserSmartContract](st, k.cdc, key)
	if err != nil {
		return err
	}

	for i := range contract.Deployments {
		if contract.Deployments[i].ChainReferenceId == targetChain {
			contract.Deployments[i].Status = status
			contract.Deployments[i].Address = contractAddr

			return keeperutil.Save(st, k.cdc, key, contract)
		}
	}

	return fmt.Errorf("contract %v not found for %v", id, targetChain)
}
