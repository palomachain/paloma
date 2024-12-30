package bindings

import (
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errtypes "github.com/cosmos/cosmos-sdk/types/errors"
	bindingstypes "github.com/palomachain/paloma/v2/x/scheduler/bindings/types"
	schedulerkeeper "github.com/palomachain/paloma/v2/x/scheduler/keeper"
)

type QueryPlugin struct {
	scheduler *schedulerkeeper.Keeper
}

func NewQueryPlugin(s *schedulerkeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		scheduler: s,
	}
}

func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindingstypes.SchedulerQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "query")
		}
		if contractQuery.Query == nil {
			return nil, sdkerrors.Wrap(errtypes.ErrUnknownRequest, "nil query field")
		}
		queryType := contractQuery.Query

		switch {
		case queryType.JobById != nil:
			j, err := qp.scheduler.GetJob(ctx, queryType.JobById.JobId)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "failed to query for job")
			}
			if j == nil {
				return nil, sdkerrors.Wrap(errtypes.ErrNotFound, "job id")
			}

			res := bindingstypes.JobByIdResponse{
				Job: &bindingstypes.Job{
					JobId:             j.ID,
					ChainType:         j.Routing.ChainType,
					ChainReferenceId:  j.Routing.ChainReferenceID,
					Definition:        string(j.Definition),
					Payload:           string(j.Payload),
					PayloadModifiable: j.IsPayloadModifiable,
					IsMEV:             j.EnforceMEVRelay,
				},
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "failed to marshal response")
			}

			return bz, nil
		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown scheduler query variant"}
		}
	}
}
