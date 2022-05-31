package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSubmitNewJob = "submit_new_job"

var _ sdk.Msg = &MsgSubmitNewJob{}

func NewMsgSubmitNewJob(creator string) *MsgSubmitNewJob {
	return &MsgSubmitNewJob{
		Creator: creator,
	}
}

func (msg *MsgSubmitNewJob) Route() string {
	return RouterKey
}

func (msg *MsgSubmitNewJob) Type() string {
	return TypeMsgSubmitNewJob
}

func (msg *MsgSubmitNewJob) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSubmitNewJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitNewJob) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
