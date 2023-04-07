package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	proto "github.com/cosmos/gogoproto/proto"

	consensustypes "github.com/palomachain/paloma/x/consensus/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitNewJob{}, "evm/SubmitNewJob", nil)
	cdc.RegisterConcrete(&MsgUploadNewSmartContractTemp{}, "evm/UploadNewSmartContractTemp", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSubmitNewJob{},
		&MsgUploadNewSmartContractTemp{},
	)
	registry.RegisterImplementations(
		(*govv1beta1types.Content)(nil),
		&AddChainProposal{},
		&RemoveChainProposal{},
		&DeployNewSmartContractProposal{},
		&ChangeMinOnChainBalanceProposal{},
	)
	registry.RegisterImplementations(
		(*consensustypes.ConsensusMsg)(nil),
		&Message{},
		&ValidatorBalancesAttestation{},
	)
	registry.RegisterImplementations(
		(*Hashable)(nil),
		&TxExecutedProof{},
		&SmartContractExecutionErrorProof{},
		&ValidatorBalancesAttestationRes{},
	)
	// any arbitrary message
	registry.RegisterImplementations(
		(*proto.Message)(nil),
		&ValidatorBalancesAttestationRes{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
