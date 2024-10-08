package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/libmeta"
)

const TypeMsgSetLegacyLightNodeClients = "set_legacy_light_node_clients"

var _ sdk.Msg = &MsgSetLegacyLightNodeClients{}

func (msg *MsgSetLegacyLightNodeClients) Route() string {
	return RouterKey
}

func (msg *MsgSetLegacyLightNodeClients) Type() string {
	return TypeMsgAddLightNodeClientLicense
}

func (msg *MsgSetLegacyLightNodeClients) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgSetLegacyLightNodeClients) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetLegacyLightNodeClients) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
