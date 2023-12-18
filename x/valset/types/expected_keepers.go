package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)

type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.

type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

//go:generate mockery --name=StakingKeeper
//go:generate mockery --srcpkg=github.com/cosmos/cosmos-sdk/x/staking/types --name=ValidatorI --structname=StakingValidatorI
type StakingKeeper interface {
	Validator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.ValidatorI, error)
	IterateValidators(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) bool) error
	Jail(ctx context.Context, consAddr sdk.ConsAddress) error
}

type OnSnapshotBuiltListener interface {
	OnSnapshotBuilt(context.Context, *Snapshot)
}

type EvmKeeper interface {
	MissingChains(ctx context.Context, chainReferenceIDs []string) ([]string, error)
}
