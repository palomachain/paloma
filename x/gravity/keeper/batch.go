package keeper

import (
	"context"
	"fmt"
	"strconv"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
)

const OutgoingTxBatchSize = 100

// BuildOutgoingTXBatch starts the following process chain:
// - find bridged denominator for given voucher type
// - select available transactions from the outgoing transaction pool sorted by nonce desc
// - persist an outgoing batch object with an incrementing ID = nonce
// - emit an event
func (k Keeper) BuildOutgoingTXBatch(
	ctx context.Context,
	chainReferenceID string,
	contract types.EthAddress,
	maxElements uint,
) (*types.InternalOutgoingTxBatch, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if maxElements == 0 {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "max elements value")
	}

	selectedTxs, err := k.pickUnbatchedTxs(ctx, contract, maxElements)
	if err != nil {
		return nil, err
	} else if len(selectedTxs) == 0 {
		return nil, nil // Nothing to batch, so do nothing
	}

	ci, err := k.evmKeeper.GetChainInfo(ctx, chainReferenceID)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to create batch")
	}
	turnstoneID := string(ci.SmartContractUniqueID)

	nextID, err := k.autoIncrementID(ctx, types.KeyLastOutgoingBatchID)
	if err != nil {
		return nil, err
	}

	assignee, err := k.evmKeeper.PickValidatorForMessage(ctx, chainReferenceID, nil)
	if err != nil {
		return nil, err
	}

	batch, err := types.NewInternalOutgingTxBatch(nextID, k.getBatchTimeoutHeight(ctx), selectedTxs, contract, 0, chainReferenceID, turnstoneID, assignee)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to create batch")
	}
	// set the current block height when storing the batch
	batch.PalomaBlockCreated = uint64(sdkCtx.BlockHeight())
	err = k.StoreBatch(ctx, *batch)
	if err != nil {
		return nil, err
	}
	// Get the checkpoint and store it as a legit past batch
	checkpoint, err := batch.GetCheckpoint(turnstoneID)
	if err != nil {
		return nil, err
	}
	k.SetPastEthSignatureCheckpoint(ctx, checkpoint)

	bridgeContract, err := k.GetBridgeContractAddress(ctx)
	if err != nil {
		return nil, err
	}

	return batch, sdkCtx.EventManager().EmitTypedEvent(
		&types.EventOutgoingBatch{
			BridgeContract: bridgeContract.GetAddress().Hex(),
			BridgeChainId:  strconv.Itoa(int(k.GetBridgeChainID(ctx))),
			BatchId:        string(types.GetOutgoingTxBatchKey(contract, nextID)),
			Nonce:          fmt.Sprint(nextID),
			Assignee:       assignee,
		},
	)
}

func (k Keeper) getBatchTimeoutHeight(ctx context.Context) uint64 {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return uint64(sdkCtx.BlockTime().Add(10 * time.Minute).Unix())
}

// OutgoingTxBatchExecuted is run when the Cosmos chain detects that a batch has been executed on Ethereum
// It frees all the transactions in the batch
func (k Keeper) OutgoingTxBatchExecuted(ctx context.Context, tokenContract types.EthAddress, claim types.MsgBatchSendToEthClaim) error {
	b, err := k.GetOutgoingTXBatch(ctx, tokenContract, claim.BatchNonce)
	if err != nil {
		return err
	}
	if b == nil {
		return fmt.Errorf("unknown batch nonce for outgoing tx batch %s %d", tokenContract.GetAddress().Hex(), claim.BatchNonce)
	}
	if b.BatchTimeout <= claim.EthBlockHeight {
		return fmt.Errorf("Batch with nonce %d submitted after it timed out (submission %d >= timeout %d)?", claim.BatchNonce, claim.EthBlockHeight, b.BatchTimeout)
	}

	totalToBurn := math.NewInt(0)
	for _, tx := range b.Transactions {
		totalToBurn = totalToBurn.Add(tx.Erc20Token.Amount)
	}

	denom, err := k.GetDenomOfERC20(ctx, claim.GetChainReferenceId(), tokenContract)
	if err != nil {
		return err
	}

	burnVouchers := sdk.NewCoins(sdk.NewCoin(denom, totalToBurn))
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnVouchers); err != nil {
		return err
	}

	// Delete batch and confirms since it is finished
	err = k.DeleteBatch(ctx, *b)
	if err != nil {
		return err
	}

	return k.DeleteBatchConfirms(ctx, *b)
}

