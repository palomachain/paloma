package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

//go:generate mockery --name=SlashingKeeper
type SlashingKeeper interface {
	SignedBlocksWindow(sdk.Context) int64
	IterateValidatorSigningInfos(sdk.Context, func(sdk.ConsAddress, slashingtypes.ValidatorSigningInfo) (stop bool))
}

//go:generate mockery --name=StakingKeeper
type StakingKeeper interface {
	GetValidator(sdk.Context, sdk.ValAddress) (stakingtypes.Validator, bool)
	GetValidatorByConsAddr(sdk.Context, sdk.ConsAddress) (stakingtypes.Validator, bool)
}
