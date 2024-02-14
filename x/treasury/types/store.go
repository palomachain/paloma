package types

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TreasuryStore interface {
	TreasuryStore(ctx sdk.Context) sdk.KVStore
	RelayerFeeStore(ctx sdk.Context) sdk.KVStore
}

type Store struct {
	StoreKey storetypes.StoreKey
}

func (sg Store) TreasuryStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(sg.StoreKey), []byte("treasury"))
}

func (sg Store) RelayerFeeStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(sg.StoreKey), []byte("relayer-fees"))
}
