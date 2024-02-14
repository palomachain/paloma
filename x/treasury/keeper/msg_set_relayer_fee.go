package keeper

import (
	"context"

	"github.com/palomachain/paloma/x/treasury/types"
)

// SetRelayerFee implements types.MsgServer.
func (msgServer) SetRelayerFee(context.Context, *types.MsgSetRelayerFee) (*types.SetRelayerFeeResponse, error) {
	panic("unimplemented")
}
