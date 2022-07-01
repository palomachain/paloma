package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitNewJob{}, "evm/SubmitNewJob", nil)
	cdc.RegisterConcrete(&MsgUploadNewSmartContractTemp{}, "evm/UploadNewSmartContractTemp", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitNewJob{},
		&MsgUploadNewSmartContractTemp{},
	)
	registry.RegisterImplementations((*gov.Content)(nil),
		&AddChainProposal{},
		&RemoveChainProposal{},
		&DeployNewSmartContractProposal{},
	)
	registry.RegisterImplementations((*consensustypes.ConsensusMsg)(nil),
		&Message{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
