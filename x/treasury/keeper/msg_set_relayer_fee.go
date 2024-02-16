package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/treasury/types"
)

// SetRelayerFee implements types.MsgServer.
func (m msgServer) SetRelayerFee(ctx context.Context, req *types.MsgSetRelayerFee) (*types.SetRelayerFeeResponse, error) {
	a, err := sdk.ValAddressFromBech32(req.Fee.Validator)
	if err != nil {
		return nil, err
	}

	if err := m.setRelayerFee(ctx, a, &req.Fee); err != nil {
		return nil, err
	}

	return &types.SetRelayerFeeResponse{
		Fee: req.Fee,
	}, nil
}
