package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"reflect"
)

func Save(store sdk.KVStore, pm protoMarshaler, key []byte, val codec.ProtoMarshaler) error {
	bytez, err := pm.Marshal(val)
	if err != nil {
		return err
	}
	store.Set(key, bytez)
	return nil
}

func Load[T codec.ProtoMarshaler](store sdk.KVStore, pu protoUnmarshaler, key []byte) (T, error) {
	var val T
	// stupid reflection :(
	va := reflect.ValueOf(&val).Elem()
	v := reflect.New(va.Type().Elem())
	va.Set(v)
	bytez := store.Get(key)

	if len(bytez) == 0 {
		return val, ErrNotFound
	}

	err := pu.Unmarshal(bytez, val)

	if err != nil {
		return val, err
	}

	return val, nil
}
