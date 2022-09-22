package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCreateJob = "create_job"

var _ sdk.Msg = &MsgCreateJob{}

func NewMsgCreateJob(creator string) *MsgCreateJob {
	return &MsgCreateJob{
		Creator: creator,
	}
}

func (msg *MsgCreateJob) Route() string {
	return RouterKey
}

func (msg *MsgCreateJob) Type() string {
	return TypeMsgCreateJob
}

func (msg *MsgCreateJob) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateJob) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return msg.Job.ValidateBasic()
}
