package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
)

const TypeMsgAddEvidence = "add_evidence"

var _ sdk.Msg = &MsgAddEvidence{}

func (msg *MsgAddEvidence) Route() string {
	return RouterKey
}

func (msg *MsgAddEvidence) Type() string {
	return TypeMsgAddEvidence
}

func (msg *MsgAddEvidence) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgAddEvidence) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddEvidence) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
