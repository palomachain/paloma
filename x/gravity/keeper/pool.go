package keeper

import (
	"fmt"
	"sort"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/palomachain/paloma/x/gravity/types"
)

// Transaction pool functions, used to manage pending Sends to Ethereum and the creation of BatchFees

// AddToOutgoingPool creates a transaction and adds it to the pool, returns the id of the unbatched transaction
// - checks a counterpart denominator exists for the given voucher type
// - burns the voucher for transfer amount and fees
// - persists an OutgoingTx
// - adds the TX to the `available` TX pool
func (k Keeper) AddToOutgoingPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	counterpartReceiver types.EthAddress,
	amount sdk.Coin,
	fee sdk.Coin,
) (uint64, error) {
	if ctx.IsZero() || sdk.VerifyAddressFormat(sender) != nil || counterpartReceiver.ValidateBasic() != nil ||
		!amount.IsValid() || !fee.IsValid() || fee.Denom != amount.Denom {
		return 0, sdkerrors.Wrap(types.ErrInvalid, "arguments")
	}
	totalAmount := amount.Add(fee)
	totalInVouchers := sdk.Coins{totalAmount}

	// If the coin is a gravity voucher, burn the coins. If not, check if there is a deployed ERC20 contract representing it.
	// If there is, lock the coins.

	_, tokenContract, err := k.DenomToERC20Lookup(ctx, totalAmount.Denom)
	if err != nil {
		return 0, err
	}

	// lock coins in module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, totalInVouchers); err != nil {
		return 0, err
	}

	// get next tx id from keeper
	nextID := k.autoIncrementID(ctx, types.KeyLastTXPoolID)

	erc20Fee, err := types.NewInternalERC20Token(fee.Amount, tokenContract.GetAddress().Hex())
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "invalid Erc20Fee from amount %d and contract %v",
			fee.Amount, tokenContract)
	}
	erc20Token, err := types.NewInternalERC20Token(amount.Amount, tokenContract.GetAddress().Hex())
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
		Erc20Fee:    erc20Fee.ToExternal(),
	}.ToInternal()
	if err != nil { // This should never happen since all the components are validated
		panic(sdkerrors.Wrap(err, "unable to create InternalOutgoingTransferTx"))
	}

	// add a second index with the fee
	err = k.addUnbatchedTX(ctx, outgoing)
	if err != nil {
		panic(err)
	}

	// todo: add second index for sender so that we can easily query: give pending Tx by sender
	// todo: what about a second index for receiver?

	return nextID, ctx.EventManager().EmitTypedEvent(
		&types.EventWithdrawalReceived{
			BridgeContract: k.GetBridgeContractAddress(ctx).GetAddress().Hex(),
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
func (k Keeper) RemoveFromOutgoingPoolAndRefund(ctx sdk.Context, txId uint64, sender sdk.AccAddress) error {
	if ctx.IsZero() || txId < 1 || sdk.VerifyAddressFormat(sender) != nil {
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

	// An inconsistent entry should never enter the store, but this is the ideal place to exploit
	// it such a bug if it did ever occur, so we should double check to be really sure
	if tx.Erc20Fee.Contract != tx.Erc20Token.Contract {
		return sdkerrors.Wrapf(types.ErrInvalid, "Inconsistent tokens to cancel!: %s %s", tx.Erc20Fee.Contract.GetAddress().Hex(), tx.Erc20Token.Contract.GetAddress().Hex())
	}

	// delete this tx from the pool
	err = k.removeUnbatchedTX(ctx, *tx.Erc20Fee, txId)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalid, "txId %d not in unbatched index! Must be in a batch!", txId)
	}
	// Make sure the tx was removed
	oldTx, oldTxErr := k.GetUnbatchedTxByFeeAndId(ctx, *tx.Erc20Fee, tx.Id)
	if oldTx != nil || oldTxErr == nil {
		return sdkerrors.Wrapf(types.ErrInvalid, "tx with id %d was not fully removed from the pool, a duplicate must exist", txId)
	}

	// Calculate refund
	_, denom := k.ERC20ToDenomLookup(ctx, tx.Erc20Token.Contract)
	totalToRefund := sdk.NewCoin(denom, tx.Erc20Token.Amount)
	totalToRefund.Amount = totalToRefund.Amount.Add(tx.Erc20Fee.Amount)
	totalToRefundCoins := sdk.NewCoins(totalToRefund)

	// Perform refund
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, totalToRefundCoins); err != nil {
		return sdkerrors.Wrap(err, "transfer vouchers")
	}

	return ctx.EventManager().EmitTypedEvent(
		&types.EventWithdrawCanceled{
			Sender:         sender.String(),
			TxId:           fmt.Sprint(txId),
			BridgeContract: k.GetBridgeContractAddress(ctx).GetAddress().Hex(),
			BridgeChainId:  strconv.Itoa(int(k.GetBridgeChainID(ctx))),
		},
	)
}

