package gravity

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/keeper"
	log "github.com/sirupsen/logrus"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx context.Context, k keeper.Keeper) {
	// slashing(ctx, k)

	err := createBatch(ctx, k)
	if err != nil {
		log.Error(err)
	}

	err = attestationTally(ctx, k)
	if err != nil {
		log.Error(err)
	}

	err = cleanupTimedOutBatches(ctx, k)
	if err != nil {
		log.Error(err)
	}

	err = pruneAttestations(ctx, k)
	if err != nil {
		log.Error(err)
	}
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
func attestationTally(ctx context.Context, k keeper.Keeper) error {
	attmap, keys, err := k.GetAttestationMapping(ctx)
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
			lastEventNonce, err := k.GetLastObservedEventNonce(ctx)
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
func pruneAttestations(ctx context.Context, k keeper.Keeper) error {
	attmap, keys, err := k.GetAttestationMapping(ctx)
	if err != nil {
		return err
	}

	// we delete all attestations earlier than the current event nonce
	// minus some buffer value. This buffer value is purely to allow
	// frontends and other UI components to view recent oracle history
	const eventsToKeep = 1000
	lastNonce, err := k.GetLastObservedEventNonce(ctx)
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
				err := k.DeleteAttestation(ctx, att)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
