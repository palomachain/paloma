package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
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

//go:generate mockery --name=ValsetKeeper
type ValsetKeeper interface {
	GetSigningKey(ctx context.Context, valAddr sdk.ValAddress, chainType, chainReferenceID, signedByAddress string) ([]byte, error)
	GetCurrentSnapshot(ctx context.Context) (*valsettypes.Snapshot, error)
	CanAcceptValidator(ctx context.Context, valAddr sdk.ValAddress) error
	KeepValidatorAlive(ctx context.Context, valAddr sdk.ValAddress, pigeonVersion string) error
	Jail(ctx context.Context, valAddr sdk.ValAddress, reason string) error
}

//go:generate mockery --name=EvmKeeper
type EvmKeeper interface {
	PickValidatorForMessage(ctx context.Context, chainReferenceID string, requirements *xchain.JobRequirements) (string, error)
}
