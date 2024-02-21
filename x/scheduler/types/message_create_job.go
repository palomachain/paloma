package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
	"github.com/palomachain/paloma/x/valset/types"
)

const TypeMsgCreateJob = "create_job"

var _ sdk.Msg = &MsgCreateJob{}

func NewMsgCreateJob(creator string) *MsgCreateJob {
	return &MsgCreateJob{
		Metadata: types.MsgMetadata{
			Creator: creator,
		},
	}
}

func (msg *MsgCreateJob) Route() string {
	return RouterKey
}

func (msg *MsgCreateJob) Type() string {
	return TypeMsgCreateJob
}

func (msg *MsgCreateJob) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgCreateJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateJob) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