// StoreBatch stores a transaction batch, it will refuse to overwrite an existing
// batch and errors instead, once a batch is stored in state signature collection begins
// so no mutation of a batch in state can ever be valid
func (k Keeper) StoreBatch(ctx context.Context, batch types.InternalOutgoingTxBatch) error {
	if err := batch.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "attempted to store invalid batch")
	}
	externalBatch := batch.ToExternal()
	store := k.GetStore(ctx)
	key := types.GetOutgoingTxBatchKey(batch.TokenContract, batch.BatchNonce)
	if store.Has(key) {
		return fmt.Errorf("should never overwrite batch")
	}
	store.Set(key, k.cdc.MustMarshal(&externalBatch))
	return nil
}

// DeleteBatch deletes an outgoing transaction batch
func (k Keeper) DeleteBatch(ctx context.Context, batch types.InternalOutgoingTxBatch) error {
	if err := batch.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "attempted to delete invalid batch")
	}
	store := k.GetStore(ctx)
	outgoingTxBatchKey := types.GetOutgoingTxBatchKey(batch.TokenContract, batch.BatchNonce)
	store.Delete(outgoingTxBatchKey)
	return nil
}

// pickUnbatchedTxs moves unbatched Txs from the pool into a collection ready for batching
func (k Keeper) pickUnbatchedTxs(
	ctx context.Context,
	contractAddress types.EthAddress,
	maxElements uint,
) ([]*types.InternalOutgoingTransferTx, error) {
	var selectedTxs []*types.InternalOutgoingTransferTx
	var g whoops.Group
	g.Add(
		k.IterateUnbatchedTransactionsByContract(ctx, contractAddress, func(_ []byte, tx *types.InternalOutgoingTransferTx) bool {
			if tx != nil && tx.Erc20Token != nil {
				selectedTxs = append(selectedTxs, tx)
				err := k.removeUnbatchedTX(ctx, *tx.Erc20Token, tx.Id)
				if err != nil {
					g.Add(sdkerrors.Wrap(err, "failed to remote tx from unbatched queue"))
					return true
				}

				// double check that no duplicates exist in the index
				oldTx, oldTxErr := k.GetUnbatchedTxByAmountAndId(ctx, *tx.Erc20Token, tx.Id)
				if oldTx != nil || oldTxErr == nil {
					g.Add(sdkerrors.Wrap(err, "picked a duplicate transaction from the pool, duplicates should never exist"))
					return true
				}

				return uint(len(selectedTxs)) == maxElements
			} else {
				g.Add(fmt.Errorf("tx should never be nil"))
				return true
			}
		}),
	)
	if len(g) > 0 {
		return nil, g
	}
	return selectedTxs, nil
}

// GetOutgoingTXBatch loads a batch object. Returns nil when not exists.
func (k Keeper) GetOutgoingTXBatch(ctx context.Context, tokenContract types.EthAddress, nonce uint64) (*types.InternalOutgoingTxBatch, error) {
	store := k.GetStore(ctx)
	key := types.GetOutgoingTxBatchKey(tokenContract, nonce)
	bz := store.Get(key)
	if len(bz) == 0 {
		return nil, nil
	}
	var b types.OutgoingTxBatch
	k.cdc.MustUnmarshal(bz, &b)
	for _, tx := range b.Transactions {
		tx.Erc20Token.Contract = tokenContract.GetAddress().Hex()
	}
	ret, err := b.ToInternal()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "found invalid batch in store")
	}
	return ret, err
}

// CancelOutgoingTXBatch releases all TX in the batch and deletes the batch
func (k Keeper) CancelOutgoingTXBatch(ctx context.Context, tokenContract types.EthAddress, nonce uint64) error {
	batch, err := k.GetOutgoingTXBatch(ctx, tokenContract, nonce)
	if err != nil {
		return err
	}
	if batch == nil {
		return types.ErrUnknown
	}
	for _, tx := range batch.Transactions {
		err := k.addUnbatchedTX(ctx, tx)
		if err != nil {
			return sdkerrors.Wrapf(err, "unable to add batched transaction back into pool %v", tx)
		}
	}

	// Delete batch since it is finished
	err = k.DeleteBatch(ctx, *batch)
	if err != nil {
		return err
	}

	// Delete its confirmations as well
	err = k.DeleteBatchConfirms(ctx, *batch)
	if err != nil {
		return err
	}

	bridgeContract, err := k.GetBridgeContractAddress(ctx)
	if err != nil {
		return err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.EventManager().EmitTypedEvent(
		&types.EventOutgoingBatchCanceled{
			BridgeContract: bridgeContract.GetAddress().Hex(),
			BridgeChainId:  strconv.Itoa(int(k.GetBridgeChainID(ctx))),
			BatchId:        string(types.GetOutgoingTxBatchKey(tokenContract, nonce)),
			Nonce:          fmt.Sprint(nonce),
		},
	)
}

// IterateOutgoingTxBatches iterates through all outgoing batches in ASC order.
func (k Keeper) IterateOutgoingTxBatches(ctx context.Context, cb func(key []byte, batch types.InternalOutgoingTxBatch) bool) error {
	prefixStore := prefix.NewStore(k.GetStore(ctx), types.OutgoingTXBatchKey)
	iter := prefixStore.ReverseIterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var batch types.OutgoingTxBatch
		k.cdc.MustUnmarshal(iter.Value(), &batch)
		intBatch, err := batch.ToInternal()
		if err != nil || intBatch == nil {
			return sdkerrors.Wrap(err, "found invalid batch in store")
		}
		// cb returns true to stop early
		if cb(iter.Key(), *intBatch) {
			break
		}
	}
	return nil
}

