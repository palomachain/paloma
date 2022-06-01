package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDeleteJob = "delete_job"

var _ sdk.Msg = &MsgDeleteJob{}

func NewMsgDeleteJob(creator string, queueTypeName string, messageID uint64) *MsgDeleteJob {
	return &MsgDeleteJob{
		Creator:       creator,
		QueueTypeName: queueTypeName,
		MessageID:     messageID,
	}
}

func (msg *MsgDeleteJob) Route() string {
	return RouterKey
}

func (msg *MsgDeleteJob) Type() string {
	return TypeMsgDeleteJob
}

func (msg *MsgDeleteJob) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteJob) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
