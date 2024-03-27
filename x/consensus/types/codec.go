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
	cdc.RegisterConcrete(&MsgAddEvidence{}, "consensus/AddEvidence", nil)
	cdc.RegisterConcrete(&MsgSetPublicAccessData{}, "consensus/SetPublicAccessData", nil)
	cdc.RegisterConcrete(&MsgSetErrorData{}, "consensus/SetErrorData", nil)
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
		&Batch{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddEvidence{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetPublicAccessData{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
