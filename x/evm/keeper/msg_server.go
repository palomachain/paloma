package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/palomachain/paloma/v2/x/evm/types"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (k msgServer) ProposeNewSmartContractDeployment(ctx context.Context, req *types.MsgDeployNewSmartContractProposalV2) (*emptypb.Empty, error) {
	if err := req.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "message validation failed: %v", err)
	}
	if req.Authority != req.Metadata.Creator {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "creator mismatch; expected %s, got %s", k.authority, req.Metadata.Creator)
	}
	if req.Metadata.Creator != k.authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Metadata.Creator)
	}
	sc, err := k.SaveNewSmartContract(ctx, req.GetAbiJSON(), req.Bytecode())
	if err != nil {
		return nil, err
	}

	if err := k.SetAsCompassContract(ctx, sc); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
