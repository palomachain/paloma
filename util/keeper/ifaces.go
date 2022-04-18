package keeper

import "github.com/cosmos/cosmos-sdk/codec"

type protoUnmarshaler interface {
	Unmarshal(bz []byte, ptr codec.ProtoMarshaler) error
}

type protoMarshaler interface {
	Marshal(ptr codec.ProtoMarshaler) ([]byte, error)
}
