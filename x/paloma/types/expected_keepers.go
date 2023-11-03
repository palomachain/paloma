package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

type ValsetKeeper interface {
	GetUnjailedValidators(ctx sdk.Context) []stakingtypes.ValidatorI
	Jail(ctx sdk.Context, valAddr sdk.ValAddress, reason string) error
	GetValidatorChainInfos(ctx sdk.Context, valAddr sdk.ValAddress) ([]*valsettypes.ExternalChainInfo, error)
}

type ExternalChainSupporterKeeper interface {
	xchain.Info
}

type UpgradeKeeper interface {
	GetLastCompletedUpgrade(ctx sdk.Context) (string, int64)
}

type FeegrantKeeper interface {
	AllowancesByGranter(ctx context.Context, req *feegrant.QueryAllowancesByGranterRequest) (*feegrant.QueryAllowancesByGranterResponse, error)
}
