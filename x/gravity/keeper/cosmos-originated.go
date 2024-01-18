package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/store/prefix"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	evmtypes "github.com/palomachain/paloma/x/evm/types"
	"github.com/palomachain/paloma/x/gravity/types"
)

func (k Keeper) GetDenomOfERC20(ctx context.Context, chainReferenceId string, tokenContract types.EthAddress) (string, error) {
	erc20ToDenom, err := keeperutil.Load[*types.ERC20ToDenom](k.GetStore(ctx), k.cdc, types.GetERC20ToDenomKey(chainReferenceId, tokenContract))
	if errors.Is(err, keeperutil.ErrNotFound) {
		return "", types.ErrDenomNotFound
	}

	if err != nil {
		return "", err
	}

	return erc20ToDenom.Denom, nil
}

func (k Keeper) GetERC20OfDenom(ctx context.Context, chainReferenceId, denom string) (*types.EthAddress, error) {
	erc20ToDenom, err := keeperutil.Load[*types.ERC20ToDenom](k.GetStore(ctx), k.cdc, types.GetDenomToERC20Key(chainReferenceId, denom))
	if errors.Is(err, keeperutil.ErrNotFound) {
		return nil, types.ErrERC20NotFound
	}

	if err != nil {
		return nil, err
	}

	ethAddr, err := types.NewEthAddress(erc20ToDenom.Erc20)
	if err != nil {
		return nil, err
	}

	return ethAddr, nil
}

func (k Keeper) GetAllERC20ToDenoms(ctx context.Context) ([]*types.ERC20ToDenom, error) {
	store := k.GetStore(ctx)
	prefixStore := prefix.NewStore(store, types.DenomToERC20Key)
	_, all, err := keeperutil.IterAll[*types.ERC20ToDenom](prefixStore, k.cdc)

	return all, err
}

func (k Keeper) CastAllERC20ToDenoms(ctx context.Context) ([]evmtypes.ERC20Record, error) {
	all, err := k.GetAllERC20ToDenoms(ctx)
	if err != nil {
		return nil, err
	}

	cast := make([]evmtypes.ERC20Record, len(all))
	for i, v := range all {
		cast[i] = v
	}

	return cast, nil
}

func (k Keeper) setDenomToERC20(ctx context.Context, chainReferenceId, denom string, tokenContract types.EthAddress) error {
	store := k.GetStore(ctx)

	denomToERC20 := types.ERC20ToDenom{
		ChainReferenceId: chainReferenceId,
		Denom:            denom,
		Erc20:            tokenContract.GetAddress().String(),
	}

	err := keeperutil.Save(store, k.cdc, types.GetDenomToERC20Key(chainReferenceId, denom), &denomToERC20)
	if err != nil {
		return err
	}

	return keeperutil.Save(store, k.cdc, types.GetERC20ToDenomKey(chainReferenceId, tokenContract), &denomToERC20)
}

func (k Keeper) GetAllDenomToERC20s(ctx context.Context) ([]*types.ERC20ToDenom, error) {
	store := k.GetStore(ctx)
	prefixStore := prefix.NewStore(store, types.DenomToERC20Key)
	_, all, err := keeperutil.IterAll[*types.ERC20ToDenom](prefixStore, k.cdc)

	return all, err
}
