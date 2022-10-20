package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
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

//go:generate mockery --name=ConsensusKeeper
type ConsensusKeeper interface {
	PutMessageInQueue(ctx sdk.Context, queueTypeName string, msg consensus.ConsensusMsg, opts *consensus.PutOptions) error
	RemoveConsensusQueue(ctx sdk.Context, queueTypeName string) error
	GetMessagesFromQueue(ctx sdk.Context, queueTypeName string, n int) (msgs []consensustypes.QueuedSignedMessageI, err error)
	DeleteJob(ctx sdk.Context, queueTypeName string, id uint64) (err error)
}

//go:generate mockery --name=ValsetKeeper
type ValsetKeeper interface {
	FindSnapshotByID(ctx sdk.Context, id uint64) (*valsettypes.Snapshot, error)
	GetCurrentSnapshot(ctx sdk.Context) (*valsettypes.Snapshot, error)
	KeepValidatorAlive(ctx sdk.Context, valAddr sdk.ValAddress) error
	Jail(ctx sdk.Context, valAddr sdk.ValAddress, reason string) error
	IsJailed(ctx sdk.Context, val sdk.ValAddress) bool
	SetValidatorBalance(ctx sdk.Context, valAddr sdk.ValAddress, chainType string, chainReferenceID string, externalAddress string, balance *big.Int) error
}

type JobScheduler interface {
	ScheduleNow(ctx sdk.Context, jobID string, in []byte) error
}
