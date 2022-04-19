package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"reflect"
)

func IterAll[T codec.ProtoMarshaler](store sdk.KVStore, pu protoUnmarshaler) ([]T, error) {
	res := []T{}
	err := IterAllFnc(store, pu, func(val T) bool {
		res = append(res, val)
		return true
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func IterAllFnc[T codec.ProtoMarshaler](store sdk.KVStore, pu protoUnmarshaler, fnc func(T) bool) error {
	res := []T{}
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		iterData := iterator.Value()

		var val T
		// stupid reflection :(
		va := reflect.ValueOf(&val).Elem()
		v := reflect.New(va.Type().Elem())
		va.Set(v)

		if err := pu.Unmarshal(iterData, val); err != nil {
			return err
		}
		if !fnc(val) {
			return nil
		}

		res = append(res, val)
	}
	if err := iterator.Error(); err != nil {
		return err
	}

	return nil
}
