package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
//
//go:generate mockery --name=AccountKeeper
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	SetAccount(ctx context.Context, acc sdk.AccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
//
//go:generate mockery --name=BankKeeper
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

// EvmKeeper defines the expected interface for interacting with the evm module
//
//go:generate mockery --name=EvmKeeper
type EvmKeeper interface {
	PreJobExecution(ctx context.Context, job *Job) error
	PickValidatorForMessage(ctx context.Context, chainReferenceID string, req *xchain.JobRequirements) (string, error)
}

// ValsetKeeper defines the expected interface for interacting with the valset module
//
//go:generate mockery --name=ValsetKeeper
type ValsetKeeper interface {
	GetCurrentSnapshot(ctx context.Context) (*valsettypes.Snapshot, error)
}