// addUnbatchedTx creates a new transaction in the pool
// WARNING: Do not make this function public
func (k Keeper) addUnbatchedTX(ctx sdk.Context, val *types.InternalOutgoingTransferTx) error {
	store := ctx.KVStore(k.storeKey)
	idxKey := types.GetOutgoingTxPoolKey(*val.Erc20Fee, val.Id)
	if store.Has(idxKey) {
		return sdkerrors.Wrap(types.ErrDuplicate, "transaction already in pool")
	}

	extVal := val.ToExternal()

	bz, err := k.cdc.Marshal(&extVal)
	if err != nil {
		return err
	}

	store.Set(idxKey, bz)
	return err
}

// removeUnbatchedTXIndex removes the tx from the pool
// WARNING: Do not make this function public
func (k Keeper) removeUnbatchedTX(ctx sdk.Context, fee types.InternalERC20Token, txID uint64) error {
	store := ctx.KVStore(k.storeKey)
	idxKey := types.GetOutgoingTxPoolKey(fee, txID)
	if !store.Has(idxKey) {
		return sdkerrors.Wrap(types.ErrUnknown, "pool transaction")
	}
	store.Delete(idxKey)
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////
//////////////// Unbatched Tx Search and Collection Methods ///////////////////////////
///////////////////////////////////////////////////////////////////////////////////////

// GetUnbatchedTxByFeeAndId grabs a tx from the pool given its fee and txID
func (k Keeper) GetUnbatchedTxByFeeAndId(ctx sdk.Context, fee types.InternalERC20Token, txID uint64) (*types.InternalOutgoingTransferTx, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOutgoingTxPoolKey(fee, txID))
	if bz == nil {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "pool transaction")
	}
	var r types.OutgoingTransferTx
	err := k.cdc.Unmarshal(bz, &r)
	if err != nil {
		panic(sdkerrors.Wrapf(err, "invalid unbatched tx in store: %v", r))
	}
	intR, err := r.ToInternal()
	if err != nil {
		panic(sdkerrors.Wrapf(err, "invalid unbatched tx in store: %v", r))
	}
	return intR, nil
}

// GetUnbatchedTxById grabs a tx from the pool given only the txID
// note that due to the way unbatched txs are indexed, the GetUnbatchedTxByFeeAndId method is much faster
func (k Keeper) GetUnbatchedTxById(ctx sdk.Context, txID uint64) (*types.InternalOutgoingTransferTx, error) {
	var r *types.InternalOutgoingTransferTx = nil
	k.IterateUnbatchedTransactions(ctx, func(_ []byte, tx *types.InternalOutgoingTransferTx) bool {
		if tx.Id == txID {
			r = tx
			return true
		}
		return false // iterating DESC, exit early
	})

	if r == nil {
		// We have no return tx, it was either batched or never existed
		return nil, sdkerrors.Wrap(types.ErrUnknown, "pool transaction")
	}
	return r, nil
}

