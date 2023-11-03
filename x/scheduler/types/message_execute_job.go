package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
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
	return libmeta.GetSigners(msg)
}

func (msg *MsgExecuteJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgExecuteJob) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
