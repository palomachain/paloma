package keeper

import "github.com/cosmos/cosmos-sdk/codec"

type ProtoUnmarshaler interface {
	Unmarshal(bz []byte, ptr codec.ProtoMarshaler) error
}

type ProtoMarshaler interface {
	Marshal(ptr codec.ProtoMarshaler) ([]byte, error)
}
