package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errtypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/palomachain/paloma/v2/util/libeth"
	tftypes "github.com/palomachain/paloma/v2/x/tokenfactory/types"
)

type Message struct {
	// Adds a new entry to the transaction pool to withdraw an amount
	// from the remote chain bridge contract. This will not execute
	// until a batch is requested and then actually relayed. Your
	// funds can be reclaimed using cancel-tx so long as they
	// remain in the pool.
	SendTx *SendTx `json:"send_tx,omitempty"`
	// Removes an entry from the transaction pool, preventing your
	// tokens from going to the remote chain and refunding the send.
	CancelTx *CancelTx `json:"cancel_tx,omitempty"`
	// Creates or updates the ERC20 tracking address mapping
	// for the given denom. Must have admin authority to do so.
	SetErc20ToDenom *SetErc20ToDenom `json:"set_erc_20_to_denom,omitempty"`
}

type (
	// Adds a new entry to the transaction pool to withdraw an amount
	// from the remote chain bridge contract. This will not execute
	// until a batch is requested and then actually relayed. Your
	// funds can be reclaimed using cancel-tx so long as they
	// remain in the pool.
	// RemoteChainDestinationAddress:address on the remote chain to send
	// the tokens to, e.g. 0x1234...5678
	// Amount: amount of tokens to send, e.g. 1000000000ugrain
	// ChainReferenceId: reference ID of the target chain, e.g. eth-main
	SendTx struct {
		RemoteChainDestinationAddress string `json:"remote_chain_destination_address"`
		Amount                        string `json:"amount"`
		ChainReferenceId              string `json:"chain_reference_id"`
	}

	// Removes an entry from the transaction pool, preventing your
	// tokens from going to the remote chain and refunding the send.
	// TransactionId: ID of the transaction to cancel
	CancelTx struct {
		TransactionId uint64 `json:"transaction_id"`
	}

	// Creates or updates the ERC20 tracking address mapping
	// for the given denom. Must have admin authority to do so.
	// Erc20Address: address of the ERC20 contract to track
	// TokenDenom: denom of the token to map to, e.g. ugrain
	// ChainReferenceId: reference ID of the target chain, e.g. eth-main
	SetErc20ToDenom struct {
		Erc20Address     string `json:"erc20_address"`
		TokenDenom       string `json:"token_denom"`
		ChainReferenceId string `json:"chain_reference_id"`
	}
)

func (m SendTx) ValidateBasic() error {
	err := libeth.ValidateEthAddress(m.RemoteChainDestinationAddress)
	if err != nil {
		return sdkerrors.Wrap(errtypes.ErrInvalidRequest, err.Error())
	}

	_, err = sdk.ParseCoinsNormalized(m.Amount)
	if err != nil {
		return sdkerrors.Wrap(err, "amount")
	}

	if m.ChainReferenceId == "" {
		return sdkerrors.Wrap(errtypes.ErrInvalidRequest, "chain_reference_id")
	}
	return nil
}

func (m CancelTx) ValidateBasic() error {
	if m.TransactionId == 0 {
		return sdkerrors.Wrap(errtypes.ErrInvalidRequest, "transaction_id")
	}
	return nil
}

func (m SetErc20ToDenom) ValidateBasic() error {
	err := libeth.ValidateEthAddress(m.Erc20Address)
	if err != nil {
		return sdkerrors.Wrap(errtypes.ErrInvalidRequest, err.Error())
	}
	if m.ChainReferenceId == "" {
		return sdkerrors.Wrap(errtypes.ErrInvalidRequest, "chain_reference_id")
	}
	if _, _, err = tftypes.DeconstructDenom(m.TokenDenom); err != nil {
		return sdkerrors.Wrap(errtypes.ErrInvalidRequest, "chain_reference_id")
	}
	return nil
}