// GetUnbatchedTransactionsByContract grabs all unbatched transactions from the tx pool for the given contract
// unbatched transactions are sorted by fee amount in DESC order
func (k Keeper) GetUnbatchedTransactionsByContract(ctx sdk.Context, contractAddress types.EthAddress) []*types.InternalOutgoingTransferTx {
	return k.collectUnbatchedTransactions(ctx, types.GetOutgoingTxPoolContractPrefix(contractAddress))
}

// GetUnbatchedTransactions grabs all transactions from the tx pool, useful for queries or genesis save/load
func (k Keeper) GetUnbatchedTransactions(ctx sdk.Context) []*types.InternalOutgoingTransferTx {
	return k.collectUnbatchedTransactions(ctx, types.OutgoingTXPoolKey)
}

// Aggregates all unbatched transactions in the store with a given prefix
func (k Keeper) collectUnbatchedTransactions(ctx sdk.Context, prefixKey []byte) (out []*types.InternalOutgoingTransferTx) {
	k.filterAndIterateUnbatchedTransactions(ctx, prefixKey, func(_ []byte, tx *types.InternalOutgoingTransferTx) bool {
		out = append(out, tx)
		return false
	})
	return
}

// IterateUnbatchedTransactionsByContract iterates through unbatched transactions from the tx pool for the given contract,
// executing the given callback on each discovered Tx. Return true in cb to stop iteration, false to continue.
// unbatched transactions are sorted by fee amount in DESC order
func (k Keeper) IterateUnbatchedTransactionsByContract(ctx sdk.Context, contractAddress types.EthAddress, cb func(key []byte, tx *types.InternalOutgoingTransferTx) bool) {
	k.filterAndIterateUnbatchedTransactions(ctx, types.GetOutgoingTxPoolContractPrefix(contractAddress), cb)
}

// IterateUnbatchedTransactions iterates through all unbatched transactions in DESC order, executing the given callback
// on each discovered Tx. Return true in cb to stop iteration, false to continue.
// For finer grained control, use filterAndIterateUnbatchedTransactions or one of the above methods
func (k Keeper) IterateUnbatchedTransactions(ctx sdk.Context, cb func(key []byte, tx *types.InternalOutgoingTransferTx) (stop bool)) {
	k.filterAndIterateUnbatchedTransactions(ctx, types.OutgoingTXPoolKey, cb)
}

// filterAndIterateUnbatchedTransactions iterates through all unbatched transactions whose keys begin with prefixKey in DESC order
// prefixKey should be either OutgoingTXPoolKey or some more granular key, passing the wrong key will cause a panic
func (k Keeper) filterAndIterateUnbatchedTransactions(ctx sdk.Context, prefixKey []byte, cb func(key []byte, tx *types.InternalOutgoingTransferTx) bool) {
	prefixStore := ctx.KVStore(k.storeKey)
	iter := prefixStore.ReverseIterator(prefixRange(prefixKey))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var transact types.OutgoingTransferTx
		k.cdc.MustUnmarshal(iter.Value(), &transact)
		intTx, err := transact.ToInternal()
		if err != nil {
			panic(sdkerrors.Wrapf(err, "invalid unbatched transaction in store: %v", transact))
		}
		// cb returns true to stop early
		if cb(iter.Key(), intTx) {
			break
		}
	}
}

///////////////////////////////////////////////////////////////////////////////////////
//////////////////////////// Batch Fee Methods ////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////

