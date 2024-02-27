package keeper

import (
	"reflect"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/gogoproto/proto"
)

func IterAll[T proto.Message](store storetypes.KVStore, pu ProtoUnmarshaler) ([][]byte, []T, error) {
	res := []T{}
	keys := [][]byte{}
	err := IterAllFnc(store, pu, func(key []byte, val T) bool {
		res = append(res, val)
		keys = append(keys, key)
		return true
	})
	if err != nil {
		return nil, nil, err
	}
	return keys, res, nil
}

func IterAllRaw(store storetypes.KVStore, pu ProtoUnmarshaler) (keys [][]byte, values [][]byte, _err error) {
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, iterator.Key())
		values = append(values, iterator.Value())
	}
	return
}

func IterAllFnc[T proto.Message](store storetypes.KVStore, pu ProtoUnmarshaler, fnc func([]byte, T) bool) error {
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
		if !fnc(iterator.Key(), val) {
			return nil
		}
	}

	return nil
}
