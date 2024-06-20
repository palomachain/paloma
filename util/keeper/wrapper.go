package keeper

import (
	"context"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

type Byter interface {
	Bytes() []byte
}

func NewKvStoreWrapper[T proto.Message](s StoreGetterFn, ps ProtoSerializer) *kvStoreWrapper[T] {
	return &kvStoreWrapper[T]{
		sg: s,
		ps: ps,
	}
}

//go:generate mockery --name=KVStoreWrapper
type KVStoreWrapper[T proto.Message] interface {
	Get(ctx types.Context, key Byter) (T, error)
	Set(ctx types.Context, key Byter, value T) error
	Iterate(ctx types.Context, fn func([]byte, T) bool) error
}

type kvStoreWrapper[T proto.Message] struct {
	sg StoreGetterFn
	ps ProtoSerializer
}

func (v *kvStoreWrapper[T]) Get(ctx types.Context, key Byter) (T, error) {
	return Load[T](v.sg(ctx), v.ps, key.Bytes())
}

func (v *kvStoreWrapper[T]) Set(ctx types.Context, key Byter, value T) error {
	return Save(v.sg(ctx), v.ps, key.Bytes(), value)
}

func (v *kvStoreWrapper[T]) Iterate(ctx types.Context, fn func([]byte, T) bool) error {
	return IterAllFnc(v.sg(ctx), v.ps, fn)
}

func StoreFactory(storeKey corestore.KVStoreService, p string) func(ctx context.Context) storetypes.KVStore {
	return func(ctx context.Context) storetypes.KVStore {
		s := runtime.KVStoreAdapter(storeKey.OpenKVStore(ctx))
		return prefix.NewStore(s, []byte(p))
	}
}
