package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdker "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAddStatusUpdate = "add_status_update"

var _ sdk.Msg = &MsgAddStatusUpdate{}

func (msg *MsgAddStatusUpdate) Route() string {
	return RouterKey
}

func (msg *MsgAddStatusUpdate) Type() string {
	return TypeMsgAddStatusUpdate
}

func (msg *MsgAddStatusUpdate) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddStatusUpdate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddStatusUpdate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdker.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
