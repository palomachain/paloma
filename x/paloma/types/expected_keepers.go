package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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
	UnjailedValidators(ctx sdk.Context) []stakingtypes.ValidatorI
	Jail(ctx sdk.Context, valAddr sdk.ValAddress, reason string) error
	GetValidatorChainInfos(ctx sdk.Context, valAddr sdk.ValAddress) ([]*valsettypes.ExternalChainInfo, error)
}

type ExternalChainSupporterKeeper interface {
	ChainType(sdk.Context) string
	ChainReferenceIDs(sdk.Context) []string
}
