package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
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
	contract *types.UserSmartContract,
) (uint64, error) {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger())
	logger.WithFields("val_address", addr).Debug("save user smart contract")

	stKey := types.UserSmartContractStoreKey(addr)
	contract.Id = k.ider.IncrementNextID(ctx, string(stKey))

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
