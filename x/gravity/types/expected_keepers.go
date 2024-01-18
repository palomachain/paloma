package types

import (
	context "context"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	evmtypes "github.com/palomachain/paloma/x/evm/types"
)

// StakingKeeper defines the expected staking keeper methods
type StakingKeeper interface {
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
	GetLastValidatorPower(ctx context.Context, operator sdk.ValAddress) (int64, error)
	GetLastTotalPower(ctx context.Context) (math.Int, error)
	IterateValidators(context.Context, func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error
	ValidatorQueueIterator(ctx context.Context, endTime time.Time, endHeight int64) (storetypes.Iterator, error)
	GetParams(ctx context.Context) (stakingtypes.Params, error)
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found error)
	IterateBondedValidatorsByPower(context.Context, func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error
	IterateLastValidators(context.Context, func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error
	Validator(context.Context, sdk.ValAddress) (stakingtypes.ValidatorI, error)
	ValidatorByConsAddr(context.Context, sdk.ConsAddress) (stakingtypes.ValidatorI, error)
	Slash(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor math.LegacyDec) (math.Int, error)
	Jail(context.Context, sdk.ConsAddress) error
}

// BankKeeper defines the expected bank keeper methods
type BankKeeper interface {
	GetSupply(ctx context.Context, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule string, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, name string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetDenomMetaData(ctx context.Context, denom string) (bank.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData bank.Metadata)
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error
}

type SlashingKeeper interface {
	GetValidatorSigningInfo(ctx context.Context, address sdk.ConsAddress) (info slashingtypes.ValidatorSigningInfo, found error)
}

// AccountKeeper defines the interface contract required for account
// functionality.
type AccountKeeper interface {
	GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error)
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

type DistributionKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

type EVMKeeper interface {
	GetChainInfo(ctx context.Context, targetChainReferenceID string) (*evmtypes.ChainInfo, error)
	PickValidatorForMessage(ctx context.Context, chainReferenceID string, requirements *xchain.JobRequirements) (string, error)
	GetEthAddressByValidator(ctx context.Context, validator sdk.ValAddress, chainReferenceId string) (ethAddress *EthAddress, found bool, err error)
	GetValidatorAddressByEthAddress(ctx context.Context, ethAddr EthAddress, chainReferenceId string) (valAddr sdk.ValAddress, found bool, err error)
}
