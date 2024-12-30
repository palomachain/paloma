package bindings

import (
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errtypes "github.com/cosmos/cosmos-sdk/types/errors"
	bindingstypes "github.com/palomachain/paloma/v2/x/scheduler/bindings/types"
	schedulerkeeper "github.com/palomachain/paloma/v2/x/scheduler/keeper"
	schedulertypes "github.com/palomachain/paloma/v2/x/scheduler/types"
)

func CustomMessageDecorator(scheduler *schedulerkeeper.Keeper) func(wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &CustomMessenger{
			wrapped:   old,
			scheduler: scheduler,
		}
	}
}

type CustomMessenger struct {
	wrapped   wasmkeeper.Messenger
	scheduler *schedulerkeeper.Keeper
}

var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

func (m *CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	if msg.Custom != nil {
		var contractMsg bindingstypes.SchedulerMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, nil, sdkerrors.Wrap(err, "scheduler msg")
		}
		if contractMsg.Message == nil {
			return nil, nil, nil, sdkerrors.Wrap(errtypes.ErrUnknownRequest, "nil message field")
		}
		msgType := contractMsg.Message
		if msgType.CreateJob != nil {
			return m.createJob(ctx, contractAddr, msgType.CreateJob)
		}
	}
	return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

func (m *CustomMessenger) createJob(ctx sdk.Context, contractAddr sdk.AccAddress, createJob *bindingstypes.CreateJob) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
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

	msgServer := schedulerkeeper.NewMsgServerImpl(m.scheduler)
	msgCreateJob := schedulertypes.NewMsgCreateJob(contractAddr.String(), j)
	if err := msgCreateJob.ValidateBasic(); err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed validating MsgCreateJob")
	}

	resp, err := msgServer.CreateJob(ctx, msgCreateJob)
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to create job")
	}

	bz, err := resp.Marshal()
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to marshal response")
	}
	return nil, [][]byte{bz}, nil, nil
}
