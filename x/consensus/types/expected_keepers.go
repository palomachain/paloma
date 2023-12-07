package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
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

//go:generate mockery --name=ValsetKeeper
type ValsetKeeper interface {
	GetSigningKey(ctx sdk.Context, valAddr sdk.ValAddress, chainType, chainReferenceID, signedByAddress string) ([]byte, error)
	GetCurrentSnapshot(ctx sdk.Context) (*valsettypes.Snapshot, error)
	CanAcceptValidator(ctx sdk.Context, valAddr sdk.ValAddress) error
	KeepValidatorAlive(ctx sdk.Context, valAddr sdk.ValAddress, pigeonVersion string) error
}

//go:generate mockery --name=EvmKeeper
type EvmKeeper interface {
	PickValidatorForMessage(ctx sdk.Context, chainReferenceID string, requirements *xchain.JobRequirements) (string, error)
}
