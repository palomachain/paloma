package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
)

const TypeMsgAddExternalChainInfoForValidator = "add_external_chain_info_for_validator"

var _ sdk.Msg = &MsgAddExternalChainInfoForValidator{}

func (msg *MsgAddExternalChainInfoForValidator) Route() string {
	return RouterKey
}

func (msg *MsgAddExternalChainInfoForValidator) Type() string {
	return TypeMsgAddExternalChainInfoForValidator
}

func (msg *MsgAddExternalChainInfoForValidator) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgAddExternalChainInfoForValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddExternalChainInfoForValidator) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
