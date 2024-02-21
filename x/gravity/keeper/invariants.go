package keeper

import (
	"fmt"
	"sort"

	"cosmossdk.io/math"
	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/gravity/types"
)

/*	Gravity Module Invariants
	For background: https://docs.cosmos.network/main/building-modules/invariants
	Invariants on Gravity Bridge chain will be enforced by most validators every 200 blocks, see module/cmd/root.go for
	the automatic configuration. These settings are overrideable and not consensus breaking so there are no firm
	guarantees of invariant checking no matter what is put here.
*/

// AllInvariants collects any defined invariants below
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := StoreValidityInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		return ModuleBalanceInvariant(k)(ctx)

		/*
			Example additional invariants:
			res, stop := FutureInvariant(k)(ctx)
			if stop {
				return res, stop
			}
			return AnotherFutureInvariant(k)(ctx)
		*/
	}
}

// ModuleBalanceInvariant checks that the module account's balance is equal to the balance of unbatched transactions and unobserved batches
// Note that the returned bool should be true if there is an error, e.g. an unexpected module balance
func ModuleBalanceInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		modAcc := k.accountKeeper.GetModuleAddress(types.ModuleName)
		actualBals := k.bankKeeper.GetAllBalances(ctx, modAcc)
		expectedBals := make(map[string]*math.Int, len(actualBals)) // Collect balances by contract
		for _, v := range actualBals {
			newInt := math.NewInt(0)
			expectedBals[v.Denom] = &newInt
		}
		expectedBals, err := sumUnconfirmedBatchModuleBalances(ctx, k, expectedBals)
		if err != nil {
			panic(err)
		}
		expectedBals, err = sumUnbatchedTxModuleBalances(ctx, k, expectedBals)
		if err != nil {
			panic(err)
		}

		// Compare actual vs expected balances
		for _, actual := range actualBals {
			denom := actual.GetDenom()
			_, err := k.GetERC20OfDenom(ctx, "test-chain", denom)
			if err != nil {
				// Here we do not return because a user could halt the chain by gifting gravity a cosmos asset with no erc20 repr
				liblog.FromSDKLogger(k.Logger(sdkCtx)).WithFields("asset", denom).Error("Unexpected gravity module balance of paloma-originated asset with no erc20 representation")
				continue
			}
			expected, ok := expectedBals[denom]
			if !ok {
				return fmt.Sprint("Could not find expected balance for actual module balance of ", actual), true
			}

			if !actual.Amount.Equal(*expected) {
				return fmt.Sprint("Mismatched balance of eth-originated ", denom, ": actual balance ", actual.Amount, " != expected balance ", expected), true
			}
		}
		return "", false
	}
}

/////// MODULE BALANCE HELPERS

// sumUnconfirmedBatchModuleBalances calculate the value the module should have stored due to unconfirmed batches
func sumUnconfirmedBatchModuleBalances(ctx sdk.Context, k Keeper, expectedBals map[string]*math.Int) (map[string]*math.Int, error) {
	err := k.IterateOutgoingTxBatches(ctx, func(_ []byte, batch types.InternalOutgoingTxBatch) bool {
		batchTotal := math.NewInt(0)
		// Collect the send amount for each tx
		for _, tx := range batch.Transactions {
			newTotal := batchTotal.Add(tx.Erc20Token.Amount)
			batchTotal = newTotal
		}
		contract := batch.TokenContract
		denom, _ := k.GetDenomOfERC20(ctx, batch.ChainReferenceID, contract)
		// Add the batch total to the contract counter
		_, ok := expectedBals[denom]
		if !ok {
			zero := math.ZeroInt()
			expectedBals[denom] = &zero
		}

		*expectedBals[denom] = expectedBals[denom].Add(batchTotal)

		return false // continue iterating
	})

	return expectedBals, err
}

// sumUnbatchedTxModuleBalances calculates the value the module should have stored due to unbatched txs
func sumUnbatchedTxModuleBalances(ctx sdk.Context, k Keeper, expectedBals map[string]*math.Int) (map[string]*math.Int, error) {
	// It is also given the balance of all unbatched txs in the pool
	err := k.IterateUnbatchedTransactions(ctx, func(_ []byte, tx *types.InternalOutgoingTransferTx) bool {
		contract := tx.Erc20Token.Contract
		denom, _ := k.GetDenomOfERC20(ctx, tx.Erc20Token.ChainReferenceID, contract)

		_, ok := expectedBals[denom]
		if !ok {
			zero := math.ZeroInt()
			expectedBals[denom] = &zero
		}
		*expectedBals[denom] = expectedBals[denom].Add(tx.Erc20Token.Amount)

		return false // continue iterating
	})
	if err != nil {
		return nil, err
	}
	return expectedBals, nil
}

