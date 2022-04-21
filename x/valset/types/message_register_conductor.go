package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRegisterConductor = "register_conductor"

var _ sdk.Msg = &MsgRegisterConductor{}

func NewMsgRegisterConductor(creator string, pubKey, signedPubKey []byte) *MsgRegisterConductor {
	return &MsgRegisterConductor{
		Creator:      creator,
		PubKey:       pubKey,
		SignedPubKey: signedPubKey,
	}
}

func (msg *MsgRegisterConductor) Route() string {
	return RouterKey
}

func (msg *MsgRegisterConductor) Type() string {
	return TypeMsgRegisterConductor
}

func (msg *MsgRegisterConductor) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRegisterConductor) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRegisterConductor) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
