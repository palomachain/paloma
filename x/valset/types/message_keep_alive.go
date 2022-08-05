package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgKeepAlive = "keep_alive"

var _ sdk.Msg = &MsgKeepAlive{}

func NewMsgKeepAlive(creator string) *MsgKeepAlive {
	return &MsgKeepAlive{
		Creator: creator,
	}
}

func (msg *MsgKeepAlive) Route() string {
	return RouterKey
}

func (msg *MsgKeepAlive) Type() string {
	return TypeMsgKeepAlive
}

func (msg *MsgKeepAlive) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgKeepAlive) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgKeepAlive) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
