package skyway

import (
	"context"
	"errors"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/libcons"
	"github.com/palomachain/paloma/v2/util/liblog"
	"github.com/palomachain/paloma/v2/util/slice"
	"github.com/palomachain/paloma/v2/x/skyway/keeper"
	"github.com/palomachain/paloma/v2/x/skyway/types"
	valsettypes "github.com/palomachain/paloma/v2/x/valset/types"
)

const updateValidatorNoncesPeriod = 50

// EndBlocker is called at the end of every block
func EndBlocker(ctx context.Context, k keeper.Keeper, cc *libcons.ConsensusChecker) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	logger := liblog.FromKeeper(ctx, k).WithComponent("skyway-endblocker")
	defer func() {
		if r := recover(); r != nil {
			logger.WithFields("original-error", r).Warn("Recovered panic.")
		}
	}()
	logger.Debug("Running endblocker")

	// slashing(ctx, k)
	chains := k.EVMKeeper.GetActiveChainNames(ctx)

	err := createBatch(ctx, k)
	if err != nil {
		logger.WithError(err).Warn("Failed to create batches")
	}

	for _, v := range chains {
		err = attestationTally(ctx, k, v)
		if err != nil {
			logger.WithError(err).Warn("Failed to tally attestations.")
		}

		err = pruneAttestations(ctx, k, v)
		if err != nil {
			logger.WithError(err).Warn("Failed to prune attestations.")
		}

		err = slashUnattestingValidators(ctx, k, v)
		if err != nil {
			logger.WithError(err).Warn("Failed to slash unattesting validators.")
		}

		if sdkCtx.BlockHeight()%updateValidatorNoncesPeriod == 0 {
			// Update all validator nonces to the latest observed nonce, if
			// suitable This makes sure validators don't get stuck waiting for
			// events they can't get from the RPC, even though these events are
			// already observed
			err = k.UpdateValidatorNoncesToLatest(ctx, v)
			if err != nil {
				logger.WithError(err).Warn("Failed to update validator nonces.")
			}
		}
	}

	err = processGasEstimates(ctx, k, cc)
	if err != nil {
		logger.WithError(err).Warn("Failed to process gas estimates.")
	}

	err = cleanupTimedOutBatches(ctx, k)
	if err != nil {
		logger.WithError(err).Warn("Failed to cleanup timed out batches.")
	}
}

// lce is a light data structure to
// implementing libcons.Estimate
type lce struct {
	addr sdk.ValAddress
	v    uint64
}

func (l lce) GetValAddress() sdk.ValAddress { return l.addr }
func (l lce) GetValue() uint64              { return l.v }

// Iterates over all outgoing batches and builds a gas estimate for each one
// if enough estimates have been collected.
func processGasEstimates(ctx context.Context, k keeper.Keeper, cc *libcons.ConsensusChecker) error {
	err := k.IterateOutgoingTxBatches(ctx, func(key []byte, batch types.InternalOutgoingTxBatch) bool {
		// Skip batches that already have a gas estimate
		if batch.GasEstimate > 0 {
			return false
		}
		logger := liblog.FromKeeper(ctx, k).
			WithComponent("skyway-process-gas-estimates").
			WithFields(
				"batch-nonce", batch.BatchNonce,
				"token-contract", batch.TokenContract,
				"chain-reference-id", batch.ChainReferenceID,
			)
		logger.Debug("Processing gas estimates for batch")
		batchEstimates, err := k.GetBatchGasEstimateByNonceAndTokenContract(ctx, batch.BatchNonce, batch.TokenContract)
		if err != nil {
			logger.WithError(err).Warn("Failed to get gas estimates")
			return false
		}
		if len(batchEstimates) < 1 {
			logger.Debug("Skipping batch, not enough gas estimates")
			return false
		}

		lc := make([]lce, len(batchEstimates))
		for i, estimate := range batchEstimates {
			addr, err := sdk.AccAddressFromBech32(estimate.Metadata.Creator)
			if err != nil {
				logger.WithFields(
					"validator", estimate.Metadata.Creator,
				).WithError(err).
					Warn("Failed to get validator address from estimate")
				return false
			}
			lc[i] = lce{sdk.ValAddress(addr), estimate.Estimate}
		}

		// Try to get consensus on the gas estimate
		estimate, err := cc.VerifyGasEstimates(ctx, k,
			slice.Map(lc, func(ge lce) libcons.GasEstimate {
				return ge
			}))
		if err != nil {
			if errors.Is(err, libcons.ErrConsensusNotAchieved) {
				logger.WithError(err).Debug("Consensus not achieved for gas estimates")
				return false
			}
			err = fmt.Errorf("failed to verify gas estimates: %w", err)
			logger.WithError(err).Warn("Failed to verify gas estimates")
			return false
		}

		if err := k.UpdateBatchGasEstimate(ctx, batch, estimate); err != nil {
			logger.WithError(err).Warn("Failed to set gas estimate")
		}

		return false
	})
	if err != nil {
		return fmt.Errorf("failed to iterate outgoing tx batches: %w", err)
	}
	return nil
}

