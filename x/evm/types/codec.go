package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
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
	)
	registry.RegisterImplementations((*consensustypes.ConsensusMsg)(nil),
		&ArbitrarySmartContractCall{},
		&Message{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
	&MsgUploadNewSmartContractTemp{},
)
// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
