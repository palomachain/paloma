package types

import "reflect"

type TypeChecker func(any) bool

func StaticTypeChecker(typ any) TypeChecker {
	return func(i any) bool {
		expectedType := reflect.TypeOf(typ)
		gotType := reflect.TypeOf(i)
		return expectedType == gotType
	}
}
