package types

import (
	"reflect"

	types "github.com/cosmos/cosmos-sdk/codec/types"
)

type AnyUnpacker = types.AnyUnpacker

type TypeChecker func(any) bool

func StaticTypeChecker(typ any) TypeChecker {
	return func(i any) bool {
		expectedType := reflect.TypeOf(typ)
		gotType := reflect.TypeOf(i)
		return expectedType == gotType
	}
}

func BatchedTypeChecker(static TypeChecker) TypeChecker {
	return func(i any) bool {
		batch, ok := i.(*Batch)
		if !ok {
			return false
		}
		for _, msg := range batch.GetMsgs() {
			var _t ConsensusMsg
			if err := ModuleCdc.UnpackAny(msg, &_t); err != nil {
				return false
			}
			if static(_t) == false {
				return false
			}
		}
		return true
	}
}
