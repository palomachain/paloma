package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types"
)

var _ Byter = Key("")

type Byter interface {
	Bytes() []byte
}

type Key string

func (k Key) Bytes() []byte {
	return []byte(k)
}

//go:generate mockery --name=KVStoreWrapper
type KVStoreWrapper[T codec.ProtoMarshaler] interface {
	Get(ctx context.Context, key Byter) (T, error)
	Set(ctx context.Context, key Byter, value T) error
	Iterate(ctx context.Context, fn func([]byte, T) bool) error
}

func NewKvStoreWrapper[T codec.ProtoMarshaler](s StoreGetterFn, ps ProtoSerializer) KVStoreWrapper[T] {
	return &kvStoreWrapper[T]{
		sg: s,
		ps: ps,
	}
}

type kvStoreWrapper[T codec.ProtoMarshaler] struct {
	sg StoreGetterFn
	ps ProtoSerializer
}

func (v *kvStoreWrapper[T]) Get(ctx context.Context, key Byter) (T, error) {
	return Load[T](v.sg(ctx), v.ps, key.Bytes())
}

func (v *kvStoreWrapper[T]) Set(ctx context.Context, key Byter, value T) error {
	return Save(v.sg(ctx), v.ps, key.Bytes(), value)
}

func (v *kvStoreWrapper[T]) Iterate(ctx context.Context, fn func([]byte, T) bool) error {
	return IterAllFnc[T](v.sg(ctx), v.ps, fn)
}

func StoreFactory(storeKey storetypes.StoreKey, p string) func(ctx context.Context) types.KVStore {
	return func(ctx context.Context) types.KVStore {
		sdkCtx := types.UnwrapSDKContext(ctx)
		return prefix.NewStore(sdkCtx.KVStore(storeKey), []byte(p))
	}
}
