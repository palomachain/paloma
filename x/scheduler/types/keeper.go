package types

import (
	"context"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

//go:generate mockery --name=Keeper
type Keeper interface {
	AddNewJob(ctx context.Context, job *Job) error
	GetAccount(ctx context.Context, addr sdk.AccAddress) types.AccountI
	GetJob(ctx context.Context, jobID string) (*Job, error)
	Logger(ctx context.Context) log.Logger
	PreJobExecution(ctx context.Context, job *Job) error
	ScheduleNow(ctx context.Context, jobID string, in []byte, senderAddress sdk.AccAddress, contractAddress sdk.AccAddress) (uint64, error)
	ExecuteJob(ctx context.Context, jobID string, payload []byte, senderAddress sdk.AccAddress, contractAddr sdk.AccAddress) (uint64, error)
}
