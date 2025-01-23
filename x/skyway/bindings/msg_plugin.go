package bindings

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/libwasm"
	bindingstypes "github.com/palomachain/paloma/v2/x/skyway/bindings/types"
	skywaytypes "github.com/palomachain/paloma/v2/x/skyway/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SkywayMsgServer interface {
	SetERC20ToTokenDenom(context.Context, *skywaytypes.MsgSetERC20ToTokenDenom) (*emptypb.Empty, error)
	SendToRemote(context.Context, *skywaytypes.MsgSendToRemote) (*skywaytypes.MsgSendToRemoteResponse, error)
	CancelSendToRemote(context.Context, *skywaytypes.MsgCancelSendToRemote) (*skywaytypes.MsgCancelSendToRemoteResponse, error)
}

type customMessenger struct {
	k SkywayMsgServer
}

var _ libwasm.Messenger[bindingstypes.Message] = (*customMessenger)(nil)

func NewMessenger(k SkywayMsgServer) libwasm.Messenger[bindingstypes.Message] {
	return &customMessenger{
		k: k,
	}
}

func (m *customMessenger) DispatchMsg(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	contractIBCPortID string,
	contractMsg bindingstypes.Message,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	switch {
	case contractMsg.SetErc20ToDenom != nil:
		return handleSetErc20ToDenom(ctx, m.k, contractAddr, contractMsg.SetErc20ToDenom)
	case contractMsg.SendTx != nil:
		return sendTx(ctx, m.k, contractAddr, contractMsg.SendTx)
	case contractMsg.CancelTx != nil:
		return cancelTx(ctx, m.k, contractAddr, contractMsg.CancelTx)
	}

	return nil, nil, nil, libwasm.ErrUnrecognizedMessage
}

func sendTx(ctx sdk.Context, k SkywayMsgServer, sender sdk.AccAddress, msg *bindingstypes.SendTx) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "validation failed")
	}

	dest, err := skywaytypes.NewEthAddress(msg.RemoteChainDestinationAddress)
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "invalid eth address")
	}

	amount, err := sdk.ParseCoinsNormalized(msg.Amount)
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "amount")
	}

	req := skywaytypes.NewMsgSendToRemote(sender, *dest, amount[0], msg.ChainReferenceId)
	_, err = k.SendToRemote(ctx, req)
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to dispatch message")
	}

	return nil, nil, nil, nil
}

func cancelTx(ctx sdk.Context, k SkywayMsgServer, sender sdk.AccAddress, msg *bindingstypes.CancelTx) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "validation failed")
	}

	req := skywaytypes.NewMsgCancelSendToRemote(sender, msg.TransactionId)
	_, err := k.CancelSendToRemote(ctx, req)
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to dispatch message")
	}

	return nil, nil, nil, nil
}

func handleSetErc20ToDenom(ctx sdk.Context, k SkywayMsgServer, sender sdk.AccAddress, msg *bindingstypes.SetErc20ToDenom) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "validation failed")
	}

	erc20, err := skywaytypes.NewEthAddress(msg.Erc20Address)
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "invalid eth address")
	}

	req := skywaytypes.NewMsgSetERC20ToTokenDenom(sender, *erc20, msg.ChainReferenceId, msg.TokenDenom)
	_, err = k.SetERC20ToTokenDenom(ctx, req)
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to dispatch message")
	}

	return nil, nil, nil, nil
}
