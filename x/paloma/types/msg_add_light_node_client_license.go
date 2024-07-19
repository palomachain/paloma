package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
)

const TypeMsgAddLightNodeClientLicense = "add_light_node_client_license"

var _ sdk.Msg = &MsgAddLightNodeClientLicense{}

func (msg *MsgAddLightNodeClientLicense) Route() string {
	return RouterKey
}

func (msg *MsgAddLightNodeClientLicense) Type() string {
	return TypeMsgAddLightNodeClientLicense
}

func (msg *MsgAddLightNodeClientLicense) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgAddLightNodeClientLicense) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddLightNodeClientLicense) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
