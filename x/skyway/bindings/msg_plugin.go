package bindings

import (
	"context"
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

var _ wasmkeeper.Messenger = (*customMessenger)(nil)

func NewMessenger(k SkywayMsgServer) wasmkeeper.Messenger {
	return &customMessenger{
		k: k,
	}
}

func (m *customMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	if msg.Custom != nil {
		var contractMsg bindingstypes.Message
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, nil, sdkerrors.Wrap(err, "skyway msg")
		}

		switch {
		case contractMsg.SetErc20ToDenom != nil:
		case contractMsg.SendTx != nil:
		case contractMsg.CancelTx != nil:
			return nil, nil, nil, nil
		}
	}
	return nil, nil, nil, nil
}

func sendTx(ctx sdk.Context, k SkywayMsgServer, sender sdk.AccAddress, msg bindingstypes.SendTx) error {
	if err := msg.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "validation failed")
	}

	dest, err := skywaytypes.NewEthAddress(msg.RemoteChainDestinationAddress)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid eth address")
	}

	amount, err := sdk.ParseCoinsNormalized(msg.Amount)
	if err != nil {
		return sdkerrors.Wrap(err, "amount")
	}

	req := skywaytypes.NewMsgSendToRemote(sender, *dest, amount[0], msg.ChainReferenceId)
	_, err = k.SendToRemote(ctx, req)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to dispatch message")
	}

	return nil
}

func cancelTx(ctx sdk.Context, k SkywayMsgServer, sender sdk.AccAddress, msg bindingstypes.CancelTx) error {
	if err := msg.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "validation failed")
	}

	req := skywaytypes.NewMsgCancelSendToRemote(sender, msg.TransactionId)
	_, err := k.CancelSendToRemote(ctx, req)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to dispatch message")
	}

	return nil
}

func handleSetErc20ToDenom(ctx sdk.Context, k SkywayMsgServer, sender sdk.AccAddress, msg bindingstypes.SetErc20ToDenom) error {
	if err := msg.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "validation failed")
	}

	erc20, err := skywaytypes.NewEthAddress(msg.Erc20Address)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid eth address")
	}

	req := skywaytypes.NewMsgSetERC20ToTokenDenom(sender, *erc20, msg.ChainReferenceId, msg.TokenDenom)
	_, err = k.SetERC20ToTokenDenom(ctx, req)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to dispatch message")
	}

	return nil
}
