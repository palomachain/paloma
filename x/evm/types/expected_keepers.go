package types

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
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

//go:generate mockery --name=GravityKeeper
type GravityKeeper interface {
	GetLastObservedEventNonce(ctx context.Context) (uint64, error)
	CastAllERC20ToDenoms(ctx context.Context) ([]ERC20Record, error)
}
