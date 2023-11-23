package keeper

import (
	"fmt"
	"sort"
	"strconv"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/VolumeFi/whoops"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
)

func (k Keeper) Attest(
	ctx sdk.Context,
	claim types.EthereumClaim,
	anyClaim *codectypes.Any,
) (*types.Attestation, error) {
	val, found, err := k.GetOrchestratorValidator(ctx, claim.GetClaimer())
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("could not find ValAddr for delegate key, should be checked by now")
	}
	valAddr := val.GetOperator()
	if err := sdk.VerifyAddressFormat([]byte(valAddr)); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid orchestrator validator address")
	}
	// Check that the nonce of this event is exactly one higher than the last nonce stored by this validator.
	// We check the event nonce in processAttestation as well,
	// but checking it here gives individual eth signers a chance to retry,
	// and prevents validators from submitting two claims with the same nonce.
	// This prevents there being two attestations with the same nonce that get 2/3s of the votes
	// in the endBlocker.
	lastEventNonce, err := k.GetLastEventNonceByValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}
	if claim.GetEventNonce() != lastEventNonce+1 {
		return nil, fmt.Errorf(types.ErrNonContiguousEventNonce.Error(), lastEventNonce+1, claim.GetEventNonce())
	}

	// Tries to get an attestation with the same eventNonce and claim as the claim that was submitted.
	hash, err := claim.ClaimHash()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to compute claim hash")
	}
	att := k.GetAttestation(ctx, claim.GetEventNonce(), hash)

	// If it does not exist, create a new one.
	if att == nil {
		att = &types.Attestation{
			Observed: false,
			Votes:    []string{},
			Height:   uint64(ctx.BlockHeight()),
			Claim:    anyClaim,
		}
	}

	ethClaim, err := k.UnpackAttestationClaim(att)
	if err != nil {
		return nil, fmt.Errorf("could not unpack stored attestation claim, %v", err)
	}

	if ethClaim.GetEthBlockHeight() == claim.GetEthBlockHeight() {

		// Add the validator's vote to this attestation // TODO : Only do if it's not already there
		att.Votes = append(att.Votes, valAddr.String())

		k.SetAttestation(ctx, claim.GetEventNonce(), hash, att)
		err = k.SetLastEventNonceByValidator(ctx, valAddr, claim.GetEventNonce())
		if err != nil {
			return nil, err
		}

		return att, nil
	} else {
		return nil, fmt.Errorf("invalid height - this claim's height is %v while the stored height is %v", claim.GetEthBlockHeight(), ethClaim.GetEthBlockHeight())
	}
}

// TryAttestation checks if an attestation has enough votes to be applied to the consensus state
// and has not already been marked Observed, then calls processAttestation to actually apply it to the state,
// and then marks it Observed and emits an event.
func (k Keeper) TryAttestation(ctx sdk.Context, att *types.Attestation) error {
	claim, err := k.UnpackAttestationClaim(att)
	if err != nil {
		return fmt.Errorf("could not cast to claim")
	}
	hash, err := claim.ClaimHash()
	if err != nil {
		return fmt.Errorf("unable to compute claim hash")
	}
	// If the attestation has not yet been Observed, sum up the votes and see if it is ready to apply to the state.
	// This conditional stops the attestation from accidentally being applied twice.
	if !att.Observed {
		// Sum the current powers of all validators who have voted and see if it passes the current threshold
		// TODO: The different integer types and math here needs a careful review
		totalPower := k.StakingKeeper.GetLastTotalPower(ctx)
		requiredPower := types.AttestationVotesPowerThreshold.Mul(totalPower).Quo(math.NewInt(100))
		attestationPower := math.NewInt(0)
		for _, validator := range att.Votes {
			val, err := sdk.ValAddressFromBech32(validator)
			if err != nil {
				return err
			}
			validatorPower := k.StakingKeeper.GetLastValidatorPower(ctx, val)
			// Add it to the attestation power's sum
			attestationPower = attestationPower.Add(math.NewInt(validatorPower))
			// If the power of all the validators that have voted on the attestation is higher or equal to the threshold,
			// process the attestation, set Observed to true, and break
			if attestationPower.GT(requiredPower) {
				lastEventNonce, err := k.GetLastObservedEventNonce(ctx)
				if err != nil {
					return err
				}

				if claim.GetEventNonce() != lastEventNonce+1 {
					return fmt.Errorf("attempting to apply events to state out of order")
				}

				err = k.setLastObservedEventNonce(ctx, claim.GetEventNonce())
				if err != nil {
					return err
				}

				err = k.SetLastObservedEthereumBlockHeight(ctx, claim.GetEthBlockHeight())
				if err != nil {
					return err
				}

				att.Observed = true
				k.SetAttestation(ctx, claim.GetEventNonce(), hash, att)

				err = k.processAttestation(ctx, att, claim)
				if err != nil {
					return err
				}

				err = k.emitObservedEvent(ctx, att, claim)
				if err != nil {
					return err
				}

				break
			}
		}
	} else {
		// We error here because this should never happen
		return fmt.Errorf("attempting to process observed attestation")
	}
	return nil
}

