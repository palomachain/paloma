package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/evm/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetValsetByID returns the valset given chain id and valset id. if the valset
// id is non-pozitive then it returns the latest valset existing.
func (k Keeper) GetValsetByID(goCtx context.Context, req *types.QueryGetValsetByIDRequest) (*types.QueryGetValsetByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var snapshot *valsettypes.Snapshot
	var err error

	if req.GetValsetID() > 0 {
		snapshot, err = k.Valset.FindSnapshotByID(ctx, req.GetValsetID())
	} else {
		snapshot, err = k.Valset.GetCurrentSnapshot(ctx)
	}

	if err != nil {
		return nil, err
	}

	valset := transformSnapshotToTurnstoneValset(snapshot, req.GetChainID())

	return &types.QueryGetValsetByIDResponse{
		Valset: &valset,
	}, nil
}
