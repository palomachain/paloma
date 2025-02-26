package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gogoproto/proto"
	consensustypes "github.com/palomachain/paloma/v2/x/consensus/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRemoveSmartContractDeploymentRequest{}, "evm/DeleteSmartContractDeployment", nil)
	cdc.RegisterConcrete(&RelayWeightsProposal{}, "evm/RelayWeightsProposal", nil)
	cdc.RegisterConcrete(&SetFeeManagerAddressProposal{}, "evm/SetFeeManagerAddressProposal", nil)
	cdc.RegisterConcrete(&SetSmartContractDeployersProposal{}, "evm/SetSmartContractDeployersProposal", nil)
	cdc.RegisterConcrete(&MsgDeployNewSmartContractProposalV2{}, "evm/ProposeNewSmartContractDeployment", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRemoveSmartContractDeploymentRequest{},
		&MsgDeployNewSmartContractProposalV2{},
	)
	registry.RegisterImplementations(
		(*govv1beta1types.Content)(nil),
		&AddChainProposal{},
		&RemoveChainProposal{},
		&DeployNewSmartContractProposal{},
		&ChangeMinOnChainBalanceProposal{},
		&RelayWeightsProposal{},
		&SetFeeManagerAddressProposal{},
		&SetSmartContractDeployersProposal{},
	)
	registry.RegisterImplementations(
		(*consensustypes.ConsensusMsg)(nil),
		&Message{},
		&ValidatorBalancesAttestation{},
		&ReferenceBlockAttestation{},
	)
	registry.RegisterImplementations(
		(*Hashable)(nil),
		&TxExecutedProof{},
		&SmartContractExecutionErrorProof{},
		&ValidatorBalancesAttestationRes{},
		&ReferenceBlockAttestationRes{},
	)
	registry.RegisterImplementations(
		(*TurnstoneMsg)(nil),
		&Message{},
		&ValidatorBalancesAttestation{},
		&CollectFunds{},
	)
	// any arbitrary message
	registry.RegisterImplementations(
		(*proto.Message)(nil),
		&ValidatorBalancesAttestationRes{},
		&ReferenceBlockAttestationRes{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
