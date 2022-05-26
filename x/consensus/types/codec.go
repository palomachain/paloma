package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&QueuedSignedMessage{}, "consensus/QueuedSignedMessage", nil)
	cdc.RegisterConcrete(&SignData{}, "consensus/SignData", nil)
	cdc.RegisterConcrete(&MsgAddMessagesSignatures{}, "consensus/AddMessagesSignatures", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*QueuedSignedMessageI)(nil),
		&QueuedSignedMessage{},
	)
	registry.RegisterImplementations((*MessageQueuedForBatchingI)(nil),
		&BatchOfConsensusMessages{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddMessagesSignatures{},
	)
	registry.RegisterImplementations((*ConsensusMsg)(nil),
		&SignSmartContractExecute{},
		&Batch{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