// GetBatchFeeByTokenType picks transactions in the pool for a potential batch and gives the total fees the batch would
// grant if created right now. This info is both presented to relayers for the purpose of determining when to request
// batches and also used by the batch creation process to decide not to create a new batch (fees must be increasing)
func (k Keeper) GetBatchFeeByTokenType(ctx sdk.Context, tokenContractAddr types.EthAddress, maxElements uint) *types.BatchFees {
	batchFee := types.BatchFees{Token: tokenContractAddr.GetAddress().Hex(), TotalFees: sdk.NewInt(0), TxCount: 0}

	// Since transactions are stored with keys [ prefix | contract | fee_amount] and since this iterator returns results
	// in DESC order, we can safely pick the first N and have a batch with maximal fees for relaying
	k.IterateUnbatchedTransactionsByContract(ctx, tokenContractAddr, func(key []byte, tx *types.InternalOutgoingTransferTx) bool {
		if !k.IsOnBlacklist(ctx, *tx.DestAddress) {
			fee := tx.Erc20Fee
			if fee.Contract.GetAddress() != tokenContractAddr.GetAddress() {
				panic(fmt.Errorf("unexpected fee contract %s under key %v when getting batch fees for contract %s", fee.Contract.GetAddress().Hex(), key, tokenContractAddr.GetAddress().Hex()))
			}
			batchFee.TotalFees = batchFee.TotalFees.Add(fee.Amount)
			batchFee.TxCount += 1
			return batchFee.TxCount == uint64(maxElements)
		} else {
			// if the tx was on the blacklist we return false
			// to continue to the next loop iteration
			return false
		}
	})
	return &batchFee
}

// GetAllBatchFees creates a fee entry for every batch type currently in the store
// this can be used by relayers to determine what batch types are desireable to request
func (k Keeper) GetAllBatchFees(ctx sdk.Context, maxElements uint) (batchFees []types.BatchFees) {
	batchFeesMap := k.createBatchFees(ctx, maxElements)
	// create array of batchFees
	for _, batchFee := range batchFeesMap {
		batchFees = append(batchFees, batchFee)
	}

	// quick sort by token to make this function safe for use
	// in consensus computations
	sort.Slice(batchFees, func(i, j int) bool {
		return batchFees[i].Token < batchFees[j].Token
	})

	return batchFees
}

// createBatchFees iterates over the unbatched transaction pool and creates batch token fee map
// Implicitly creates batches with the highest potential fee because the transaction keys enforce an order which goes
// fee contract address -> fee amount -> transaction nonce
func (k Keeper) createBatchFees(ctx sdk.Context, maxElements uint) map[string]types.BatchFees {
	batchFeesMap := make(map[string]types.BatchFees)

	k.IterateUnbatchedTransactions(ctx, func(_ []byte, tx *types.InternalOutgoingTransferTx) bool {
		feeAddrStr := tx.Erc20Fee.Contract.GetAddress()

		if fees, ok := batchFeesMap[feeAddrStr.Hex()]; ok {
			if fees.TxCount < uint64(maxElements) {
				fees.TotalFees = batchFeesMap[feeAddrStr.Hex()].TotalFees.Add(tx.Erc20Fee.Amount)
				fees.TxCount++
				batchFeesMap[feeAddrStr.Hex()] = fees
			}
		} else {
			batchFeesMap[feeAddrStr.Hex()] = types.BatchFees{
				Token:     feeAddrStr.Hex(),
				TotalFees: tx.Erc20Fee.Amount,
				TxCount:   1,
			}
		}

		return false
	})

	return batchFeesMap
}

// a specialized function used for iterating store counters, handling
// returning, initializing and incrementing all at once. This is particularly
// used for the transaction pool and batch pool where each batch or transaction is
// assigned a unique ID.
func (k Keeper) autoIncrementID(ctx sdk.Context, idKey []byte) uint64 {
	id := k.getID(ctx, idKey)
	id += 1
	k.setID(ctx, id, idKey)
	return id
}

// gets a generic uint64 counter from the store, initializing to 1 if no value exists
func (k Keeper) getID(ctx sdk.Context, idKey []byte) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(idKey)
	id := types.UInt64FromBytesUnsafe(bz)
	return id
}

// sets a generic uint64 counter in the store
func (k Keeper) setID(ctx sdk.Context, id uint64, idKey []byte) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.Uint64ToBigEndian(id)
	store.Set(idKey, bz)
}
