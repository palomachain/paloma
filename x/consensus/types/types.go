package types

import (
	"encoding/binary"

	proto "github.com/gogo/protobuf/proto"
)

func uint64ToByte(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

type ConsensusMsg interface {
	proto.Message
	// returns which bytes need to get signed
	GetSignBytes() []byte
}

func (s *SignSmartContractExecute) GetSignBytes() []byte {
	return uint64ToByte(s.Id)
}
