package bindings

import (
	"context"
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errtypes "github.com/cosmos/cosmos-sdk/types/errors"
	bindingstypes "github.com/palomachain/paloma/v2/x/scheduler/bindings/types"
	schedulertypes "github.com/palomachain/paloma/v2/x/scheduler/types"
)

type Schedulerkeeper interface {
	GetJob(context.Context, string) (*schedulertypes.Job, error)
	ExecuteJob(context.Context, string, []byte, sdk.AccAddress, sdk.AccAddress) (uint64, error)
}

type QueryPlugin struct {
	k Schedulerkeeper
}

func NewQueryPlugin(k Schedulerkeeper) wasmkeeper.CustomQuerier {
	return customQuerier(&QueryPlugin{
		k: k,
	})
}

func customQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var q bindingstypes.Query
		if err := json.Unmarshal(request, &q); err != nil {
			return nil, sdkerrors.Wrap(err, "query")
		}

		switch {
		case q.JobById != nil:
			j, err := qp.k.GetJob(ctx, q.JobById.JobId)
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
