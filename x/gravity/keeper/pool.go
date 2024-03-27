package keeper

import (
	"context"
	"fmt"
	"strconv"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
)

// Transaction pool functions, used to manage pending Sends to Ethereum

// AddToOutgoingPool creates a transaction and adds it to the pool, returns the id of the unbatched transaction
// - checks a counterpart denominator exists for the given voucher type
// - burns the voucher for transfer amount
// - persists an OutgoingTx
// - adds the TX to the `available` TX pool
func (k Keeper) AddToOutgoingPool(
	ctx context.Context,
	sender sdk.AccAddress,
	counterpartReceiver types.EthAddress,
	amount sdk.Coin,
	chainReferenceID string,
) (uint64, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.IsZero() || sdk.VerifyAddressFormat(sender) != nil || counterpartReceiver.ValidateBasic() != nil ||
		!amount.IsValid() {
		return 0, sdkerrors.Wrap(types.ErrInvalid, "arguments")
	}
	amountInVouchers := sdk.Coins{amount}

	// If the coin is a gravity voucher, burn the coins. If not, check if there is a deployed ERC20 contract representing it.
	// If there is, lock the coins.

	tokenContract, err := k.GetERC20OfDenom(ctx, chainReferenceID, amount.Denom)
	if err != nil {
		return 0, err
	}
	// lock coins in module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amountInVouchers); err != nil {
		return 0, err
	}

	// get next tx id from keeper
	nextID, err := k.autoIncrementID(ctx, types.KeyLastTXPoolID)
	if err != nil {
		return 0, err
	}

	erc20Token, err := types.NewInternalERC20Token(amount.Amount, tokenContract.GetAddress().Hex(), chainReferenceID)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "invalid ERC20Token from amount %d and contract %v",
			amount.Amount, tokenContract)
	}
	// construct outgoing tx, as part of this process we represent
	// the token as an ERC20 token since it is preparing to go to ETH
	// rather than the denom that is the input to this function.
	outgoing, err := types.OutgoingTransferTx{
		Id:          nextID,
		Sender:      sender.String(),
		DestAddress: counterpartReceiver.GetAddress().Hex(),
		Erc20Token:  erc20Token.ToExternal(),
	}.ToInternal()
	if err != nil { // This should never happen since all the components are validated
		return 0, sdkerrors.Wrap(err, "unable to create InternalOutgoingTransferTx")
	}

	err = k.addUnbatchedTX(ctx, outgoing)
	if err != nil {
		return 0, err
	}

	bridgeContractAddress, err := k.GetBridgeContractAddress(ctx)
	if err != nil {
		return 0, err
	}
	return nextID, sdkCtx.EventManager().EmitTypedEvent(
		&types.EventWithdrawalReceived{
			BridgeContract: bridgeContractAddress.GetAddress().Hex(),
			BridgeChainId:  strconv.Itoa(int(k.GetBridgeChainID(ctx))),
			OutgoingTxId:   strconv.Itoa(int(nextID)),
			Nonce:          fmt.Sprint(nextID),
		},
	)
}

