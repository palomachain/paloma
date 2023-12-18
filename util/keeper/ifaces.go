package keeper

import (
	"github.com/cosmos/gogoproto/proto"
)

type ProtoSerializer interface {
	ProtoMarshaler
	ProtoUnmarshaler
}

type ProtoUnmarshaler interface {
	Unmarshal(bz []byte, ptr proto.Message) error
}

type ProtoMarshaler interface {
	Marshal(ptr proto.Message) ([]byte, error)
}
