package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperUtilI[T codec.ProtoMarshaler] interface {
	Save(store sdk.KVStore, pm ProtoMarshaler, key []byte, val codec.ProtoMarshaler) error
	Load(store sdk.KVStore, pu ProtoUnmarshaler, key []byte) (T, error)
}

type KeeperUtil[T codec.ProtoMarshaler] struct {
	Val T
}

func (ku KeeperUtil[T]) Save(store sdk.KVStore, pm ProtoMarshaler, key []byte, val codec.ProtoMarshaler) error {
	return Save(store, pm, key, val)
}

func (ku KeeperUtil[T]) Load(store sdk.KVStore, pu ProtoUnmarshaler, key []byte) (T, error) {
	return Load[T](store, pu, key)
}
