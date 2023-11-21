package types

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TreasuryStore interface {
	TreasuryStore(ctx sdk.Context) sdk.KVStore
}

type Store struct {
	StoreKey storetypes.StoreKey
}

func (sg Store) TreasuryStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(sg.StoreKey), []byte("treasury"))
}