// StoreValidityInvariant checks that the currently stored objects are not corrupted and all pass ValidateBasic checks
// Note that the returned bool should be true if there is an error, e.g. an unexpected batch was processed
func StoreValidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		err := ValidateStore(ctx, k)
		if err != nil {
			return err.Error(), true
		}
		// Assert that batch metadata is consistent and expected
		err = CheckBatches(ctx, k)
		if err != nil {
			return err.Error(), true
		}

		// SUCCESS: If execution made it here, everything passes the sanity checks
		return "", false
	}
}

// STORE VALIDITY HELPERS

// ValidateStore checks that all values in the store can be decoded and pass a ValidateBasic check
// Returns an error string and a boolean indicating an error if true, for use in an invariant
// nolint: gocyclo
func ValidateStore(ctx sdk.Context, k Keeper) error {
	// EthAddressByValidatorKey
	var g whoops.Group
	var err error

	// LastObservedEventNonceKey (type checked when fetching)
	lastObservedEventNonce, err := k.GetLastObservedEventNonce(ctx)
	if err != nil {
		return err
	}

	lastObservedEthereumClaimHeight := uint64(0) // used later to validate ethereum claim height
	// OracleAttestationKey
	g.Add(
		k.IterateAttestations(ctx, false, func(key []byte, att types.Attestation) (stop bool) {
			er := att.ValidateBasic(k.cdc)
			if er != nil {
				g.Add(fmt.Errorf("Invalid attestation %v in IterateAttestations: %v", att, er))
				return true
			}
			claim, er := k.UnpackAttestationClaim(&att) // Already unpacked in ValidateBasic
			if er != nil {
				g.Add(fmt.Errorf("Invalid attestation claim %v in IterateAttestations: %v", att, er))
				return true
			}
			if att.Observed {
				if claim.GetEventNonce() > lastObservedEventNonce {
					g.Add(fmt.Errorf("last observed event nonce <> observed attestation nonce mismatch (%v < %v)", lastObservedEventNonce, claim.GetEventNonce()))
					return true
				}
				claimHeight := claim.GetEthBlockHeight()
				if claimHeight > lastObservedEthereumClaimHeight {
					lastObservedEthereumClaimHeight = claimHeight
				}
			}
			return false
		}),
	)
	if len(g) > 0 {
		return g
	}

	// OutgoingTXPoolKey
	g.Add(
		k.IterateUnbatchedTransactions(ctx, func(key []byte, tx *types.InternalOutgoingTransferTx) (stop bool) {
			err = tx.ValidateBasic()
			if err != nil {
				g.Add(fmt.Errorf("Invalid unbatched transaction %v under key %v in IterateUnbatchedTransactions: %v", tx, key, err))
				return true
			}
			return false
		}),
	)
	if len(g) > 0 {
		return g
	}

	// OutgoingTXBatchKey
	g.Add(
		k.IterateOutgoingTxBatches(ctx, func(key []byte, batch types.InternalOutgoingTxBatch) (stop bool) {
			err = batch.ValidateBasic()
			if err != nil {
				g.Add(fmt.Errorf("Invalid outgoing batch %v under key %v in IterateOutgoingTxBatches: %v", batch, key, err))
				return true
			}
			return false
		}),
	)
	if len(g) > 0 {
		return g
	}

	// BatchConfirmKey
	k.IterateBatchConfirms(ctx, func(key []byte, confirm types.MsgConfirmBatch) (stop bool) {
		err = confirm.ValidateBasic()
		if err != nil {
			err = fmt.Errorf("Invalid batch confirm %v under key %v in IterateBatchConfirms: %v", confirm, key, err)
			return true
		}
		return false
	})
	if err != nil {
		return err
	}
	// LastEventNonceByValidatorKey (type checked when fetching)
	err = k.IterateValidatorLastEventNonces(ctx, func(key []byte, nonce uint64) (stop bool) {
		return false
	})
	if err != nil {
		return err
	}
	// SequenceKeyPrefix HAS BEEN REMOVED

	// KeyLastTXPoolID (type checked when fetching)
	_, err = k.getID(ctx, types.KeyLastTXPoolID)
	if err != nil {
		return err
	}
	// KeyLastOutgoingBatchID (type checked when fetching)
	_, err = k.getID(ctx, types.KeyLastOutgoingBatchID)
	if err != nil {
		return err
	}
	// LastObservedEthereumBlockHeightKey
	lastEthHeight := k.GetLastObservedEthereumBlockHeight(ctx)
	if lastEthHeight.EthereumBlockHeight < lastObservedEthereumClaimHeight {
		err = fmt.Errorf(
			"Stored last observed ethereum block height is less than the actual last observed height (%d < %d)",
			lastEthHeight.EthereumBlockHeight,
			lastObservedEthereumClaimHeight,
		)
	}
	if err != nil {
		return err
	}
	// DenomToERC20Key
	allErc20ToDenoms, err := k.GetAllERC20ToDenoms(ctx)
	if err != nil {
		return err
	}
	for _, erc20 := range allErc20ToDenoms {
		if err = erc20.ValidateBasic(); err != nil {
			return fmt.Errorf("Discovered invalid erc20 %v for %s:%s: %v", erc20, erc20.GetChainReferenceId(), erc20.GetDenom(), err)
		}
	}

	// ERC20ToDenomKey
	allDenomToERC20s, err := k.GetAllDenomToERC20s(ctx)
	if err != nil {
		return err
	}
	for _, erc20 := range allDenomToERC20s {
		if err = erc20.ValidateBasic(); err != nil {
			return fmt.Errorf("Discovered invalid erc20 %v for %s:%s: %v", erc20, erc20.GetChainReferenceId(), erc20.GetDenom(), err)
		}
	}

	// LastSlashedBatchBlock (type is checked when fetching)
	_, err = k.GetLastSlashedBatchBlock(ctx)
	if err != nil {
		return err
	}

	// PastEthSignatureCheckpointKey
	err = k.IteratePastEthSignatureCheckpoints(ctx, func(key []byte, value []byte) (stop bool) {
		// Check is performed in the iterator function
		return false
	})
	if err != nil {
		return err
	}

	// Finally the params, which are not placed in the store
	params := k.GetParams(ctx)
	err = params.ValidateBasic()
	if err != nil {
		return fmt.Errorf("Discovered invalid params %v: %v", params, err)
	}

	return nil
}