// RemoveFromOutgoingPoolAndRefund performs the cancel function for pending Sends To Ethereum
// - checks that the provided tx actually exists
// - deletes the unbatched tx from the pool
// - issues the tokens back to the sender
func (k Keeper) RemoveFromOutgoingPoolAndRefund(ctx context.Context, txId uint64, sender sdk.AccAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.IsZero() || txId < 1 || sdk.VerifyAddressFormat(sender) != nil {
		return sdkerrors.Wrap(types.ErrInvalid, "arguments")
	}
	// check that we actually have a tx with that id and what it's details are
	tx, err := k.GetUnbatchedTxById(ctx, txId)
	if err != nil {
		return sdkerrors.Wrapf(err, "unknown transaction with id %d from sender %s", txId, sender.String())
	}

	// Check that this user actually sent the transaction, this prevents someone from refunding someone
	// elses transaction to themselves.
	if !tx.Sender.Equals(sender) {
		return sdkerrors.Wrapf(types.ErrInvalid, "Sender %s did not send Id %d", sender, txId)
	}

	// delete this tx from the pool
	err = k.removeUnbatchedTX(ctx, *tx.Erc20Token, txId)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalid, "txId %d not in unbatched index! Must be in a batch!", txId)
	}
	// Make sure the tx was removed
	oldTx, oldTxErr := k.GetUnbatchedTxByAmountAndId(ctx, *tx.Erc20Token, tx.Id)
	if oldTx != nil || oldTxErr == nil {
		return sdkerrors.Wrapf(types.ErrInvalid, "tx with id %d was not fully removed from the pool, a duplicate must exist", txId)
	}

	// Calculate refund
	denom, err := k.GetDenomOfERC20(ctx, tx.Erc20Token.ChainReferenceID, tx.Erc20Token.Contract)
	if err != nil {
		return err
	}
	totalToRefund := sdk.NewCoin(denom, tx.Erc20Token.Amount)
	totalToRefundCoins := sdk.NewCoins(totalToRefund)

	// Perform refund
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, totalToRefundCoins); err != nil {
		return sdkerrors.Wrap(err, "transfer vouchers")
	}

	bridgeContractAddress, err := k.GetBridgeContractAddress(ctx)
	if err != nil {
		return err
	}

	return sdkCtx.EventManager().EmitTypedEvent(
		&types.EventWithdrawCanceled{
			Sender:         sender.String(),
			TxId:           fmt.Sprint(txId),
			BridgeContract: bridgeContractAddress.GetAddress().Hex(),
			BridgeChainId:  strconv.Itoa(int(k.GetBridgeChainID(ctx))),
		},
	)
}

// addUnbatchedTx creates a new transaction in the pool
// WARNING: Do not make this function public
func (k Keeper) addUnbatchedTX(ctx context.Context, val *types.InternalOutgoingTransferTx) error {
	store := k.GetStore(ctx)
	idxKey := types.GetOutgoingTxPoolKey(*val.Erc20Token, val.Id)
	if store.Has(idxKey) {
		return sdkerrors.Wrap(types.ErrDuplicate, "transaction already in pool")
	}

	extVal := val.ToExternal()

	bz, err := k.cdc.Marshal(&extVal)
	if err != nil {
		return err
	}
	store.Set(idxKey, bz)
	return nil
}

// removeUnbatchedTXIndex removes the tx from the pool
// WARNING: Do not make this function public
func (k Keeper) removeUnbatchedTX(ctx context.Context, token types.InternalERC20Token, txID uint64) error {
	store := k.GetStore(ctx)
	idxKey := types.GetOutgoingTxPoolKey(token, txID)
	if !store.Has(idxKey) {
		return sdkerrors.Wrap(types.ErrUnknown, "pool transaction")
	}
	store.Delete(idxKey)
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////
//////////////// Unbatched Tx Search and Collection Methods ///////////////////////////
///////////////////////////////////////////////////////////////////////////////////////

// GetUnbatchedTxByAmountAndId grabs a tx from the pool given its txID
func (k Keeper) GetUnbatchedTxByAmountAndId(ctx context.Context, token types.InternalERC20Token, txID uint64) (*types.InternalOutgoingTransferTx, error) {
	store := k.GetStore(ctx)
	bz := store.Get(types.GetOutgoingTxPoolKey(token, txID))
	if bz == nil {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "pool transaction")
	}
	var r types.OutgoingTransferTx
	err := k.cdc.Unmarshal(bz, &r)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "invalid unbatched tx in store: %v", r)
	}
	intR, err := r.ToInternal()
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "invalid unbatched tx in store: %v", r)
	}
	return intR, nil
}

// GetUnbatchedTxById grabs a tx from the pool given only the txID
// note that due to the way unbatched txs are indexed, the GetUnbatchedTxByAmountAndId method is much faster
func (k Keeper) GetUnbatchedTxById(ctx context.Context, txID uint64) (*types.InternalOutgoingTransferTx, error) {
	var r *types.InternalOutgoingTransferTx = nil
	err := k.IterateUnbatchedTransactions(ctx, func(_ []byte, tx *types.InternalOutgoingTransferTx) bool {
		if tx.Id == txID {
			r = tx
			return true
		}
		return false // iterating DESC, exit early
	})
	if err != nil {
		return nil, err
	}

	if r == nil {
		// We have no return tx, it was either batched or never existed
		return nil, sdkerrors.Wrap(types.ErrUnknown, "pool transaction")
	}
	return r, nil
}

