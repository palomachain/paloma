package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
	types "github.com/palomachain/paloma/x/valset/types"
)

const TypeMsgDeployNewSmartContractRequest = "deploy_new_smart_contract_request"

func NewMsgDeployNewSmartContractRequest(creator sdk.AccAddress, title string, description string, abiJSON string, bytecode string) *MsgDeployNewSmartContractRequest {
	return &MsgDeployNewSmartContractRequest{
		Title:       title,
		Description: description,
		AbiJSON:     abiJSON,
		BytecodeHex: bytecode,
		Metadata: types.MsgMetadata{
			Creator: creator.String(),
			Signers: []string{creator.String()},
		},
	}
}

func (msg *MsgDeployNewSmartContractRequest) Route() string {
	return RouterKey
}

func (msg *MsgDeployNewSmartContractRequest) Type() string {
	return TypeMsgDeployNewSmartContractRequest
}

func (msg *MsgDeployNewSmartContractRequest) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgDeployNewSmartContractRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeployNewSmartContractRequest) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}
