package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgExecuteJob = "execute_job"

var _ sdk.Msg = &MsgExecuteJob{}

func NewMsgExecuteJob(creator string, jobID string) *MsgExecuteJob {
	return &MsgExecuteJob{
		Creator: creator,
		JobID:   jobID,
	}
}

func (msg *MsgExecuteJob) Route() string {
	return RouterKey
}

func (msg *MsgExecuteJob) Type() string {
	return TypeMsgExecuteJob
}

func (msg *MsgExecuteJob) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgExecuteJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgExecuteJob) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