// processAttestation actually applies the attestation to the consensus state
func (k Keeper) processAttestation(ctx sdk.Context, att *types.Attestation, claim types.EthereumClaim) error {
	hash, err := claim.ClaimHash()
	if err != nil {
		return fmt.Errorf("unable to compute claim hash")
	}
	// then execute in a new Tx so that we can store state on failure
	xCtx, commit := ctx.CacheContext()
	if err := k.AttestationHandler.Handle(xCtx, *att, claim); err != nil { // execute with a transient storage
		// If the attestation fails, something has gone wrong and we can't recover it. Log and move on
		// The attestation will still be marked "Observed", allowing the oracle to progress properly
		k.Logger(ctx).Error("attestation failed",
			"cause", err.Error(),
			"claim type", claim.GetType(),
			"id", types.GetAttestationKey(claim.GetEventNonce(), hash),
			"nonce", fmt.Sprint(claim.GetEventNonce()),
		)
	} else {
		commit() // persist transient storage
	}
	return nil
}

// emitObservedEvent emits an event with information about an attestation that has been applied to
// consensus state.
func (k Keeper) emitObservedEvent(ctx sdk.Context, att *types.Attestation, claim types.EthereumClaim) error {
	hash, err := claim.ClaimHash()
	if err != nil {
		return sdkerrors.Wrap(err, "unable to compute claim hash")
	}

	bridgeContract, err := k.GetBridgeContractAddress(ctx)
	if err != nil {
		return err
	}

	return ctx.EventManager().EmitTypedEvent(
		&types.EventObservation{
			AttestationType: string(claim.GetType()),
			BridgeContract:  bridgeContract.GetAddress().Hex(),
			BridgeChainId:   strconv.Itoa(int(k.GetBridgeChainID(ctx))),
			AttestationId:   string(types.GetAttestationKey(claim.GetEventNonce(), hash)),
			Nonce:           fmt.Sprint(claim.GetEventNonce()),
		},
	)
}

// SetAttestation sets the attestation in the store
func (k Keeper) SetAttestation(ctx sdk.Context, eventNonce uint64, claimHash []byte, att *types.Attestation) {
	store := k.GetStore(ctx)
	aKey := types.GetAttestationKey(eventNonce, claimHash)
	store.Set(aKey, k.cdc.MustMarshal(att))
}

// GetAttestation return an attestation given a nonce
func (k Keeper) GetAttestation(ctx sdk.Context, eventNonce uint64, claimHash []byte) *types.Attestation {
	store := k.GetStore(ctx)
	aKey := types.GetAttestationKey(eventNonce, claimHash)
	bz := store.Get(aKey)
	if len(bz) == 0 {
		return nil
	}
	var att types.Attestation
	k.cdc.MustUnmarshal(bz, &att)
	return &att
}

// DeleteAttestation deletes the given attestation
func (k Keeper) DeleteAttestation(ctx sdk.Context, att types.Attestation) error {
	claim, err := k.UnpackAttestationClaim(&att)
	if err != nil {
		return fmt.Errorf("bad attestation in DeleteAttestation")
	}
	hash, err := claim.ClaimHash()
	if err != nil {
		return sdkerrors.Wrap(err, "unable to compute claim hash")
	}
	store := k.GetStore(ctx)

	store.Delete(types.GetAttestationKey(claim.GetEventNonce(), hash))
	return nil
}

