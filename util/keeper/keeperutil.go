package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

type KeeperUtilI[T codec.ProtoMarshaler] interface {
	Save(store storetypes.KVStore, pm ProtoMarshaler, key []byte, val codec.ProtoMarshaler) error
	Load(store storetypes.KVStore, pu ProtoUnmarshaler, key []byte) (T, error)
}

type KeeperUtil[T codec.ProtoMarshaler] struct {
	Val T
}

func (ku KeeperUtil[T]) Save(store storetypes.KVStore, pm ProtoMarshaler, key []byte, val codec.ProtoMarshaler) error {
	return Save(store, pm, key, val)
}

func (ku KeeperUtil[T]) Load(store storetypes.KVStore, pu ProtoUnmarshaler, key []byte) (T, error) {
	return Load[T](store, pu, key)
}
