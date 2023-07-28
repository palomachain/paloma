package keeper

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

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
		modAcc := k.accountKeeper.GetModuleAddress(types.ModuleName)
		actualBals := k.bankKeeper.GetAllBalances(ctx, modAcc)
		expectedBals := make(map[string]*sdk.Int, len(actualBals)) // Collect balances by contract
		for _, v := range actualBals {
			newInt := sdk.NewInt(0)
			expectedBals[v.Denom] = &newInt
		}
		expectedBals = sumUnconfirmedBatchModuleBalances(ctx, k, expectedBals)
		expectedBals = sumUnbatchedTxModuleBalances(ctx, k, expectedBals)
		expectedBals = sumPendingIbcAutoForwards(ctx, k, expectedBals)

		// Compare actual vs expected balances
		for _, actual := range actualBals {
			denom := actual.GetDenom()
			cosmosOriginated, _, err := k.DenomToERC20Lookup(ctx, denom)
			if err != nil {
				// Here we do not return because a user could halt the chain by gifting gravity a cosmos asset with no erc20 repr
				ctx.Logger().Error("Unexpected gravity module balance of cosmos-originated asset with no erc20 representation", "asset", denom)
				continue
			}
			expected, ok := expectedBals[denom]
			if !ok {
				return fmt.Sprint("Could not find expected balance for actual module balance of ", actual), true
			}

			if cosmosOriginated { // Cosmos originated mismatched balance
				// We cannot make any assertions about cosmosOriginated assets because we do not have enough information.
				// There is no index of denom => amount bridged, which would force us to parse all logs in existence
			} else if !actual.Amount.Equal(*expected) { // Eth originated mismatched balance
				return fmt.Sprint("Mismatched balance of eth-originated ", denom, ": actual balance ", actual.Amount, " != expected balance ", expected), true
			}
		}
		return "", false
	}
}

/////// MODULE BALANCE HELPERS

// sumUnconfirmedBatchModuleBalances calculate the value the module should have stored due to unconfirmed batches
func sumUnconfirmedBatchModuleBalances(ctx sdk.Context, k Keeper, expectedBals map[string]*sdk.Int) map[string]*sdk.Int {
	k.IterateOutgoingTxBatches(ctx, func(_ []byte, batch types.InternalOutgoingTxBatch) bool {
		batchTotal := sdk.NewInt(0)
		// Collect the send amount + fee amount for each tx
		for _, tx := range batch.Transactions {
			newTotal := batchTotal.Add(tx.Erc20Token.Amount.Add(tx.Erc20Fee.Amount))
			batchTotal = newTotal
		}
		contract := batch.TokenContract
		_, denom := k.ERC20ToDenomLookup(ctx, contract)
		// Add the batch total to the contract counter
		_, ok := expectedBals[denom]
		if !ok {
			zero := sdk.ZeroInt()
			expectedBals[denom] = &zero
		}

		*expectedBals[denom] = expectedBals[denom].Add(batchTotal)

		return false // continue iterating
	})

	return expectedBals
}

// sumUnbatchedTxModuleBalances calculates the value the module should have stored due to unbatched txs
func sumUnbatchedTxModuleBalances(ctx sdk.Context, k Keeper, expectedBals map[string]*sdk.Int) map[string]*sdk.Int {
	// It is also given the balance of all unbatched txs in the pool
	k.IterateUnbatchedTransactions(ctx, func(_ []byte, tx *types.InternalOutgoingTransferTx) bool {
		contract := tx.Erc20Token.Contract
		_, denom := k.ERC20ToDenomLookup(ctx, contract)

		// Collect the send amount + fee amount for each tx
		txTotal := tx.Erc20Token.Amount.Add(tx.Erc20Fee.Amount)
		_, ok := expectedBals[denom]
		if !ok {
			zero := sdk.ZeroInt()
			expectedBals[denom] = &zero
		}
		*expectedBals[denom] = expectedBals[denom].Add(txTotal)

		return false // continue iterating
	})

	return expectedBals
}