// GetAttestationMapping returns a mapping of eventnonce -> attestations at that nonce
// it also returns a pre-sorted array of the keys, this assists callers of this function
// by providing a deterministic iteration order. You should always iterate over ordered keys
// if you are iterating this map at all.
func (k Keeper) GetAttestationMapping(ctx sdk.Context) (attestationMapping map[uint64][]types.Attestation, orderedKeys []uint64, err error) {
	attestationMapping = make(map[uint64][]types.Attestation)
	var g whoops.Group
	g.Add(
		k.IterateAttestations(ctx, false, func(_ []byte, att types.Attestation) bool {
			claim, err := k.UnpackAttestationClaim(&att)
			if err != nil {
				g.Add(err)
				return true
			}

			if val, ok := attestationMapping[claim.GetEventNonce()]; !ok {
				attestationMapping[claim.GetEventNonce()] = []types.Attestation{att}
			} else {
				attestationMapping[claim.GetEventNonce()] = append(val, att)
			}
			return false
		}),
	)
	if len(g) > 0 {
		return attestationMapping, orderedKeys, g
	}
	orderedKeys = make([]uint64, 0, len(attestationMapping))
	for k := range attestationMapping {
		orderedKeys = append(orderedKeys, k)
	}
	sort.Slice(orderedKeys, func(i, j int) bool { return orderedKeys[i] < orderedKeys[j] })

	return
}

// IterateAttestations iterates through all attestations executing a given callback on each discovered attestation
// If reverse is true, attestations will be returned in descending order by key (aka by event nonce and then claim hash)
// cb should return true to stop iteration, false to continue
func (k Keeper) IterateAttestations(ctx sdk.Context, reverse bool, cb func(key []byte, att types.Attestation) (stop bool)) error {
	store := k.GetStore(ctx)
	keyPrefix := types.OracleAttestationKey
	start, end, err := prefixRange(keyPrefix)
	if err != nil {
		return err
	}

	var iter storetypes.Iterator
	if reverse {
		iter = store.ReverseIterator(start, end)
	} else {
		iter = store.Iterator(start, end)
	}
	defer func(iter storetypes.Iterator) {
		err := iter.Close()
		if err != nil {
			panic("Unable to close attestation iterator!")
		}
	}(iter)

	for ; iter.Valid(); iter.Next() {
		att := types.Attestation{
			Observed: false,
			Votes:    []string{},
			Height:   0,
			Claim: &codectypes.Any{
				TypeUrl:              "",
				Value:                []byte{},
				XXX_NoUnkeyedLiteral: struct{}{},
				XXX_unrecognized:     []byte{},
				XXX_sizecache:        0,
			},
		}
		k.cdc.MustUnmarshal(iter.Value(), &att)
		// cb returns true to stop early
		if cb(iter.Key(), att) {
			return nil
		}
	}
	return nil
}

// IterateClaims iterates through all attestations, filtering them for claims of a given type
// If reverse is true, attestations will be returned in descending order by key (aka by event nonce and then claim hash)
// cb should return true to stop iteration, false to continue
func (k Keeper) IterateClaims(ctx sdk.Context, reverse bool, claimType types.ClaimType, cb func(key []byte, att types.Attestation, claim types.EthereumClaim) (stop bool)) error {
	typeUrl := types.ClaimTypeToTypeUrl(claimType) // Used to avoid unpacking undesired attestations

	var g whoops.Group
	g.Add(
		k.IterateAttestations(ctx, reverse, func(key []byte, att types.Attestation) bool {
			if att.Claim.TypeUrl == typeUrl {
				claim, err := k.UnpackAttestationClaim(&att)
				if err != nil {
					g.Add(fmt.Errorf("Discovered invalid claim in attestation %v under key %v: %v", att, key, err))
					return true
				}

				return cb(key, att, claim)
			}
			return false
		}),
	)
	if len(g) > 0 {
		return g
	}
	return nil
}

// GetMostRecentAttestations returns sorted (by nonce) attestations up to a provided limit number of attestations
// Note: calls GetAttestationMapping in the hopes that there are potentially many attestations
// which are distributed between few nonces to minimize sorting time
func (k Keeper) GetMostRecentAttestations(ctx sdk.Context, limit uint64) ([]types.Attestation, error) {
	attestationMapping, keys, err := k.GetAttestationMapping(ctx)
	if err != nil {
		return nil, err
	}
	attestations := make([]types.Attestation, 0, limit)

	// Iterate the nonces and collect the attestations
	count := 0
	for _, nonce := range keys {
		if count >= int(limit) {
			break
		}
		for _, att := range attestationMapping[nonce] {
			if count >= int(limit) {
				break
			}
			attestations = append(attestations, att)
			count++
		}
	}

	return attestations, nil
}

// GetLastObservedEventNonce returns the latest observed event nonce
func (k Keeper) GetLastObservedEventNonce(ctx sdk.Context) (uint64, error) {
	store := k.GetStore(ctx)
	bytes := store.Get(types.LastObservedEventNonceKey)

	if len(bytes) == 0 {
		return 0, nil
	}
	if len(bytes) > 8 {
		return 0, fmt.Errorf("last observed event nonce is not a uint64")
	}
	return types.UInt64FromBytes(bytes)
}

