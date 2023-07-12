package gravity

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/palomachain/paloma/x/gravity/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
)

// BeginBlocker is called at the beginning of every block
// NOTE: begin blocker also emits events which are helpful for
// clients listening to the chain and creating transactions
// based on the events (i.e. orchestrators)
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	cleanupTimedOutBatchTxs(ctx, k)
	cleanupTimedOutContractCallTxs(ctx, k)
	createSignerSetTxs(ctx, k)
	createBatchTxs(ctx, k)
	pruneSignerSetTxs(ctx, k)
}

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	outgoingTxSlashing(ctx, k)
	eventVoteRecordTally(ctx, k)
	updateObservedEthereumHeight(ctx, k)
}

func createBatchTxs(ctx sdk.Context, k keeper.Keeper) {
	// TODO: this needs some more work, is super naieve
	if ctx.BlockHeight()%10 == 0 {
		cm := map[string]bool{}
		k.IterateUnbatchedSendToEthereums(ctx, func(ste *types.SendToEthereum) bool {
			cm[ste.Erc20Token.Contract] = true
			return false
		})

		var contracts []string
		for k := range cm {
			contracts = append(contracts, k)
		}
		sort.Strings(contracts)

		for _, c := range contracts {
			// NOTE: this doesn't emit events which would be helpful for client processes
			k.CreateBatchTx(ctx, common.HexToAddress(c), 100)
		}
	}
}

func createSignerSetTxs(ctx sdk.Context, k keeper.Keeper) {
	// Auto signerset tx creation.
	// 1. If there are no signer set requests, create a new one.
	// 2. If there is at least one validator who started unbonding in current block. (we persist last unbonded block height in hooks.go)
	//      This will make sure the unbonding validator has to provide an ethereum signature to a new signer set tx
	//	    that excludes him before he completely Unbonds.  Otherwise he will be slashed
	// 3. If power change between validators of Current signer set and latest signer set request is > 5%
	latestSignerSetTx := k.GetLatestSignerSetTx(ctx)
	if latestSignerSetTx == nil {
		k.CreateSignerSetTx(ctx)
		return
	}

	lastUnbondingHeight := k.GetLastUnbondingBlockHeight(ctx)
	blockHeight := uint64(ctx.BlockHeight())
	powerDiff := types.EthereumSigners(k.CurrentSignerSet(ctx)).PowerDiff(latestSignerSetTx.Signers)

	shouldCreate := (lastUnbondingHeight == blockHeight) || (powerDiff > 0.05)
	k.Logger(ctx).Info(
		"considering signer set tx creation",
		"blockHeight", blockHeight,
		"lastUnbondingHeight", lastUnbondingHeight,
		"latestSignerSetTx.Nonce", latestSignerSetTx.Nonce,
		"powerDiff", powerDiff,
		"shouldCreate", shouldCreate,
	)

	if shouldCreate {
		k.CreateSignerSetTx(ctx)
	}
}

func pruneSignerSetTxs(ctx sdk.Context, k keeper.Keeper) {
	params := k.GetParams(ctx)
	// Validator set pruning
	// prune all validator sets with a nonce less than the
	// last observed nonce, they can't be submitted any longer
	//
	// Only prune valsets after the signed valsets window has passed
	// so that slashing can occur the block before we remove them
	lastObserved := k.GetLastObservedSignerSetTx(ctx)
	currentBlock := uint64(ctx.BlockHeight())
	tooEarly := currentBlock < params.SignedSignerSetTxsWindow
	if lastObserved != nil && !tooEarly {
		earliestToPrune := currentBlock - params.SignedSignerSetTxsWindow
		for _, set := range k.GetSignerSetTxs(ctx) {
			if set.Nonce < lastObserved.Nonce && set.Height < earliestToPrune {
				k.DeleteOutgoingTx(ctx, set.GetStoreIndex())
			}
		}
	}
}

