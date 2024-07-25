package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAddStatusUpdate{}, "paloma/AddStatusUpdate", nil)
	cdc.RegisterConcrete(&MsgRegisterLightNodeClient{}, "paloma/RegisterLightNodeClient", nil)
	cdc.RegisterConcrete(&MsgAddLightNodeClientLicense{}, "paloma/AddLightNodeClientLicense", nil)
	cdc.RegisterConcrete(&SetLightNodeClientFeegranterProposal{}, "paloma/SetLightNodeClientFeegranterProposal", nil)
	cdc.RegisterConcrete(&SetLightNodeClientFundersProposal{}, "paloma/SetLightNodeClientFundersProposal", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddStatusUpdate{},
		&MsgRegisterLightNodeClient{},
		&MsgAddLightNodeClientLicense{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&SetLightNodeClientFeegranterProposal{},
		&SetLightNodeClientFundersProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
