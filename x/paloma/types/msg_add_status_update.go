package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
)

const TypeMsgAddStatusUpdate = "add_status_update"

var _ sdk.Msg = &MsgAddStatusUpdate{}

func (msg *MsgAddStatusUpdate) Route() string {
	return RouterKey
}

func (msg *MsgAddStatusUpdate) Type() string {
	return TypeMsgAddStatusUpdate
}

func (msg *MsgAddStatusUpdate) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgAddStatusUpdate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddStatusUpdate) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
