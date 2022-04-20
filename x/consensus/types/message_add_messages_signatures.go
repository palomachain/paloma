package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAddMessagesSignatures = "add_messages_signatures"

var _ sdk.Msg = &MsgAddMessagesSignatures{}

func NewMsgAddMessagesSignatures(creator string) *MsgAddMessagesSignatures {
	return &MsgAddMessagesSignatures{
		Creator: creator,
	}
}

func (msg *MsgAddMessagesSignatures) Route() string {
	return RouterKey
}

func (msg *MsgAddMessagesSignatures) Type() string {
	return TypeMsgAddMessagesSignatures
}

func (msg *MsgAddMessagesSignatures) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddMessagesSignatures) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddMessagesSignatures) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
