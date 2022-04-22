package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAddExternalChainInfoForValidator = "add_external_chain_info_for_validator"

var _ sdk.Msg = &MsgAddExternalChainInfoForValidator{}

func (msg *MsgAddExternalChainInfoForValidator) Route() string {
	return RouterKey
}

func (msg *MsgAddExternalChainInfoForValidator) Type() string {
	return TypeMsgAddExternalChainInfoForValidator
}

func (msg *MsgAddExternalChainInfoForValidator) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddExternalChainInfoForValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddExternalChainInfoForValidator) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
