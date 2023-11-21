package types

import (
	govv1beta1types "cosmossdk.io/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&CommunityFundFeeProposal{}, "treasury/CommunityFundFeeProposal", nil)
	cdc.RegisterConcrete(&SecurityFeeProposal{}, "treasury/SecurityFeeProposal", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)

	registry.RegisterImplementations(
		(*govv1beta1types.Content)(nil),
		&CommunityFundFeeProposal{},
		&SecurityFeeProposal{},
	)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