// GetUnbatchedTransactionsByContract grabs all unbatched transactions from the tx pool for the given contract
// unbatched transactions are sorted by nonce in DESC order
func (k Keeper) GetUnbatchedTransactionsByContract(ctx context.Context, contractAddress types.EthAddress) ([]*types.InternalOutgoingTransferTx, error) {
	return k.collectUnbatchedTransactions(ctx, types.GetOutgoingTxPoolContractPrefix(contractAddress))
}

// GetUnbatchedTransactions grabs all transactions from the tx pool, useful for queries or genesis save/load
func (k Keeper) GetUnbatchedTransactions(ctx context.Context) ([]*types.InternalOutgoingTransferTx, error) {
	return k.collectUnbatchedTransactions(ctx, types.OutgoingTXPoolKey)
}

// Aggregates all unbatched transactions in the store with a given prefix
func (k Keeper) collectUnbatchedTransactions(ctx context.Context, prefixKey []byte) (out []*types.InternalOutgoingTransferTx, err error) {
	err = k.filterAndIterateUnbatchedTransactions(ctx, prefixKey, func(_ []byte, tx *types.InternalOutgoingTransferTx) bool {
		out = append(out, tx)
		return false
	})
	return
}

// IterateUnbatchedTransactionsByContract iterates through unbatched transactions from the tx pool for the given contract,
// executing the given callback on each discovered Tx. Return true in cb to stop iteration, false to continue.
// unbatched transactions are sorted by nonce in DESC order
func (k Keeper) IterateUnbatchedTransactionsByContract(ctx context.Context, contractAddress types.EthAddress, cb func(key []byte, tx *types.InternalOutgoingTransferTx) bool) error {
	return k.filterAndIterateUnbatchedTransactions(ctx, types.GetOutgoingTxPoolContractPrefix(contractAddress), cb)
}

// IterateUnbatchedTransactions iterates through all unbatched transactions in DESC order, executing the given callback
// on each discovered Tx. Return true in cb to stop iteration, false to continue.
// For finer grained control, use filterAndIterateUnbatchedTransactions or one of the above methods
func (k Keeper) IterateUnbatchedTransactions(ctx context.Context, cb func(key []byte, tx *types.InternalOutgoingTransferTx) (stop bool)) error {
	return k.filterAndIterateUnbatchedTransactions(ctx, types.OutgoingTXPoolKey, cb)
}

// filterAndIterateUnbatchedTransactions iterates through all unbatched transactions whose keys begin with prefixKey in DESC order
// prefixKey should be either OutgoingTXPoolKey or some more granular key, passing the wrong key will cause an error
func (k Keeper) filterAndIterateUnbatchedTransactions(ctx context.Context, prefixKey []byte, cb func(key []byte, tx *types.InternalOutgoingTransferTx) bool) error {
	prefixStore := k.GetStore(ctx)
	start, end, err := prefixRange(prefixKey)
	if err != nil {
		return err
	}
	iter := prefixStore.ReverseIterator(start, end)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var transact types.OutgoingTransferTx
		k.cdc.MustUnmarshal(iter.Value(), &transact)
		intTx, err := transact.ToInternal()
		if err != nil {
			return sdkerrors.Wrapf(err, "invalid unbatched transaction in store: %v", transact)
		}
		// cb returns true to stop early
		if cb(iter.Key(), intTx) {
			break
		}
	}
	return nil
}

// a specialized function used for iterating store counters, handling
// returning, initializing and incrementing all at once. This is particularly
// used for the transaction pool and batch pool where each batch or transaction is
// assigned a unique ID.
func (k Keeper) autoIncrementID(ctx context.Context, idKey []byte) (uint64, error) {
	id, err := k.getID(ctx, idKey)
	if err != nil {
		return 0, err
	}
	id += 1
	k.setID(ctx, id, idKey)
	return id, nil
}

// gets a generic uint64 counter from the store, initializing to 1 if no value exists
func (k Keeper) getID(ctx context.Context, idKey []byte) (uint64, error) {
	store := k.GetStore(ctx)
	bz := store.Get(idKey)
	id, err := types.UInt64FromBytes(bz)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// sets a generic uint64 counter in the store
func (k Keeper) setID(ctx context.Context, id uint64, idKey []byte) {
	store := k.GetStore(ctx)
	bz := sdk.Uint64ToBigEndian(id)
	store.Set(idKey, bz)
}
