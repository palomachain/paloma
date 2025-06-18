package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateParams{}, "valset/UpdateParams", nil)
	cdc.RegisterConcrete(&MsgAddExternalChainInfoForValidator{}, "valset/AddExternalChainInfoForValidator", nil)
	cdc.RegisterConcrete(&MsgKeepAlive{}, "valset/KeepAlive", nil)
	cdc.RegisterConcrete(&SetPigeonRequirementsProposal{}, "valset/SetPigeonRequirementsProposal", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgAddExternalChainInfoForValidator{},
		&MsgKeepAlive{},
	)
	registry.RegisterImplementations(
		(*govv1beta1types.Content)(nil),
		&SetPigeonRequirementsProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

//- Jail validators who fail to attest bridge claim
//- Stop sending new batches while waiting for claim confirmation
