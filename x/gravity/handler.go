package gravity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/palomachain/paloma/x/gravity/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
)

// NewHandler returns a handler for "Gravity" type messages.
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgSendToEthereum:
			res, err := msgServer.SendToEthereum(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgCancelSendToEthereum:
			res, err := msgServer.CancelSendToEthereum(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgSubmitEthereumTxConfirmation:
			res, err := msgServer.SubmitEthereumTxConfirmation(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgSubmitEthereumEvent:
			res, err := msgServer.SubmitEthereumEvent(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgDelegateKeys:
			res, err := msgServer.SetDelegateKeys(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgEthereumHeightVote:
			res, err := msgServer.SubmitEthereumHeightVote(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

func NewCommunityPoolEthereumSpendProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CommunityPoolEthereumSpendProposal:
			return k.HandleCommunityPoolEthereumSpendProposal(ctx, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized gravity proposal content type: %T", c)
		}
	}
}
