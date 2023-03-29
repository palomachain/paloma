package types

import (
	"encoding/binary"

	proto "github.com/cosmos/gogoproto/proto"
)

func Uint64ToByte(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

type ConsensusMsg interface {
	proto.Message
}
