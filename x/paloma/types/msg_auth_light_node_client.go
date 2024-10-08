package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/libmeta"
)

const TypeMsgAuthLightNodeClient = "auth_light_node_client"

var _ sdk.Msg = &MsgAuthLightNodeClient{}

func (msg *MsgAuthLightNodeClient) Route() string {
	return RouterKey
}

func (msg *MsgAuthLightNodeClient) Type() string {
	return TypeMsgAddLightNodeClientLicense
}

func (msg *MsgAuthLightNodeClient) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgAuthLightNodeClient) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAuthLightNodeClient) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
