package types

import (
	"context"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/v2/x/consensus/types"
	metrixtypes "github.com/palomachain/paloma/v2/x/metrix/types"
	valsettypes "github.com/palomachain/paloma/v2/x/valset/types"
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

type ConsensusKeeper interface {
	PutMessageInQueue(ctx context.Context, queueTypeName string, msg consensus.ConsensusMsg, opts *consensus.PutOptions) (uint64, error)
	RemoveConsensusQueue(ctx context.Context, queueTypeName string) error
	GetMessagesFromQueue(ctx context.Context, queueTypeName string, n int) (msgs []consensustypes.QueuedSignedMessageI, err error)
	DeleteJob(ctx context.Context, queueTypeName string, id uint64) (err error)
}

//go:generate mockery --name=ValsetKeeper
type ValsetKeeper interface {
	FindSnapshotByID(ctx context.Context, id uint64) (*valsettypes.Snapshot, error)
	GetCurrentSnapshot(ctx context.Context) (*valsettypes.Snapshot, error)
	SetSnapshotOnChain(ctx context.Context, snapshotID uint64, chainReferenceID string) error
	GetLatestSnapshotOnChain(ctx context.Context, chainReferenceID string) (*valsettypes.Snapshot, error)
	KeepValidatorAlive(ctx context.Context, valAddr sdk.ValAddress, pigeonVersion string) error
	Jail(ctx context.Context, valAddr sdk.ValAddress, reason string) error
	IsJailed(ctx context.Context, val sdk.ValAddress) (bool, error)
	SetValidatorBalance(ctx context.Context, valAddr sdk.ValAddress, chainType string, chainReferenceID string, externalAddress string, balance *big.Int) error
	GetValidatorChainInfos(ctx context.Context, valAddr sdk.ValAddress) ([]*valsettypes.ExternalChainInfo, error)
	GetAllChainInfos(ctx context.Context) ([]*valsettypes.ValidatorExternalAccounts, error)
}

//go:generate mockery --name=ConsensusKeeper
type SchedulerKeeper interface {
	ScheduleNow(ctx context.Context, jobID string, in []byte, SenderAddress sdk.AccAddress, contractAddress sdk.AccAddress) error
}

type ERC20Record interface {
	GetErc20() string
	GetDenom() string
	GetChainReferenceId() string
}

//go:generate mockery --name=SkywayKeeper
type SkywayKeeper interface {
	GetLastObservedSkywayNonce(ctx context.Context, chainReferenceID string) (uint64, error)
	CastAllERC20ToDenoms(ctx context.Context) ([]ERC20Record, error)
	CastChainERC20ToDenoms(ctx context.Context, chainReferenceID string) ([]ERC20Record, error)
}

//go:generate mockery --name=MetrixKeeper
type MetrixKeeper interface {
	Validators(goCtx context.Context, _ *metrixtypes.Empty) (*metrixtypes.QueryValidatorsResponse, error)
}

//go:generate mockery --name=TreasuryKeeper
type TreasuryKeeper interface {
	GetRelayerFeesByChainReferenceID(ctx context.Context, chainReferenceID string) (map[string]math.LegacyDec, error)
}
