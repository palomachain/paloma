package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
)

const TypeMsgRegisterLightNodeClient = "register_light_node_client"

var _ sdk.Msg = &MsgRegisterLightNodeClient{}

func (msg *MsgRegisterLightNodeClient) Route() string {
	return RouterKey
}

func (msg *MsgRegisterLightNodeClient) Type() string {
	return TypeMsgRegisterLightNodeClient
}

func (msg *MsgRegisterLightNodeClient) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgRegisterLightNodeClient) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRegisterLightNodeClient) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
