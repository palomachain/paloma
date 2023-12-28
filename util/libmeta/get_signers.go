package libmeta

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Metadata interface {
	GetSigners() []string
	GetCreator() string
}

type MsgWithMetadata[T Metadata] interface {
	GetMetadata() T
}

func GetSigners[E Metadata, T MsgWithMetadata[E]](msg T) []sdk.AccAddress {
	md := msg.GetMetadata()
	signers := make([]sdk.AccAddress, len(md.GetSigners()))
	for i, v := range md.GetSigners() {
		signer, err := sdk.AccAddressFromBech32(v)
		if err != nil {
			panic(err)
		}
		signers[i] = signer
	}
	return signers
}
