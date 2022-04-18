package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

		var typ T
		if err := pu.Unmarshal(iterData, typ); err != nil {
			return err
		}
		if !fnc(typ) {
			return nil
		}

		res = append(res, typ)
	}
	if err := iterator.Error(); err != nil {
		return err
	}

	return nil
}