// Iterate over all attestations currently being voted on in order of nonce and
// "Observe" those who have passed the threshold. Break the loop once we see
// an attestation that has not passed the threshold
func eventVoteRecordTally(ctx sdk.Context, k keeper.Keeper) {
	attmap := k.GetEthereumEventVoteRecordMapping(ctx)

	// We make a slice with all the event nonces that are in the attestation mapping
	keys := make([]uint64, 0, len(attmap))
	for k := range attmap {
		keys = append(keys, k)
	}
	// Then we sort it
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	// This iterates over all keys (event nonces) in the attestation mapping. Each value contains
	// a slice with one or more attestations at that event nonce. There can be multiple attestations
	// at one event nonce when validators disagree about what event happened at that nonce.
	for _, nonce := range keys {
		// This iterates over all attestations at a particular event nonce.
		// They are ordered by when the first attestation at the event nonce was received.
		// This order is not important.
		for _, att := range attmap[nonce] {
			// We check if the event nonce is exactly 1 higher than the last attestation that was
			// observed. If it is not, we just move on to the next nonce. This will skip over all
			// attestations that have already been observed.
			//
			// Once we hit an event nonce that is one higher than the last observed event, we stop
			// skipping over this conditional and start calling tryAttestation (counting votes)
			// Once an attestation at a given event nonce has enough votes and becomes observed,
			// every other attestation at that nonce will be skipped, since the lastObservedEventNonce
			// will be incremented.
			//
			// Then we go to the next event nonce in the attestation mapping, if there is one. This
			// nonce will once again be one higher than the lastObservedEventNonce.
			// If there is an attestation at this event nonce which has enough votes to be observed,
			// we skip the other attestations and move on to the next nonce again.
			// If no attestation becomes observed, when we get to the next nonce, every attestation in
			// it will be skipped. The same will happen for every nonce after that.
			if nonce == uint64(k.GetLastObservedEventNonce(ctx))+1 {
				k.TryEventVoteRecord(ctx, att)
			}
		}
	}
}

// Periodically, every orchestrator will submit their latest observed Ethereum and Cosmos heights in
// order to keep this information current regardless of the level of bridge activity.
//
// We determine if we should update the latest heights based on the following criteria:
//  1. A consensus of validators agrees that the proposed height is equal to or less than their
//     last observed height, in order to reconcile the many different heights that will be submitted.
//     The highest height that meets this criteria will be the proposed height.
//  2. The proposed consensus heights from this process are greater than the values stored from the last time
//     we observed an Ethereum event from the bridge
func updateObservedEthereumHeight(ctx sdk.Context, k keeper.Keeper) {
	// wait some minutes before checking the height votes
	if ctx.BlockHeight()%50 != 0 {
		return
	}

	ethereumHeightPowers := make(map[uint64]sdk.Int)
	cosmosHeightPowers := make(map[uint64]sdk.Int)
	// we can use the same value as event vote records for this threshold
	requiredPower := types.EventVoteRecordPowerThreshold(k.StakingKeeper.GetLastTotalPower(ctx))

	// populate the list
	k.IterateEthereumHeightVotes(ctx, func(valAddres sdk.ValAddress, height types.LatestEthereumBlockHeight) bool {
		if _, ok := ethereumHeightPowers[height.EthereumHeight]; !ok {
			ethereumHeightPowers[height.EthereumHeight] = sdk.NewInt(0)
		}

		if _, ok := cosmosHeightPowers[height.CosmosHeight]; !ok {
			cosmosHeightPowers[height.CosmosHeight] = sdk.NewInt(0)
		}

		return false
	})

	// vote on acceptable height values (less than or equal to the validator's observed value)
	k.IterateEthereumHeightVotes(ctx, func(valAddress sdk.ValAddress, height types.LatestEthereumBlockHeight) bool {
		validatorPower := sdk.NewInt(k.StakingKeeper.GetLastValidatorPower(ctx, valAddress))

		for ethereumVoteHeight, ethereumPower := range ethereumHeightPowers {
			if ethereumVoteHeight <= height.EthereumHeight {
				ethereumHeightPowers[ethereumVoteHeight] = ethereumPower.Add(validatorPower)
			}
		}

		for cosmosVoteHeight, cosmosPower := range cosmosHeightPowers {
			if cosmosVoteHeight <= height.CosmosHeight {
				cosmosHeightPowers[cosmosVoteHeight] = cosmosPower.Add(validatorPower)
			}
		}

		return false
	})

	// find the highest height submitted that a consensus of validators agreed was acceptable
	ethereumHeight := uint64(0)
	cosmosHeight := uint64(0)

	for ethereumVoteHeight, ethereumPower := range ethereumHeightPowers {
		if ethereumVoteHeight > ethereumHeight && ethereumPower.GTE(requiredPower) {
			ethereumHeight = ethereumVoteHeight
		}
	}

	for cosmosVoteHeight, cosmosPower := range cosmosHeightPowers {
		if cosmosVoteHeight > cosmosHeight && cosmosPower.GTE(requiredPower) {
			cosmosHeight = cosmosVoteHeight
		}
	}

	lastObservedHeights := k.GetLastObservedEthereumBlockHeight(ctx)
	if ethereumHeight > lastObservedHeights.EthereumHeight && cosmosHeight > lastObservedHeights.CosmosHeight {
		k.SetLastObservedEthereumBlockHeightWithCosmos(ctx, ethereumHeight, cosmosHeight)
	}
}

