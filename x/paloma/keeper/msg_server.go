package keeper

import (
	"context"
	"fmt"
	"os"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrtypes "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/palomachain/paloma/v2/util/liblog"
	"github.com/palomachain/paloma/v2/x/paloma/types"
)

const cPigeonStatusUpdateFF = "PALOMA_FF_PIGEON_STATUS_UPDATE"

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) AddStatusUpdate(goCtx context.Context, msg *types.MsgAddStatusUpdate) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Avoid log spamming
	_, ok := os.LookupEnv(cPigeonStatusUpdateFF)
	if !ok {
		return &types.EmptyResponse{}, nil
	}

	creator, err := sdk.AccAddressFromBech32(msg.GetMetadata().GetCreator())
	if err != nil {
		err = fmt.Errorf("failed to parse creator address: %w", err)
		liblog.FromSDKLogger(k.Logger(ctx)).WithError(err).WithFields("component", "pigeon-status-update").Error("Failed to parse creator from message")
		return nil, err
	}
	valAddr := sdk.ValAddress(creator.Bytes())
	status := msg.GetStatus()

	var logFn func(string, ...interface{})
	switch msg.Level {
	case types.MsgAddStatusUpdate_LEVEL_DEBUG:
		logFn = liblog.FromSDKLogger(k.Logger(ctx)).Debug
	case types.MsgAddStatusUpdate_LEVEL_INFO:
		logFn = liblog.FromSDKLogger(k.Logger(ctx)).Info
	case types.MsgAddStatusUpdate_LEVEL_ERROR:
		logFn = liblog.FromSDKLogger(k.Logger(ctx)).Error
	}

	args := make([]interface{}, 0, len(msg.GetArgs())*2+6) // len: (k+v) + 6 for static appends
	for _, v := range msg.GetArgs() {
		args = append(args, fmt.Sprintf("msg.args.%s", v.GetKey()), v.GetValue())
	}

	args = append(args,
		"component", "pigeon-status-update",
		"status", status,
		"sender", valAddr)

	logFn(status, args...)

	return &types.EmptyResponse{}, nil
}

func (k msgServer) RegisterLightNodeClient(
	ctx context.Context,
	msg *types.MsgRegisterLightNodeClient,
) (*types.EmptyResponse, error) {
	err := k.CreateLightNodeClientAccount(ctx, msg.Metadata.Creator)
	return nil, err
}

func (k msgServer) AddLightNodeClientLicense(
	ctx context.Context,
	msg *types.MsgAddLightNodeClientLicense,
) (*types.EmptyResponse, error) {
	err := k.CreateLightNodeClientLicense(ctx, msg.Metadata.Creator,
		msg.ClientAddress, msg.Amount, msg.VestingMonths)

	return nil, err
}

func (k msgServer) AuthLightNodeClient(
	ctx context.Context,
	msg *types.MsgAuthLightNodeClient,
) (*types.EmptyResponse, error) {
	client, err := k.GetLightNodeClient(ctx, msg.Metadata.Creator)
	if err != nil {
		return nil, err
	}

	client.LastAuthAt = sdk.UnwrapSDKContext(ctx).BlockTime()
	err = k.SetLightNodeClient(ctx, client.ClientAddress, client)

	return nil, err
}

func (k msgServer) SetLegacyLightNodeClients(
	ctx context.Context,
	_ *types.MsgSetLegacyLightNodeClients,
) (*types.EmptyResponse, error) {
	clients, err := k.GetLegacyLightNodeClients(ctx)
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		err = k.SetLightNodeClient(ctx, client.ClientAddress, client)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != msg.Authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrapf(sdkerrtypes.ErrInvalidRequest, "failed to validate message: %s", err.Error())
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, sdkerrors.Wrapf(sdkerrtypes.ErrInvalidRequest, "failed to validate params: %s", err.Error())
	}

	k.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
