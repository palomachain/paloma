package keeper

import (
	"reflect"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/gogoproto/proto"
)

func Save(store storetypes.KVStore, pm ProtoMarshaler, key []byte, val proto.Message) error {
	bytez, err := pm.Marshal(val)
	if err != nil {
		return err
	}
	store.Set(key, bytez)
	return nil
}

func Load[T proto.Message](store storetypes.KVStore, pu ProtoUnmarshaler, key []byte) (T, error) {
	var val T
	if key == nil {
		return val, ErrNotFound.Format(val, key)
	}
	// stupid reflection :(
	va := reflect.ValueOf(&val).Elem()
	v := reflect.New(va.Type().Elem())
	va.Set(v)
	bytez := store.Get(key)

	if len(bytez) == 0 {
		return val, ErrNotFound.Format(val, key)
	}

	err := pu.Unmarshal(bytez, val)
	if err != nil {
		return val, err
	}

	return val, nil
}
