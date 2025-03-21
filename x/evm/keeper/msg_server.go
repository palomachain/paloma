package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/palomachain/paloma/v2/x/evm/types"
	vtypes "github.com/palomachain/paloma/v2/x/valset/types"
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
	if err := governanceMsgGuard(k.authority, req); err != nil {
		return nil, err
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

func (k msgServer) ProposeNewReferenceBlockAttestation(ctx context.Context, req *types.MsgProposeNewReferenceBlockAttestation) (*emptypb.Empty, error) {
	err := governanceMsgGuard(k.authority, req)
	if err != nil {
		return nil, err
	}
	err = k.UpdateChainReferenceBlock(ctx, req.ChainReferenceId, req.BlockHeight, req.BlockHash)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

type governanceMsg interface {
	GetAuthority() string
	GetMetadata() vtypes.MsgMetadata
	ValidateBasic() error
}

func governanceMsgGuard(authority string, req governanceMsg) error {
	if err := req.ValidateBasic(); err != nil {
		return sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "message validation failed: %v", err)
	}
	if req.GetAuthority() != req.GetMetadata().Creator {
		return sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "creator mismatch; expected %s, got %s", authority, req.GetMetadata().Creator)
	}
	if req.GetMetadata().Creator != authority {
		return sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", authority, req.GetMetadata().Creator)
	}

	return nil
}
