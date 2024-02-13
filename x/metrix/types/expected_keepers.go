package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

//go:generate mockery --name=SlashingKeeper
type SlashingKeeper interface {
	SignedBlocksWindow(context.Context) (int64, error)
	IterateValidatorSigningInfos(context.Context, func(sdk.ConsAddress, slashingtypes.ValidatorSigningInfo) (stop bool)) error
}

//go:generate mockery --name=StakingKeeper
type StakingKeeper interface {
	GetValidator(context.Context, sdk.ValAddress) (stakingtypes.Validator, error)
	GetValidatorByConsAddr(context.Context, sdk.ConsAddress) (stakingtypes.Validator, error)
	IterateValidators(context.Context, func(int64, stakingtypes.ValidatorI) bool) error
}
