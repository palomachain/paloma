package types

import (
	"reflect"
)

type (
	ValueValidatorFn func(value interface{}) error
	// ParamSetPair is used for associating paramsubspace key and field of param
	// structs.
	ParamSetPair struct {
		Key         []byte
		Value       interface{}
		ValidatorFn ValueValidatorFn
	}
)
type attribute struct {
	ty  reflect.Type
	vfn ValueValidatorFn
}

// KeyTable subspaces appropriate type for each parameter key
type KeyTable struct {
	m map[string]attribute
}

type ParamSetPairs []ParamSetPair

// ParamSet defines an interface for structs containing parameters for a module
type ParamSet interface {
	ParamSetPairs() ParamSetPairs
}
