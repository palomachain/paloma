package gravity

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errorsmod "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/palomachain/paloma/util/common"
	"github.com/palomachain/paloma/x/gravity/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
)

// NewHandler returns a handler for "Gravity" type messages.
func NewHandler(k keeper.Keeper) baseapp.MsgServiceHandler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.MsgSendToEth:
			res, err := msgServer.SendToEth(common.SdkContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgConfirmBatch:
			res, err := msgServer.ConfirmBatch(common.SdkContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgSendToPalomaClaim:
			res, err := msgServer.SendToPalomaClaim(common.SdkContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgBatchSendToEthClaim:
			res, err := msgServer.BatchSendToEthClaim(common.SdkContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgCancelSendToEth:
			res, err := msgServer.CancelSendToEth(common.SdkContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgSubmitBadSignatureEvidence:
			res, err := msgServer.SubmitBadSignatureEvidence(common.SdkContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.Wrap(errorsmod.ErrUnknownRequest, fmt.Sprintf("Unrecognized Gravity Msg type: %v", sdk.MsgTypeURL(msg)))
		}
	}
}
