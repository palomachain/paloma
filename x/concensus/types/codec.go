package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&QueuedSignedMessage{}, "concensus/QueuedSignedMessage", nil)
	cdc.RegisterConcrete(&Signer{}, "concensus/Signer", nil)
	cdc.RegisterConcrete(&MsgAddMessagesSignatures{}, "concensus/AddMessagesSignatures", nil)
// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*QueuedSignedMessageI)(nil),
		&QueuedSignedMessage{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
	&MsgAddMessagesSignatures{},
)
// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