func sumPendingIbcAutoForwards(ctx sdk.Context, k Keeper, expectedBals map[string]*sdk.Int) map[string]*sdk.Int {
	for _, forward := range k.PendingIbcAutoForwards(ctx, uint64(0)) {
		if _, ok := expectedBals[forward.Token.Denom]; !ok {
			zero := sdk.ZeroInt()
			expectedBals[forward.Token.Denom] = &zero
		} else {
			*expectedBals[forward.Token.Denom] = expectedBals[forward.Token.Denom].Add(forward.Token.Amount)
		}
	}

	return expectedBals
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

		// Assert that valsets have been updated in the expected manner
		err = CheckValsets(ctx, k)
		if err != nil {
			return err.Error(), true
		}

		// Assert that pending ibc auto-forwards are only outgoing
		err = CheckPendingIbcAutoForwards(ctx, k)
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
	// TODO: Check the newly added iterators with unit tests
	// EthAddressByValidatorKey
	var err error = nil
	k.IterateEthAddressesByValidator(ctx, func(key []byte, addr types.EthAddress) (stop bool) {
		err = addr.ValidateBasic()
		if err != nil {
			err = fmt.Errorf("Failed to validate eth address %v under key %v in IterateEthAddressesByValidator: %s", addr, key, err.Error())
			return true
		}
		return false
	})
	if err != nil {
		return err
	}
	// ValidatorByEthAddressKey
	k.IterateValidatorsByEthAddress(ctx, func(key []byte, addr sdk.ValAddress) (stop bool) {
		bech := addr.String()
		if len(bech) == 0 || !strings.HasPrefix(bech, sdk.GetConfig().GetBech32ValidatorAddrPrefix()) {
			err = fmt.Errorf("Invalid validator %v under key %v in IterateValidatorsByEthAddress", addr, key)
			return true
		}
		return false
	})
	if err != nil {
		return err
	}
	// ValsetRequestKey
	var actualLatestValsetNonce uint64 = 0
	k.IterateValsets(ctx, func(key []byte, valset *types.Valset) (stop bool) {
		err = valset.ValidateBasic()
		if err != nil {
			err = fmt.Errorf("Invalid valset %v in IterateValsets: %v", valset, err)
			return true
		}
		nonce := valset.Nonce
		if nonce > actualLatestValsetNonce {
			actualLatestValsetNonce = nonce
		}
		return false
	})
	if err != nil {
		return err
	}
	// ValsetConfirmKey
	k.IterateValsetConfirms(ctx, func(key []byte, confirms []types.MsgValsetConfirm, nonce uint64) (stop bool) {
		for _, confirm := range confirms {
			err = confirm.ValidateBasic()
			if err != nil {
				err = fmt.Errorf("Invalid MsgValsetConfirm %v for nonce %d in IterateValsetConfirms: %v", confirm, nonce, err)
				return true
			}
			if confirm.Nonce != nonce {
				panic(fmt.Errorf("Unexpected msg Nonce %d in IterateValsetConfirms, expected %d", confirm.Nonce, nonce))
			}
		}
		return false
	})
	if err != nil {
		return err
	}
	// OracleClaimKey HAS BEEN REMOVED

	// LastObservedEventNonceKey (type checked when fetching)
	lastObservedEventNonce := k.GetLastObservedEventNonce(ctx)
	lastObservedEthereumClaimHeight := uint64(0) // used later to validate ethereum claim height
	// OracleAttestationKey
	k.IterateAttestations(ctx, false, func(key []byte, att types.Attestation) (stop bool) {
		er := att.ValidateBasic(k.cdc)
		if er != nil {
			err = fmt.Errorf("Invalid attestation %v in IterateAttestations: %v", att, er)
			return true
		}
		claim, er := k.UnpackAttestationClaim(&att) // Already unpacked in ValidateBasic
		if er != nil {
			err = fmt.Errorf("Invalid attestation claim %v in IterateAttestations: %v", att, er)
			return true
		}
		if att.Observed {
			if claim.GetEventNonce() > lastObservedEventNonce {
				err = fmt.Errorf("last observed event nonce <> observed attestation nonce mismatch (%v < %v)", lastObservedEventNonce, claim.GetEventNonce())
				return true
			}
			claimHeight := claim.GetEthBlockHeight()
			if claimHeight > lastObservedEthereumClaimHeight {
				lastObservedEthereumClaimHeight = claimHeight
			}
		}
		return false
	})
	if err != nil {
		return err
	}
	// OutgoingTXPoolKey
	k.IterateUnbatchedTransactions(ctx, func(key []byte, tx *types.InternalOutgoingTransferTx) (stop bool) {
		err = tx.ValidateBasic()
		if err != nil {
			err = fmt.Errorf("Invalid unbatched transaction %v under key %v in IterateUnbatchedTransactions: %v", tx, key, err)
			return true
		}
		return false
	})
	if err != nil {
		return err
	}
	// OutgoingTXBatchKey
	k.IterateOutgoingTxBatches(ctx, func(key []byte, batch types.InternalOutgoingTxBatch) (stop bool) {
		err = batch.ValidateBasic()
		if err != nil {
			err = fmt.Errorf("Invalid outgoing batch %v under key %v in IterateOutgoingTxBatches: %v", batch, key, err)
			return true
		}
		return false
	})
	if err != nil {
		return err
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
	k.IterateValidatorLastEventNonces(ctx, func(key []byte, nonce uint64) (stop bool) {
		return false
	})
	if err != nil {
		return err
	}
	// SequenceKeyPrefix HAS BEEN REMOVED

	// KeyLastTXPoolID (type checked when fetching)
	_ = k.getID(ctx, types.KeyLastTXPoolID)
	// KeyLastOutgoingBatchID (type checked when fetching)
	_ = k.getID(ctx, types.KeyLastOutgoingBatchID)
	// KeyOrchestratorAddress
	k.IterateValidatorsByOrchestratorAddress(ctx, func(key []byte, addr sdk.ValAddress) (stop bool) {
		bech := addr.String()
		if len(bech) == 0 || !strings.HasPrefix(bech, sdk.GetConfig().GetBech32ValidatorAddrPrefix()) {
			err = fmt.Errorf("Invalid validator %v under key %v in IterateValidatorsByOrchestratorAddress", addr, key)
			return true
		}
		return false
	})
	if err != nil {
		return err
	}
	// KeyOutgoingLogicCall
	k.IterateOutgoingLogicCalls(ctx, func(key []byte, logicCall types.OutgoingLogicCall) (stop bool) {
		err = logicCall.ValidateBasic()
		if err != nil {
			err = fmt.Errorf("Invalid logic call %v under key %v in IterateOutgoingLogicCalls: %v", logicCall, key, err)
			return true
		}
		return false
	})
	if err != nil {
		return err
	}
	// KeyOutgoingLogicConfirm
	k.IterateLogicConfirms(ctx, func(key []byte, confirm *types.MsgConfirmLogicCall) (stop bool) {
		err = confirm.ValidateBasic()
		if err != nil {
			err = fmt.Errorf("Invalid logic call confirm %v under key %v in IterateLogicConfirms: %v", confirm, key, err)
			return true
		}
		return false
	})
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
	k.IterateCosmosOriginatedERC20s(ctx, func(key []byte, erc20 *types.EthAddress) (stop bool) {
		if err = erc20.ValidateBasic(); err != nil {
			err = fmt.Errorf("Discovered invalid cosmos originated erc20 %v under key %v: %v", erc20, key, err)
			return true
		}
		return false
	})
	if err != nil {
		return err
	}
	// ERC20ToDenomKey
	k.IterateERC20ToDenom(ctx, func(key []byte, erc20ToDenom *types.ERC20ToDenom) (stop bool) {
		if err = erc20ToDenom.ValidateBasic(); err != nil {
			err = fmt.Errorf("Discovered invalid ERC20ToDenom %v under key %v: %v", erc20ToDenom, key, err)
			return true
		}
		return false
	})
	if err != nil {
		return err
	}
	// LastSlashedValsetNonce (type is checked when fetching)
	_ = k.GetLastSlashedValsetNonce(ctx)

	// LatestValsetNonce
	latestValsetNonce := k.GetLatestValsetNonce(ctx)
	if latestValsetNonce != actualLatestValsetNonce {
		return fmt.Errorf(
			"GetLatestValsetNonce <> max(IterateValsets' nonce) mismatch (%d != %d)",
			latestValsetNonce,
			actualLatestValsetNonce,
		)
	}
	// LastSlashedBatchBlock (type is checked when fetching)
	_ = k.GetLastSlashedBatchBlock(ctx)

	// LastSlashedLogicCallBlock (type is checked when fetching)
	_ = k.GetLastSlashedLogicCallBlock(ctx)

	// LastUnBondingBlockHeight (type is checked when fetching)
	_ = k.GetLastUnBondingBlockHeight(ctx)

	// LastObservedValsetKey
	valset := k.GetLastObservedValset(ctx)

	if valset != nil {
		err = valset.ValidateBasic()
	}
	if err != nil {
		return fmt.Errorf("Discovered invalid last observed valset %v: %v", valset, err)
	}
	// PastEthSignatureCheckpointKey
	k.IteratePastEthSignatureCheckpoints(ctx, func(key []byte, value []byte) (stop bool) {
		// Check is performed in the iterator function
		return false
	})

	// PendingIbcAutoForwards
	k.IteratePendingIbcAutoForwards(ctx, func(key []byte, forward *types.PendingIbcAutoForward) (stop bool) {
		if err = forward.ValidateBasic(); err != nil {
			err = fmt.Errorf("Discovered invalid PendingIbcAutoForward %v under key %v: %v", forward, key, err)
			return true
		}
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
	inProgressBatches := k.GetOutgoingTxBatchesByNonce(ctx)
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
		if lastBatch.CosmosBlockCreated > batch.CosmosBlockCreated {
			return fmt.Errorf(
				"old batch created (nonce=%d created=%d) after new batch (nonce=%d created=%d)",
				lastBatch.BatchNonce, lastBatch.CosmosBlockCreated, batch.BatchNonce, batch.CosmosBlockCreated,
			)
		}
	}

	var err error = nil
	k.IterateClaims(ctx, true, types.CLAIM_TYPE_BATCH_SEND_TO_ETH, func(key []byte, att types.Attestation, claim types.EthereumClaim) (stop bool) {
		batchClaim := claim.(*types.MsgBatchSendToEthClaim)
		// Executed (aka observed) batches should have strictly lesser batch nonces than the in progress batches for the same token contract
		// note that batches for different tokens have the same nonce stream but don't invalidate each other (nonces should probably be separate per token type)
		if att.Observed {
			for _, val := range inProgressBatches {
				if batchClaim.BatchNonce >= val.BatchNonce && batchClaim.TokenContract == val.TokenContract.GetAddress().String() {
					err = fmt.Errorf("in-progress batches have incorrect nonce, should be > %d", batchClaim.BatchNonce)
					return true
				}
			}
		}
		return false
	})
	return err
}

// CheckValsets checks that all valset related data in the store is appropriate
// Returns an error string and a boolean indicating an error if true, for use in an invariant
func CheckValsets(ctx sdk.Context, k Keeper) error {
	valsets := k.GetValsets(ctx) // Highest to lowest nonce
	if len(valsets) == 0 {
		if k.GetLatestValsetNonce(ctx) != 0 {
			return fmt.Errorf("No valsets in store but the latest valset nonce is nonzero!")
		}
		return nil
	}
	latest := k.GetLatestValset(ctx) // Should have the highest nonce
	equal, err := latest.Equal(valsets[0])
	if err != nil {
		return fmt.Errorf("Unable to compare latest valsets: %s", err.Error())
	}
	if !equal {
		return fmt.Errorf("Latest stored valset (%v) is unexpectedly different from GetLatestValset() (%v)", valsets[0], latest)
	}
	// Should have power diff of less than 0.05 between the current valset and the last stored one
	current, err := k.GetCurrentValset(ctx)
	if err != nil {
		return fmt.Errorf("Unable to retrieve current valsets: %s", err.Error())
	}
	_, err = types.BridgeValidators(current.Members).ToInternal()
	if err != nil {
		return fmt.Errorf("Unable to make current BridgeValidators from the current valset: %s", err.Error())
	}
	_, err = types.BridgeValidators(latest.Members).ToInternal()
	if err != nil {
		return fmt.Errorf("Unable to make latest BridgeValidators from the latest valset: %s", err.Error())
	}

	// The previously stored valsets may have been created for multiple reasons, so we make no more power diff checks

	// Check the stored valsets against the observed valset update attestations
	k.IterateClaims(ctx, true, types.CLAIM_TYPE_VALSET_UPDATED, func(key []byte, att types.Attestation, claim types.EthereumClaim) (stop bool) {
		if !att.Observed {
			// This claim is in-progress, malicious, or erroneous - continue to the next one
			return false
		}
		claimVs := claim.(*types.MsgValsetUpdatedClaim) // Claim's valset
		claimNonce := claimVs.ValsetNonce
		storedVs := k.GetValset(ctx, claimNonce)
		if storedVs == nil {
			// We have found the earliest stored valset, all valsets before it should have been pruned
			shouldNotExist := k.GetValset(ctx, claimNonce-1)
			if shouldNotExist != nil {
				err = fmt.Errorf("Discovered a valset missing from the store: nonce %d is present but nonce %d is not", shouldNotExist.Nonce, claimNonce)
				return true
			}
			// Didn't exist, stop checking so we can return no error
			return true
		}
		// we found the valset for the claim, check that the valset powers match
		var currBvs types.BridgeValidators = storedVs.Members
		currIntBvs, er := currBvs.ToInternal()
		if er != nil {
			err = fmt.Errorf("invalid bridge validators in store for valset nonce %d", storedVs.Nonce)
			return true
		}
		var claimBvs types.BridgeValidators = claimVs.Members
		claimIntBvs, er := claimBvs.ToInternal()
		if er != nil {
			err = fmt.Errorf("invalid valset updated claim in store for event nonce %d, valset nonce %d", claimVs.EventNonce, claimVs.ValsetNonce)
			return true
		}
		if currIntBvs.PowerDiff(*claimIntBvs) > 0.0001 {
			err = fmt.Errorf("power difference discovered between stored valset %v and observed attestation valset %v", storedVs, claimVs)
			return true
		}
		return false // continue
	})

	return err
}

// CheckPendingIbcAutoForwards checks each forward is appropriate and also that the transfer module holds enough of each token
func CheckPendingIbcAutoForwards(ctx sdk.Context, k Keeper) error {
	nativeHrp := sdk.GetConfig().GetBech32AccountAddrPrefix()
	pendingForwards := k.PendingIbcAutoForwards(ctx, 0)
	for _, fwd := range pendingForwards {
		// Check the foreign address
		hrp, _, err := bech32.DecodeAndConvert(fwd.ForeignReceiver)
		if err != nil {
			return fmt.Errorf("found invalid pending IBC auto forward (%v) in the store: %v", fwd, err)
		}
		if hrp == nativeHrp {
			return fmt.Errorf("found invalid pending IBC auto forward (%v) with local receiving address %s", fwd, fwd.ForeignReceiver)
		}
		// Check the amount
		if !fwd.Token.Amount.IsPositive() {
			return fmt.Errorf("found pending IBC auto forward (%v) transferring a non-positive balance %s", fwd, fwd.Token.Amount.String())
		}
		// Check the denom and account balances
		fwdDenom := fwd.Token.Denom
		if strings.HasPrefix(strings.ToLower(fwdDenom), "ibc/") {
			_, err := k.ibcTransferKeeper.DenomPathFromHash(ctx, fwdDenom)
			if err != nil {
				return fmt.Errorf("Unable to parse path from ibc denom %s: %v", fwdDenom, err)
			}
		}
	}

	return nil
}
