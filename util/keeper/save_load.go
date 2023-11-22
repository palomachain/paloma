package keeper

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "cosmossdk.io/store/types"

)

func Save(store storetypes.KVStore, pm ProtoMarshaler, key []byte, val codec.ProtoMarshaler) error {
	bytez, err := pm.Marshal(val)
	if err != nil {
		return err
	}
	store.Set(key, bytez)
	return nil
}

func Load[T codec.ProtoMarshaler](store storetypes.KVStore, pu ProtoUnmarshaler, key []byte) (T, error) {
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