// GetOutgoingTxBatches returns the outgoing tx batches
func (k Keeper) GetOutgoingTxBatches(ctx context.Context) (out []types.InternalOutgoingTxBatch, err error) {
	err = k.IterateOutgoingTxBatches(ctx, func(_ []byte, batch types.InternalOutgoingTxBatch) bool {
		out = append(out, batch)
		return false
	})
	return
}

func (k Keeper) GetOutgoingTxBatchesByNonce(ctx context.Context) (map[uint64]types.InternalOutgoingTxBatch, error) {
	batchesByNonce := make(map[uint64]types.InternalOutgoingTxBatch)
	var g whoops.Group
	g.Add(
		k.IterateOutgoingTxBatches(ctx, func(_ []byte, batch types.InternalOutgoingTxBatch) bool {
			if _, exists := batchesByNonce[batch.BatchNonce]; exists {
				g.Add(fmt.Errorf("batch with duplicate batch nonce %d in store", batch.BatchNonce))
				return true
			}
			batchesByNonce[batch.BatchNonce] = batch
			return false
		}),
	)
	if len(g) > 0 {
		return nil, g
	}
	return batchesByNonce, nil
}

// GetLastOutgoingBatchByTokenType gets the latest outgoing tx batch by token type
func (k Keeper) GetLastOutgoingBatchByTokenType(ctx context.Context, token types.EthAddress) (*types.InternalOutgoingTxBatch, error) {
	batches, err := k.GetOutgoingTxBatches(ctx)
	if err != nil {
		return nil, err
	}
	var lastBatch *types.InternalOutgoingTxBatch = nil
	lastNonce := uint64(0)
	for i, batch := range batches {
		if batch.TokenContract.GetAddress() == token.GetAddress() && batch.BatchNonce > lastNonce {
			lastBatch = &batches[i]
			lastNonce = batch.BatchNonce
		}
	}
	return lastBatch, nil
}

// HasLastSlashedBatchBlock returns true if the last slashed batch block has been set in the store
func (k Keeper) HasLastSlashedBatchBlock(ctx context.Context) bool {
	store := k.GetStore(ctx)
	return store.Has(types.LastSlashedBatchBlock)
}

// SetLastSlashedBatchBlock sets the latest slashed Batch block height this is done by
// block height instead of nonce because batches could have individual nonces for each token type
func (k Keeper) SetLastSlashedBatchBlock(ctx context.Context, blockHeight uint64) error {
	if k.HasLastSlashedBatchBlock(ctx) {
		lastSlashedBatchBlock, err := k.GetLastSlashedBatchBlock(ctx)
		if err != nil {
			return err
		}
		if lastSlashedBatchBlock > blockHeight {
			return fmt.Errorf("attempted to decrement LastSlashedBatchBlock")
		}
	}

	store := k.GetStore(ctx)
	store.Set(types.LastSlashedBatchBlock, types.UInt64Bytes(blockHeight))
	return nil
}

// GetLastSlashedBatchBlock returns the latest slashed Batch block
func (k Keeper) GetLastSlashedBatchBlock(ctx context.Context) (uint64, error) {
	store := k.GetStore(ctx)
	bytes := store.Get(types.LastSlashedBatchBlock)

	if len(bytes) == 0 {
		return 0, fmt.Errorf("last slashed batch block not initialized from genesis")
	}
	return types.UInt64FromBytes(bytes)
}

// GetUnSlashedBatches returns all the unslashed batches in state
func (k Keeper) GetUnSlashedBatches(ctx context.Context, maxHeight uint64) (out []types.InternalOutgoingTxBatch, err error) {
	lastSlashedBatchBlock, err := k.GetLastSlashedBatchBlock(ctx)
	if err != nil {
		return nil, err
	}
	batches, err := k.GetOutgoingTxBatches(ctx)
	for _, batch := range batches {
		if batch.PalomaBlockCreated > lastSlashedBatchBlock && batch.PalomaBlockCreated < maxHeight {
			out = append(out, batch)
		}
	}
	return
}
