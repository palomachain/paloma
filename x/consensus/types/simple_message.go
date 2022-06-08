package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ ConsensusMsg = &SimpleMessage{}
var _ ConsensusMsg = &EvenSimplerMessage{}

func (msg *SimpleMessage) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg *SimpleMessage) Attest() {}

func (msg *EvenSimplerMessage) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg *SimpleMessage) ConsensusSignBytes() BytesToSignFunc {
	return TypedBytesToSign(func(m *SimpleMessage, nonce uint64) []byte {
		return append(m.GetSignBytes(), Uint64ToByte(nonce)...)
	})
}
