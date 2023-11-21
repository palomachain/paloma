package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// ModuleCdc is the codec for the module
var ModuleCdc = codec.NewLegacyAmino()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterInterfaces registers the interfaces for the proto stuff
// nolint: exhaustruct
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSendToEth{},
		&MsgConfirmBatch{},
		&MsgSendToPalomaClaim{},
		&MsgBatchSendToEthClaim{},
		&MsgCancelSendToEth{},
		&MsgSubmitBadSignatureEvidence{},
	)

	registry.RegisterInterface(
		"palomachain.paloma.gravity.EthereumClaim",
		(*EthereumClaim)(nil),
		&MsgSendToPalomaClaim{},
		&MsgBatchSendToEthClaim{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&SetERC20ToDenomProposal{},
	)

	registry.RegisterInterface("palomachain.paloma.gravity.EthereumSigned", (*EthereumSigned)(nil),
		&OutgoingTxBatch{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterCodec registers concrete types on the Amino codec
// nolint: exhaustruct
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*EthereumClaim)(nil), nil)
	cdc.RegisterConcrete(&MsgSendToEth{}, "gravity/MsgSendToEth", nil)
	cdc.RegisterConcrete(&MsgConfirmBatch{}, "gravity/MsgConfirmBatch", nil)
	cdc.RegisterConcrete(&MsgSendToPalomaClaim{}, "gravity/MsgSendToPalomaClaim", nil)
	cdc.RegisterConcrete(&MsgBatchSendToEthClaim{}, "gravity/MsgBatchSendToEthClaim", nil)
	cdc.RegisterConcrete(&OutgoingTxBatch{}, "gravity/OutgoingTxBatch", nil)
	cdc.RegisterConcrete(&MsgCancelSendToEth{}, "gravity/MsgCancelSendToEth", nil)
	cdc.RegisterConcrete(&OutgoingTransferTx{}, "gravity/OutgoingTransferTx", nil)
	cdc.RegisterConcrete(&ERC20Token{}, "gravity/ERC20Token", nil)
	cdc.RegisterConcrete(&IDSet{}, "gravity/IDSet", nil)
	cdc.RegisterConcrete(&Attestation{}, "gravity/Attestation", nil)
	cdc.RegisterConcrete(&MsgSubmitBadSignatureEvidence{}, "gravity/MsgSubmitBadSignatureEvidence", nil)
	cdc.RegisterConcrete(&SetERC20ToDenomProposal{}, "gravity/SetERC20ToDenomProposal", nil)
}