// CheckBatches checks that all batch related data in the store is appropriate
// Returns an error string and a boolean indicating an error if true, for use in an invariant
func CheckBatches(ctx sdk.Context, k Keeper) error {
	// Get the in-progress batches by the batch nonces
	inProgressBatches, err := k.GetOutgoingTxBatchesByNonce(ctx)
	if err != nil {
		return err
	}
	if len(inProgressBatches) == 0 {
		return nil // nothing to check against
	}
	// Sort the nonces
	var sortedNonces []uint64
	for k := range inProgressBatches {
		sortedNonces = append(sortedNonces, k)
	}
	sort.Slice(sortedNonces, func(i int, j int) bool { return sortedNonces[i] < sortedNonces[j] })
	// Now we can make assertions about the ordered batches
	for i, nonce := range sortedNonces {
		if i == 0 {
			continue // skip the first
		}
		lastBatch := inProgressBatches[sortedNonces[i-1]]
		batch := inProgressBatches[nonce]

		// Batches with lower nonces should not have been created in more-recent cosmos blocks
		if lastBatch.PalomaBlockCreated > batch.PalomaBlockCreated {
			return fmt.Errorf(
				"old batch created (nonce=%d created=%d) after new batch (nonce=%d created=%d)",
				lastBatch.BatchNonce, lastBatch.PalomaBlockCreated, batch.BatchNonce, batch.PalomaBlockCreated,
			)
		}
	}

	var g whoops.Group
	g.Add(
		k.IterateClaims(ctx, true, types.CLAIM_TYPE_BATCH_SEND_TO_ETH, func(key []byte, att types.Attestation, claim types.EthereumClaim) (stop bool) {
			batchClaim := claim.(*types.MsgBatchSendToEthClaim)
			// Executed (aka observed) batches should have strictly lesser batch nonces than the in progress batches for the same token contract
			// note that batches for different tokens have the same nonce stream but don't invalidate each other (nonces should probably be separate per token type)
			if att.Observed {
				for _, val := range inProgressBatches {
					if batchClaim.BatchNonce >= val.BatchNonce && batchClaim.TokenContract == val.TokenContract.GetAddress().String() {
						g.Add(fmt.Errorf("in-progress batches have incorrect nonce, should be > %d", batchClaim.BatchNonce))
						return true
					}
				}
			}
			return false
		}),
	)
	if len(g) > 0 {
		return g
	}
	return nil
}