// cleanupTimedOutBatchTxs deletes batches that have passed their expiration on Ethereum
// keep in mind several things when modifying this function
// A) unlike nonces timeouts are not monotonically increasing, meaning batch 5 can have a later timeout than batch 6
// this means that we MUST only cleanup a single batch at a time
//
// B) it is possible for ethereumHeight to be zero if no events have ever occurred, make sure your code accounts for this
// C) When we compute the timeout we do our best to estimate the Ethereum block height at that very second. But what we work with
//
//	here is the Ethereum block height at the time of the last Deposit or Withdraw to be observed. It's very important we do not
//	project, if we do a slowdown on ethereum could cause a double spend. Instead timeouts will *only* occur after the timeout period
//	AND any deposit or withdraw has occurred to update the Ethereum block height.
func cleanupTimedOutBatchTxs(ctx sdk.Context, k keeper.Keeper) {
	ethereumHeight := k.GetLastObservedEthereumBlockHeight(ctx).EthereumHeight
	k.IterateOutgoingTxsByType(ctx, types.BatchTxPrefixByte, func(key []byte, otx types.OutgoingTx) bool {
		btx, _ := otx.(*types.BatchTx)

		if btx.Timeout < ethereumHeight {
			k.CancelBatchTx(ctx, btx)
		}

		return false
	})
}

// cleanupTimedOutContractCallTxs deletes logic calls that have passed their expiration on Ethereum
// keep in mind several things when modifying this function
// A) unlike nonces timeouts are not monotonically increasing, meaning call 5 can have a later timeout than batch 6
//
//	this means that we MUST only cleanup a single call at a time
//
// B) it is possible for ethereumHeight to be zero if no events have ever occurred, make sure your code accounts for this
// C) When we compute the timeout we do our best to estimate the Ethereum block height at that very second. But what we work with
//
//	here is the Ethereum block height at the time of the last Deposit or Withdraw to be observed. It's very important we do not
//	project, if we do a slowdown on ethereum could cause a double spend. Instead timeouts will *only* occur after the timeout period
//	AND any deposit or withdraw has occurred to update the Ethereum block height.
func cleanupTimedOutContractCallTxs(ctx sdk.Context, k keeper.Keeper) {
	ethereumHeight := k.GetLastObservedEthereumBlockHeight(ctx).EthereumHeight
	k.IterateOutgoingTxsByType(ctx, types.ContractCallTxPrefixByte, func(_ []byte, otx types.OutgoingTx) bool {
		cctx, _ := otx.(*types.ContractCallTx)
		if cctx.Timeout < ethereumHeight {
			k.DeleteOutgoingTx(ctx, cctx.GetStoreIndex())
		}
		return true
	})
}

