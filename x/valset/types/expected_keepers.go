package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

//go:generate mockery --name=StakingKeeper
//go:generate mockery --srcpkg=github.com/cosmos/cosmos-sdk/x/staking/types --name=ValidatorI --structname=StakingValidatorI
type StakingKeeper interface {
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
	IterateValidators(ctx sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	// Deprecated: use SlashingKeeper instead.
	Jail(ctx sdk.Context, consAddr sdk.ConsAddress)
}

//go:generate mockery --name=SlashingKeeper
type SlashingKeeper interface {
	Jail(sdk.Context, sdk.ConsAddress)
	JailUntil(sdk.Context, sdk.ConsAddress, time.Time)
}

// TODO: Move this to metrix module to reduce dependencies
type OnSnapshotBuiltListener interface {
	OnSnapshotBuilt(sdk.Context, *Snapshot)
}

type EvmKeeper interface {
	MissingChains(ctx sdk.Context, chainReferenceIDs []string) ([]string, error)
}
