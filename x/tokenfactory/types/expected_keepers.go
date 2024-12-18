package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

//go:generate mockery --name=BankKeeper
type BankKeeper interface {
	// Methods imported from bank should be defined here
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)

	HasSupply(ctx context.Context, denom string) bool

	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error

	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	HasBalance(ctx context.Context, addr sdk.AccAddress, amt sdk.Coin) bool
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

//go:generate mockery --name=AccountKeeper
type AccountKeeper interface {
	SetModuleAccount(ctx context.Context, macc sdktypes.ModuleAccountI)
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdktypes.AccountI
}

//go:generate mockery --name=CommunityPoolKeeper
type CommunityPoolKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}
