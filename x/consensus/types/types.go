package types

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/codec"
	proto "github.com/cosmos/gogoproto/proto"
)

func Uint64ToByte(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

type CodecMarshaler interface {
	AnyUnpacker
	MarshalInterface(i proto.Message) ([]byte, error)
	UnmarshalInterface(bz []byte, ptr interface{}) error
	Unmarshal(bz []byte, ptr codec.ProtoMarshaler) error
}

type ConsensusMsg interface {
	proto.Message
}
