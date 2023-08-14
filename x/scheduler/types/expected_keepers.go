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
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	HasAccount(ctx sdk.Context, addr sdk.AccAddress) bool
	SetAccount(ctx sdk.Context, acc types.AccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

// EvmKeeper defines the expected interface for interacting with the evm module
type EvmKeeper interface {
	PreJobExecution(ctx sdk.Context, job *Job) error
	PickValidatorForMessage(ctx sdk.Context, chainReferenceID string, req *xchain.JobRequirements) (string, error)
}

// ValsetKeeper defines the expected interface for interacting with the valset module
type ValsetKeeper interface {
	GetCurrentSnapshot(ctx sdk.Context) (*valsettypes.Snapshot, error)
}