// GetLastObservedEthereumBlockHeight height gets the block height to of the last observed attestation from
// the store
func (k Keeper) GetLastObservedEthereumBlockHeight(ctx sdk.Context) types.LastObservedEthereumBlockHeight {
	store := k.GetStore(ctx)
	bytes := store.Get(types.LastObservedEthereumBlockHeightKey)

	if len(bytes) == 0 {
		return types.LastObservedEthereumBlockHeight{
			PalomaBlockHeight:   0,
			EthereumBlockHeight: 0,
		}
	}
	height := types.LastObservedEthereumBlockHeight{
		PalomaBlockHeight:   0,
		EthereumBlockHeight: 0,
	}
	k.cdc.MustUnmarshal(bytes, &height)
	return height
}

// SetLastObservedEthereumBlockHeight sets the block height in the store.
func (k Keeper) SetLastObservedEthereumBlockHeight(ctx sdk.Context, ethereumHeight uint64) error {
	store := k.GetStore(ctx)
	previous := k.GetLastObservedEthereumBlockHeight(ctx)
	if previous.EthereumBlockHeight > ethereumHeight {
		return fmt.Errorf("attempt to roll back Ethereum block height")
	}
	height := types.LastObservedEthereumBlockHeight{
		EthereumBlockHeight: ethereumHeight,
		PalomaBlockHeight:   uint64(ctx.BlockHeight()),
	}
	store.Set(types.LastObservedEthereumBlockHeightKey, k.cdc.MustMarshal(&height))
	return nil
}

// setLastObservedEventNonce sets the latest observed event nonce
func (k Keeper) setLastObservedEventNonce(ctx sdk.Context, nonce uint64) error {
	store := k.GetStore(ctx)
	last, err := k.GetLastObservedEventNonce(ctx)
	if err != nil {
		return err
	}
	// event nonce must increase, unless it's zero at which point allow zero to be set
	// as many times as needed (genesis test setup etc)
	zeroCase := last == 0 && nonce == 0
	if last >= nonce && !zeroCase {
		return fmt.Errorf("event nonce going backwards or replay")
	}
	store.Set(types.LastObservedEventNonceKey, types.UInt64Bytes(nonce))
	return nil
}

// GetLastEventNonceByValidator returns the latest event nonce for a given validator
func (k Keeper) GetLastEventNonceByValidator(ctx sdk.Context, validator sdk.ValAddress) (uint64, error) {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		return 0, sdkerrors.Wrap(err, "invalid validator address")
	}
	store := k.GetStore(ctx)
	lastEventNonceByValidator, err := types.GetLastEventNonceByValidatorKey(validator)
	if err != nil {
		return 0, err
	}
	bytes := store.Get(lastEventNonceByValidator)

	if len(bytes) == 0 {
		// in the case that we have no existing value this is the first
		// time a validator is submitting a claim. Since we don't want to force
		// them to replay the entire history of all events ever we can't start
		// at zero
		lastEventNonce, err := k.GetLastObservedEventNonce(ctx)
		if err != nil {
			return 0, err
		}
		if lastEventNonce >= 1 {
			return lastEventNonce - 1, nil
		} else {
			return 0, nil
		}
	}
	return types.UInt64FromBytes(bytes)
}

// SetLastEventNonceByValidator sets the latest event nonce for a give validator
func (k Keeper) SetLastEventNonceByValidator(ctx sdk.Context, validator sdk.ValAddress, nonce uint64) error {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		return sdkerrors.Wrap(err, "invalid validator address")
	}
	store := k.GetStore(ctx)
	lastEventNoneByValidator, err := types.GetLastEventNonceByValidatorKey(validator)
	if err != nil {
		return err
	}
	store.Set(lastEventNoneByValidator, types.UInt64Bytes(nonce))
	return nil
}

// IterateValidatorLastEventNonces iterates through all batch confirmations
func (k Keeper) IterateValidatorLastEventNonces(ctx sdk.Context, cb func(key []byte, nonce uint64) (stop bool)) error {
	store := k.GetStore(ctx)
	prefixStore := prefix.NewStore(store, types.LastEventNonceByValidatorKey)
	iter := prefixStore.Iterator(nil, nil)

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		nonce, err := types.UInt64FromBytes(iter.Value())
		if err != nil {
			return err
		}
		// cb returns true to stop early
		if cb(iter.Key(), nonce) {
			break
		}
	}
	return nil
}
