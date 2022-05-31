package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	consensus "github.com/palomachain/paloma/x/consensus/types"
)

var amino = codec.NewLegacyAmino()
var ModuleCdc = codec.NewAminoCodec(amino)

var _ consensus.ConsensusMsg = &SimpleMessage{}
var _ consensus.ConsensusMsg = &EvenSimplerMessage{}

func (msg *SimpleMessage) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg *SimpleMessage) Attest() {}

func (msg *EvenSimplerMessage) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg *SimpleMessage) ConsensusSignBytes() consensus.BytesToSignFunc {
	return consensus.TypedBytesToSign(func(m *SimpleMessage, nonce uint64) []byte {
		return append(m.GetSignBytes(), consensus.Uint64ToByte(nonce)...)
	})
}
