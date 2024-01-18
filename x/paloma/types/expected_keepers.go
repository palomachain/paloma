package types

import (
	context "context"

	"cosmossdk.io/x/feegrant"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

type ValsetKeeper interface {
	GetUnjailedValidators(ctx context.Context) []stakingtypes.ValidatorI
	Jail(ctx context.Context, valAddr sdk.ValAddress, reason string) error
	GetValidatorChainInfos(ctx context.Context, valAddr sdk.ValAddress) ([]*valsettypes.ExternalChainInfo, error)
}

type ExternalChainSupporterKeeper interface {
	xchain.Info
}

type UpgradeKeeper interface {
	GetLastCompletedUpgrade(ctx context.Context) (string, int64, error)
}

type FeegrantKeeper interface {
	AllowancesByGranter(ctx context.Context, req *feegrant.QueryAllowancesByGranterRequest) (*feegrant.QueryAllowancesByGranterResponse, error)
}
