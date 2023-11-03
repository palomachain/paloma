package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
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
	return libmeta.GetSigners(msg)
}

func (msg *MsgDeleteJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteJob) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
