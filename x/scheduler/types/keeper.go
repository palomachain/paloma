package types

import (
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

//go:generate mockery --name=Keeper
type Keeper interface {
	AddNewJob(ctx sdk.Context, job *Job) error
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetJob(ctx sdk.Context, jobID string) (*Job, error)
	Logger(ctx sdk.Context) log.Logger
	PreJobExecution(ctx sdk.Context, job *Job) error
	ScheduleNow(ctx sdk.Context, jobID string, in []byte, senderAddress sdk.AccAddress, contractAddress sdk.AccAddress) (uint64, error)
	ExecuteJob(ctx sdk.Context, jobID string, payload []byte, senderAddress sdk.AccAddress, contractAddr sdk.AccAddress) (uint64, error)
}
