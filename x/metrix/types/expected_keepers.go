package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

//go:generate mockery --name=SlashingKeeper
type SlashingKeeper interface {
	SignedBlocksWindow(sdk.Context) int64
	IterateValidatorSigningInfos(sdk.Context, func(sdk.ConsAddress, slashingtypes.ValidatorSigningInfo) (stop bool))
}
