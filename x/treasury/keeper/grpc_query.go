package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/treasury/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RelayerFee(ctx context.Context, req *types.QueryRelayerFeeRequest) (*types.RelayerFeeSetting, error) {
	addr, err := sdk.ValAddressFromBech32(req.ValAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse validator address: %w", err)
	}

	return k.relayerFees.Get(sdk.UnwrapSDKContext(ctx), addr)
}
