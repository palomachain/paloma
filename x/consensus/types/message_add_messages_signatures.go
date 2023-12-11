package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
)

const TypeMsgAddMessagesSignatures = "add_messages_signatures"

var _ sdk.Msg = &MsgAddMessagesSignatures{}

func NewMsgAddMessagesSignatures(creator string) *MsgAddMessagesSignatures {
	return &MsgAddMessagesSignatures{
		Creator: creator,
	}
}

func (msg *MsgAddMessagesSignatures) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgAddMessagesSignatures) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
