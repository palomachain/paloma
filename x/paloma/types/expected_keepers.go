package types

import (
	context "context"

	"cosmossdk.io/core/address"
	"cosmossdk.io/x/feegrant"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
//
//go:generate mockery --name=AccountKeeper
type AccountKeeper interface {
	AddressCodec() address.Codec
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
//
//go:generate mockery --name=BankKeeper
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	HasBalance(ctx context.Context, addr sdk.AccAddress, amt sdk.Coin) bool
}

//go:generate mockery --name=ValsetKeeper
type ValsetKeeper interface {
	GetUnjailedValidators(ctx context.Context) []stakingtypes.ValidatorI
	Jail(ctx context.Context, valAddr sdk.ValAddress, reason string) error
	GetValidatorChainInfos(ctx context.Context, valAddr sdk.ValAddress) ([]*valsettypes.ExternalChainInfo, error)
}

type ExternalChainSupporterKeeper interface {
	xchain.Info
}

//go:generate mockery --name=UpgradeKeeper
type UpgradeKeeper interface {
	GetLastCompletedUpgrade(ctx context.Context) (string, int64, error)
}

//go:generate mockery --name=FeegrantKeeper
type FeegrantKeeper interface {
	AllowancesByGranter(ctx context.Context, req *feegrant.QueryAllowancesByGranterRequest) (*feegrant.QueryAllowancesByGranterResponse, error)
	GrantAllowance(ctx context.Context, granter, grantee sdk.AccAddress, feeAllowance feegrant.FeeAllowanceI) error
}