// Create a batch of transactions for each token
func createBatch(ctx context.Context, k keeper.Keeper) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight()%50 == 0 {
		denoms, err := k.GetAllERC20ToDenoms(ctx)
		if err != nil {
			return err
		}
		for _, erc20ToDenom := range denoms {
			tokenContract, err := k.GetERC20OfDenom(ctx, erc20ToDenom.GetChainReferenceId(), erc20ToDenom.Denom)
			if err != nil {
				return err
			}

			_, err = k.BuildOutgoingTXBatch(ctx, erc20ToDenom.ChainReferenceId, *tokenContract, keeper.OutgoingTxBatchSize)
			if err != nil {
				return sdkerrors.Wrap(err, "Could not build outgoing tx batch")
			}
		}
	}
	return nil
}

// Iterate over all attestations currently being voted on in order of nonce and
// "Observe" those who have passed the threshold. Break the loop once we see
// an attestation that has not passed the threshold
func attestationTally(ctx context.Context, k keeper.Keeper, chainReferenceID string) error {
	attmap, keys, err := k.GetAttestationMapping(ctx, chainReferenceID)
	if err != nil {
		return err
	}

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
			lastEventNonce, err := k.GetLastObservedSkywayNonce(ctx, chainReferenceID)
			if err != nil {
				return err
			}

			if nonce == uint64(lastEventNonce)+1 {
				err := k.TryAttestation(ctx, &att) //nolint:gosec
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// cleanupTimedOutBatches deletes batches that have passed their expiration
func cleanupTimedOutBatches(ctx context.Context, k keeper.Keeper) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := uint64(sdkCtx.BlockTime().Unix())
	batches, err := k.GetOutgoingTxBatches(ctx)
	if err != nil {
		return err
	}
	for _, batch := range batches {
		if batch.BatchTimeout < currentTime {
			err := k.CancelOutgoingTXBatch(ctx, batch.TokenContract, batch.BatchNonce)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to cancel outgoing txbatch")
			}
		}
	}
	return nil
}

// Iterate over all attestations currently being voted on in order of nonce
// and prune those that are older than the current nonce and no longer have any
// use. This could be combined with create attestation and save some computation
// but (A) pruning keeps the iteration small in the first place and (B) there is
// already enough nuance in the other handler that it's best not to complicate it further
func pruneAttestations(ctx context.Context, k keeper.Keeper, chainReferenceID string) error {
	attmap, keys, err := k.GetAttestationMapping(ctx, chainReferenceID)
	if err != nil {
		return err
	}

	// we delete all attestations earlier than the current event nonce
	// minus some buffer value. This buffer value is purely to allow
	// frontends and other UI components to view recent oracle history
	const eventsToKeep = 1000
	lastNonce, err := k.GetLastObservedSkywayNonce(ctx, chainReferenceID)
	if err != nil {
		return err
	}

	var cutoff uint64
	if lastNonce <= eventsToKeep {
		return nil
	} else {
		cutoff = lastNonce - eventsToKeep
	}

	// This iterates over all keys (event nonces) in the attestation mapping. Each value contains
	// a slice with one or more attestations at that event nonce. There can be multiple attestations
	// at one event nonce when validators disagree about what event happened at that nonce.
	for _, nonce := range keys {
		// This iterates over all attestations at a particular event nonce.
		// They are ordered by when the first attestation at the event nonce was received.
		// This order is not important.
		for _, att := range attmap[nonce] {
			// delete all before the cutoff
			if nonce < cutoff {
				err := k.DeleteAttestation(ctx, chainReferenceID, att)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func slashUnattestingValidators(ctx context.Context, k keeper.Keeper, chainReferenceID string) error {
	at, err := k.GetMostRecentAttestations(ctx, chainReferenceID, 1)
	if err != nil {
		return fmt.Errorf("failed to get most recent attestations: %w", err)
	}
	if len(at) != 1 {
		return fmt.Errorf("expected 1 attestation, got %d", len(at))
	}

	if at[0].Observed {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if (uint64(sdkCtx.BlockHeight())-at[0].Height)%300 != 0 {
		// Not enough time has passed since the last attestation, so we don't slash
		return nil
	}

	voterLkup := map[string]struct{}{}
	for _, v := range at[0].Votes {
		voterLkup[v] = struct{}{}
	}

	vals, err := k.StakingKeeper.GetBondedValidatorsByPower(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bonded validators: %w", err)
	}

	for _, v := range vals {
		if v.IsJailed() {
			continue
		}

		if _, ok := voterLkup[v.GetOperator()]; !ok {
			// This validator did not vote on the attestation, so we slash them
			valAddr, err := sdk.ValAddressFromBech32(v.GetOperator())
			if err != nil {
				return fmt.Errorf("failed to parse validator address %s: %w", v.GetOperator(), err)
			}

			k.ValsetKeeper.Jail(ctx, valAddr, valsettypes.JailReasonMissedBridgeAttestation)
			if err != nil {
				return fmt.Errorf("failed to jail validator %s: %w", valAddr, err)
			}
		}
	}

	return nil
}
