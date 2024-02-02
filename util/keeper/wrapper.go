package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
)

type Byter interface {
	Bytes() []byte
}

func NewKvStoreWrapper[T codec.ProtoMarshaler](s StoreGetterFn, ps ProtoSerializer) *KVStoreWrapper[T] {
	return &KVStoreWrapper[T]{
		sg: s,
		ps: ps,
	}
}

type KVStoreWrapper[T codec.ProtoMarshaler] struct {
	sg StoreGetterFn
	ps ProtoSerializer
}

func (v *KVStoreWrapper[T]) Get(ctx types.Context, key Byter) (T, error) {
	return Load[T](v.sg(ctx), v.ps, key.Bytes())
}

func (v *KVStoreWrapper[T]) Set(ctx types.Context, key Byter, value T) error {
	return Save(v.sg(ctx), v.ps, key.Bytes(), value)
}

func (v *KVStoreWrapper[T]) Iterate(ctx types.Context, fn func([]byte, T) bool) error {
	return IterAllFnc[T](v.sg(ctx), v.ps, fn)
}
