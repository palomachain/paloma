package bindings

import (
	"context"
	"encoding/hex"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/libwasm"
	bindingstypes "github.com/palomachain/paloma/v2/x/scheduler/bindings/types"
	schedulertypes "github.com/palomachain/paloma/v2/x/scheduler/types"
)

type SchedulerMsgServer interface {
	CreateJob(context.Context, *schedulertypes.MsgCreateJob) (*schedulertypes.MsgCreateJobResponse, error)
}

func NewMessenger(
	k Schedulerkeeper,
	ms SchedulerMsgServer,
) libwasm.Messenger[bindingstypes.Message] {
	return &customMessenger{
		ms: ms,
		k:  k,
	}
}

type customMessenger struct {
	ms SchedulerMsgServer
	k  Schedulerkeeper
}

var _ libwasm.Messenger[bindingstypes.Message] = (*customMessenger)(nil)

func (m *customMessenger) DispatchMsg(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	_ string,
	contractMsg bindingstypes.Message,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	switch {
	case contractMsg.CreateJob != nil:
		return m.createJob(ctx, contractAddr, contractMsg.CreateJob)
	case contractMsg.ExecuteJob != nil:
		return m.executeJob(ctx, contractAddr, contractMsg.ExecuteJob)
	}

	return nil, nil, nil, libwasm.ErrUnrecognizedMessage
}

func (m *customMessenger) createJob(ctx sdk.Context, contractAddr sdk.AccAddress, createJob *bindingstypes.CreateJob) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	if createJob == nil {
		return nil, nil, nil, wasmvmtypes.InvalidRequest{Err: "null create job"}
	}
	if createJob.Job == nil {
		return nil, nil, nil, wasmvmtypes.InvalidRequest{Err: "null job"}
	}

	j := &schedulertypes.Job{
		ID: createJob.Job.JobId,
		Routing: schedulertypes.Routing{
			ChainType:        createJob.Job.ChainType,
			ChainReferenceID: createJob.Job.ChainReferenceId,
		},
		Definition:          []byte(createJob.Job.Definition),
		Payload:             []byte(createJob.Job.Payload),
		IsPayloadModifiable: createJob.Job.PayloadModifiable,
		EnforceMEVRelay:     createJob.Job.IsMEV,
	}

	if err := j.ValidateBasic(); err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to validate job")
	}

	msgCreateJob := schedulertypes.NewMsgCreateJob(contractAddr.String(), j)
	if err := msgCreateJob.ValidateBasic(); err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed validating MsgCreateJob")
	}

	resp, err := m.ms.CreateJob(ctx, msgCreateJob)
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to create job")
	}

	bz, err := resp.Marshal()
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to marshal response")
	}
	return nil, [][]byte{bz}, nil, nil
}

func (m *customMessenger) executeJob(ctx sdk.Context, contractAddr sdk.AccAddress, e *bindingstypes.ExecuteJob) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	if len(e.JobID) == 0 {
		return nil, nil, nil, wasmvmtypes.InvalidRequest{Err: "invalid job id"}
	}

	if len(e.Payload) == 0 {
		return nil, nil, nil, wasmvmtypes.InvalidRequest{Err: "missing payload"}
	}

	hexString := hex.EncodeToString(e.Payload)
	injected := []byte(fmt.Sprintf("{\"hexPayload\":\"%s\"}", hexString))

	// We use the keeper method here instead of going via msg server,
	// as that will allow us to set the sender contract address.
	_, err := m.k.ExecuteJob(ctx, e.JobID, injected, contractAddr, contractAddr)
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to trigger job execution.")
	}

	return nil, nil, nil, nil
}
