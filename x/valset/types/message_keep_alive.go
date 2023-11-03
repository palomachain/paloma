package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
)

var _ sdk.Msg = &MsgKeepAlive{}

func (msg *MsgKeepAlive) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgKeepAlive) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgKeepAlive) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