func outgoingTxSlashing(ctx sdk.Context, k keeper.Keeper) {
	params := k.GetParams(ctx)
	maxHeight := uint64(0)
	if uint64(ctx.BlockHeight()) > params.SignedBatchesWindow {
		maxHeight = uint64(ctx.BlockHeight()) - params.SignedBatchesWindow
	} else {
		return
	}

	usotxs := k.GetUnSlashedOutgoingTxs(ctx, maxHeight)
	if len(usotxs) == 0 {
		return
	}

	// get signing info for each validator
	type valInfo struct {
		val   stakingtypes.Validator
		exist bool
		sigs  slashingtypes.ValidatorSigningInfo
		cons  sdk.ConsAddress
	}

	bondedVals := k.StakingKeeper.GetBondedValidatorsByPower(ctx)
	valInfos := make([]valInfo, len(bondedVals))

	for i, val := range bondedVals {
		consAddr, err := val.GetConsAddr()
		if err != nil {
			panic(fmt.Sprintf("failed to get consensus address: %s", err))
		}

		sigs, exist := k.SlashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
		valInfos[i] = valInfo{val, exist, sigs, consAddr}
	}

	var unbondingValInfos []valInfo

	blockTime := ctx.BlockTime().Add(k.StakingKeeper.GetParams(ctx).UnbondingTime)
	blockHeight := ctx.BlockHeight()
	unbondingValIterator := k.StakingKeeper.ValidatorQueueIterator(ctx, blockTime, blockHeight)
	defer unbondingValIterator.Close()

	// All unbonding validators
	for ; unbondingValIterator.Valid(); unbondingValIterator.Next() {
		unbondingValidators := k.GetUnbondingvalidators(unbondingValIterator.Value())
		for _, valAddr := range unbondingValidators.Addresses {
			addr, err := sdk.ValAddressFromBech32(valAddr)
			if err != nil {
				panic(fmt.Sprintf("failed to bech32 decode validator address: %s", err))
			}

			validator, _ := k.StakingKeeper.GetValidator(ctx, addr)

			valConsAddr, err := validator.GetConsAddr()
			if err != nil {
				panic(fmt.Sprintf("failed to get validator consensus address: %s", err))
			}

			valSigningInfo, exist := k.SlashingKeeper.GetValidatorSigningInfo(ctx, valConsAddr)
			unbondingValInfos = append(unbondingValInfos, valInfo{validator, exist, valSigningInfo, valConsAddr})
		}
	}

	for _, otx := range usotxs {
		// SLASH BONDED VALIDATORS who didn't sign batch txs
		signatures := k.GetEthereumSignatures(ctx, otx.GetStoreIndex())
		for _, valInfo := range valInfos {
			// Don't slash validators who joined after outgoingtx is created
			if valInfo.exist && valInfo.sigs.StartHeight < int64(otx.GetCosmosHeight()) {
				if _, ok := signatures[valInfo.val.GetOperator().String()]; !ok {
					if !valInfo.val.IsJailed() {
						power := valInfo.val.ConsensusPower(k.PowerReduction)
						k.StakingKeeper.Slash(
							ctx,
							valInfo.cons,
							ctx.BlockHeight(),
							power,
							params.SlashFractionBatch,
						)
						k.StakingKeeper.Jail(ctx, valInfo.cons)

						ctx.EventManager().EmitEvent(
							sdk.NewEvent(
								slashingtypes.EventTypeSlash,
								sdk.NewAttribute(slashingtypes.AttributeKeyAddress, valInfo.cons.String()),
								sdk.NewAttribute(slashingtypes.AttributeKeyJailed, valInfo.cons.String()),
								sdk.NewAttribute(slashingtypes.AttributeKeyReason, types.AttributeMissingBridgeBatchSig),
								sdk.NewAttribute(slashingtypes.AttributeKeyPower, fmt.Sprintf("%d", power)),
							),
						)
					}
				}
			}
		}

		if sstx, ok := otx.(*types.SignerSetTx); ok {
			for _, valInfo := range unbondingValInfos {
				// Only slash validators who joined after valset is created and they are
				// unbonding and UNBOND_SLASHING_WINDOW didn't pass.
				if valInfo.exist && valInfo.sigs.StartHeight < int64(sstx.Height) &&
					valInfo.val.IsUnbonding() &&
					sstx.Height < uint64(valInfo.val.UnbondingHeight)+params.UnbondSlashingSignerSetTxsWindow {
					// check if validator has confirmed valset or not
					if _, found := signatures[valInfo.val.GetOperator().String()]; !found {
						if !valInfo.val.IsJailed() {
							// TODO: Do we want to slash jailed validators?
							power := valInfo.val.ConsensusPower(k.PowerReduction)
							k.StakingKeeper.Slash(
								ctx,
								valInfo.cons,
								ctx.BlockHeight(),
								power,
								params.SlashFractionSignerSetTx,
							)
							k.StakingKeeper.Jail(ctx, valInfo.cons)

							ctx.EventManager().EmitEvent(
								sdk.NewEvent(
									slashingtypes.EventTypeSlash,
									sdk.NewAttribute(slashingtypes.AttributeKeyAddress, valInfo.cons.String()),
									sdk.NewAttribute(slashingtypes.AttributeKeyJailed, valInfo.cons.String()),
									sdk.NewAttribute(slashingtypes.AttributeKeyReason, types.AttributeMissingBridgeBatchSig),
									sdk.NewAttribute(slashingtypes.AttributeKeyPower, fmt.Sprintf("%d", power)),
								),
							)
						}
					}
				}
			}
		}

		// then we set the latest slashed outgoing tx block
		k.SetLastSlashedOutgoingTxBlockHeight(ctx, otx.GetCosmosHeight())
	}
}
