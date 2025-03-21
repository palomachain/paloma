package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/palomachain/paloma/v2/util/libmeta"
)

// nolint: exhaustruct
var (
	_ sdk.Msg = &MsgRemoveSmartContractDeploymentRequest{}
	_ sdk.Msg = &MsgDeployNewSmartContractProposalV2{}
)

func (msg *MsgRemoveSmartContractDeploymentRequest) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}

func (msg *MsgRemoveSmartContractDeploymentRequest) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgRemoveSmartContractDeploymentRequest) Route() string { return RouterKey }

func (msg *MsgRemoveSmartContractDeploymentRequest) Type() string {
	return "remove_smart_contract_deployment"
}

func (msg *MsgRemoveSmartContractDeploymentRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg *MsgDeployNewSmartContractProposalV2) ValidateBasic() error {
	if len(msg.AbiJSON) < 1 || len(msg.BytecodeHex) < 1 {
		return ErrInvalid.Wrap("abi json or bytecode hex is empty")
	}

	if msg.Authority != msg.Metadata.Creator {
		return ErrUnauthorized.Wrapf("authority mismatch: %s", msg.Authority)
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalid.Wrap("authority address is invalid")
	}

	return libmeta.ValidateBasic(msg)
}

func (msg *MsgDeployNewSmartContractProposalV2) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgDeployNewSmartContractProposalV2) Route() string { return RouterKey }

func (msg *MsgDeployNewSmartContractProposalV2) Type() string {
	return "propose_new_smart_contract_deployment"
}

func (msg *MsgDeployNewSmartContractProposalV2) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (a *MsgDeployNewSmartContractProposalV2) Bytecode() []byte {
	return common.FromHex(a.GetBytecodeHex())
}

func (msg *MsgProposeNewReferenceBlockAttestation) ValidateBasic() error {
	if len(msg.BlockHash) < 1 || len(msg.ChainReferenceId) < 1 {
		return ErrInvalid.Wrap("block hash or chain reference id are empty")
	}

	if msg.BlockHeight < 1 {
		return ErrInvalid.Wrap("block height must be postive")
	}

	if msg.Authority != msg.Metadata.Creator {
		return ErrUnauthorized.Wrapf("authority mismatch: %s", msg.Authority)
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalid.Wrap("authority address is invalid")
	}

	return libmeta.ValidateBasic(msg)
}

func (msg *MsgProposeNewReferenceBlockAttestation) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgProposeNewReferenceBlockAttestation) Route() string { return RouterKey }

func (msg *MsgProposeNewReferenceBlockAttestation) Type() string {
	return "propose_new_reference_block_attestation"
}

func (msg *MsgProposeNewReferenceBlockAttestation) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}
