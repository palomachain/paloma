package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	evmtypes "github.com/palomachain/paloma/x/evm/types"
)

// StakingKeeper defines the expected staking keeper methods
type StakingKeeper interface {
	GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator
	GetLastValidatorPower(ctx sdk.Context, operator sdk.ValAddress) int64
	GetLastTotalPower(ctx sdk.Context) (power sdk.Int)
	IterateValidators(sdk.Context, func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	ValidatorQueueIterator(ctx sdk.Context, endTime time.Time, endHeight int64) sdk.Iterator
	GetParams(ctx sdk.Context) stakingtypes.Params
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	IterateBondedValidatorsByPower(sdk.Context, func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	IterateLastValidators(sdk.Context, func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	Validator(sdk.Context, sdk.ValAddress) stakingtypes.ValidatorI
	ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) stakingtypes.ValidatorI
	Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor sdk.Dec) math.Int
	Jail(sdk.Context, sdk.ConsAddress)
}

// BankKeeper defines the expected bank keeper methods
type BankKeeper interface {
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetDenomMetaData(ctx sdk.Context, denom string) (bank.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData bank.Metadata)
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error
	SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error
}

type SlashingKeeper interface {
	GetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (info slashingtypes.ValidatorSigningInfo, found bool)
}

// AccountKeeper defines the interface contract required for account
// functionality.
type AccountKeeper interface {
	GetSequence(ctx sdk.Context, addr sdk.AccAddress) (uint64, error)
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
}

type DistributionKeeper interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
	GetFeePool(ctx sdk.Context) (feePool distributiontypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool distributiontypes.FeePool)
}

type EVMKeeper interface {
	GetChainInfo(ctx sdk.Context, targetChainReferenceID string) (*evmtypes.ChainInfo, error)
	PickValidatorForMessage(ctx sdk.Context, chainReferenceID string, requirements *xchain.JobRequirements) (string, error)
	GetEthAddressByValidator(ctx sdk.Context, validator sdk.ValAddress, chainReferenceId string) (ethAddress *EthAddress, found bool, err error)
	GetValidatorAddressByEthAddress(ctx sdk.Context, ethAddr EthAddress, chainReferenceId string) (valAddr sdk.ValAddress, found bool, err error)
}
