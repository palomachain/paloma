package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/treasury/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) UpsertRelayerFee(ctx context.Context, req *types.MsgUpsertRelayerFee) (*types.Empty, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	addr, err := sdk.ValAddressFromBech32(req.FeeSetting.ValAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse validator address: %w", err)
	}

	// Setting of relayer fees for unknown chains is allowed
	// to support validators setting up before a new chain is onboarded

	r, err := k.relayerFees.Get(sdkCtx, addr)
	if err != nil {
		if !errors.Is(err, keeperutil.ErrNotFound) {
			return nil, err
		}
		r = &types.RelayerFeeSetting{
			ValAddress: addr.String(),
			Fees:       make([]types.RelayerFeeSetting_FeeSetting, 0, len(req.FeeSetting.Fees)),
		}
	}

	merged := &types.RelayerFeeSetting{
		ValAddress: r.ValAddress,
		Fees:       make([]types.RelayerFeeSetting_FeeSetting, 0, len(r.Fees)+len(req.FeeSetting.Fees)),
	}

	lkup := make(map[string]struct{})
	for _, v := range req.FeeSetting.Fees {
		merged.Fees = append(merged.Fees, v)
		lkup[v.ChainReferenceId] = struct{}{}
	}

	for _, v := range r.Fees {
		if _, fn := lkup[v.ChainReferenceId]; fn {
			// Don't override existing entries with stale data
			continue
		}

		merged.Fees = append(merged.Fees, v)
	}

	return &types.Empty{}, k.SetRelayerFee(ctx, addr, merged)
}
