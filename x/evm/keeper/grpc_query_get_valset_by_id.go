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

	logger := k.Logger(ctx)
	logger.Info("request info",
		"chain-reference-id", req.GetChainReferenceID(),
		"valset-id", req.GetValsetID(),
	)

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
	valset := transformSnapshotToCompass(snapshot, req.GetChainReferenceID(), logger)
	logger.Info("returning valset info",
		"valset-id", valset.ValsetID,
		"valset-validator-size", len(valset.Validators),
		"valset-power-size", len(valset.Powers),
	)

	return &types.QueryGetValsetByIDResponse{
		Valset: &valset,
	}, nil
}
