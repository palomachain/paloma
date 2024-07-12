package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
)

const TypeMsgAddLightNodeClientFunds = "add_light_node_client_funds"

var _ sdk.Msg = &MsgAddLightNodeClientFunds{}

func (msg *MsgAddLightNodeClientFunds) Route() string {
	return RouterKey
}

func (msg *MsgAddLightNodeClientFunds) Type() string {
	return TypeMsgAddLightNodeClientFunds
}

func (msg *MsgAddLightNodeClientFunds) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgAddLightNodeClientFunds) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddLightNodeClientFunds) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
