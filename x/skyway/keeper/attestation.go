package keeper

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strconv"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/VolumeFi/whoops"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	utilkeeper "github.com/palomachain/paloma/v2/util/keeper"
	"github.com/palomachain/paloma/v2/util/liblog"
	"github.com/palomachain/paloma/v2/x/skyway/types"
)

func (k Keeper) Attest(
	ctx context.Context,
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

	valAddress, err := utilkeeper.ValAddressFromBech32(k.AddressCodec, val.GetOperator())
	if err != nil {
		return nil, err
	}

	if err := sdk.VerifyAddressFormat(valAddress); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid orchestrator validator address")
	}
	// Check that the nonce of this event is exactly one higher than the last nonce stored by this validator.
	// We check the event nonce in processAttestation as well,
	// but checking it here gives individual eth signers a chance to retry,
	// and prevents validators from submitting two claims with the same nonce.
	// This prevents there being two attestations with the same nonce that get 2/3s of the votes
	// in the endBlocker.
	lastSkywayNonce, err := k.GetLastSkywayNonceByValidator(ctx, valAddress, claim.GetChainReferenceId())
	if err != nil {
		return nil, err
	}
	if claim.GetSkywayNonce() != lastSkywayNonce+1 {
		return nil, fmt.Errorf(types.ErrNonContiguousEventNonce.Error(), lastSkywayNonce+1, claim.GetSkywayNonce())
	}

	// Tries to get an attestation with the same eventNonce and claim as the claim that was submitted.
	hash, err := claim.ClaimHash()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to compute claim hash")
	}
	att := k.GetAttestation(ctx, claim.GetChainReferenceId(), claim.GetSkywayNonce(), hash)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// If it does not exist, create a new one.
	if att == nil {
		att = &types.Attestation{
			Observed: false,
			Votes:    []string{},
			Height:   uint64(sdkCtx.BlockHeight()),
			Claim:    anyClaim,
		}
	}

	ethClaim, err := k.UnpackAttestationClaim(att)
	if err != nil {
		return nil, fmt.Errorf("could not unpack stored attestation claim, %v", err)
	}

	if ethClaim.GetEthBlockHeight() == claim.GetEthBlockHeight() {

		// Add the validator's vote to this attestation // TODO : Only do if it's not already there
		att.Votes = append(att.Votes, valAddr)

		k.SetAttestation(ctx, claim.GetChainReferenceId(), claim.GetSkywayNonce(), hash, att)
		err = k.SetLastSkywayNonceByValidator(ctx, valAddress, claim.GetChainReferenceId(), claim.GetSkywayNonce())
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
func (k Keeper) TryAttestation(ctx context.Context, att *types.Attestation) error {
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
		totalPower, err := k.StakingKeeper.GetLastTotalPower(ctx)
		if err != nil {
			return err
		}
		requiredPower := types.AttestationVotesPowerThreshold.Mul(totalPower).Quo(math.NewInt(100))
		attestationPower := math.NewInt(0)
		for _, validator := range att.Votes {

			val, err := utilkeeper.ValAddressFromBech32(k.AddressCodec, validator)
			if err != nil {
				return err
			}
			validatorPower, err := k.StakingKeeper.GetLastValidatorPower(ctx, val)
			if err != nil {
				return err
			}
			// Add it to the attestation power's sum
			attestationPower = attestationPower.Add(math.NewInt(validatorPower))
			// If the power of all the validators that have voted on the attestation is higher or equal to the threshold,
			// process the attestation, set Observed to true, and break
			if attestationPower.GT(requiredPower) {
				lastSkywayNonce, err := k.GetLastObservedSkywayNonce(ctx, claim.GetChainReferenceId())
				if err != nil {
					return err
				}

				if claim.GetSkywayNonce() != lastSkywayNonce+1 {
					return fmt.Errorf("attempting to apply events to state out of order")
				}

				err = k.setLastObservedSkywayNonce(ctx, claim.GetChainReferenceId(), claim.GetSkywayNonce())
				if err != nil {
					return err
				}

				err = k.SetLastObservedEthereumBlockHeight(ctx, claim.GetChainReferenceId(), claim.GetEthBlockHeight())
				if err != nil {
					return err
				}

				att.Observed = true
				k.SetAttestation(ctx, claim.GetChainReferenceId(), claim.GetSkywayNonce(), hash, att)

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
func (k Keeper) processAttestation(goCtx context.Context, att *types.Attestation, claim types.EthereumClaim) error {
	ctx, commit := sdk.UnwrapSDKContext(goCtx).CacheContext()
	if err := k.AttestationHandler.Handle(ctx, *att, claim); err != nil {
		// execute with a transient storage
		// If the attestation fails, something has gone wrong and we can't recover it. Log and move on
		// The attestation will still be marked "Observed", allowing the oracle to progress properly
		hash, err := claim.ClaimHash()
		if err != nil {
			return fmt.Errorf("processAttestation: unable to compute claim hash: %w", err)
		}

		liblog.FromKeeper(ctx, k).
			WithComponent("process-attestation").
			WithFields(
				"claim type", claim.GetType(),
				"id", types.GetAttestationKey(claim.GetSkywayNonce(), hash),
				"nonce", claim.GetSkywayNonce()).
			WithError(err).
			Warn("attestation failed")
	} else {
		commit() // persist transient storage
	}
	return nil
}

// emitObservedEvent emits an event with information about an attestation that has been applied to
// consensus state.
func (k Keeper) emitObservedEvent(ctx context.Context, att *types.Attestation, claim types.EthereumClaim) error {
	hash, err := claim.ClaimHash()
	if err != nil {
		return sdkerrors.Wrap(err, "unable to compute claim hash")
	}

	ci, err := k.EVMKeeper.GetChainInfo(ctx, claim.GetChainReferenceId())
	if err != nil {
		return err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.EventManager().EmitTypedEvent(
		&types.EventObservation{
			AttestationType: string(claim.GetType()),
			BridgeContract:  ci.SmartContractAddr,
			BridgeChainId:   strconv.Itoa(int(ci.ChainID)),
			AttestationId:   string(types.GetAttestationKey(claim.GetSkywayNonce(), hash)),
			Nonce:           fmt.Sprint(claim.GetSkywayNonce()),
		},
	)
}

// SetAttestation sets the attestation in the store
func (k Keeper) SetAttestation(ctx context.Context, chainReferenceID string, eventNonce uint64, claimHash []byte, att *types.Attestation) {
	store := k.GetStore(ctx, chainReferenceID)
	aKey := types.GetAttestationKey(eventNonce, claimHash)
	store.Set(aKey, k.cdc.MustMarshal(att))
}

// GetAttestation return an attestation given a nonce
func (k Keeper) GetAttestation(ctx context.Context, chainReferenceID string, eventNonce uint64, claimHash []byte) *types.Attestation {
	store := k.GetStore(ctx, chainReferenceID)
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
func (k Keeper) DeleteAttestation(ctx context.Context, chainReferenceID string, att types.Attestation) error {
	claim, err := k.UnpackAttestationClaim(&att)
	if err != nil {
		return fmt.Errorf("bad attestation in DeleteAttestation")
	}
	hash, err := claim.ClaimHash()
	if err != nil {
		return sdkerrors.Wrap(err, "unable to compute claim hash")
	}
	store := k.GetStore(ctx, chainReferenceID)

	store.Delete(types.GetAttestationKey(claim.GetSkywayNonce(), hash))
	return nil
}

// GetAttestationMapping returns a mapping of eventnonce -> attestations at that nonce
// it also returns a pre-sorted array of the keys, this assists callers of this function
// by providing a deterministic iteration order. You should always iterate over ordered keys
// if you are iterating this map at all.
func (k Keeper) GetAttestationMapping(ctx context.Context, chainReferenceID string) (attestationMapping map[uint64][]types.Attestation, orderedKeys []uint64, err error) {
	attestationMapping = make(map[uint64][]types.Attestation)
	var g whoops.Group

	lastCompassID := k.GetLatestCompassID(ctx, chainReferenceID)

	g.Add(
		k.IterateAttestations(ctx, chainReferenceID, false, func(_ []byte, att types.Attestation) bool {
			claim, err := k.UnpackAttestationClaim(&att)
			if err != nil {
				g.Add(err)
				return true
			}

			// Do not include claims originating from the other compass
			// versions, since we won't be able to attest them anyway. They may
			// also repeat the nonce, so we can't list them here.
			if lastCompassID != "" && claim.GetCompassID() != lastCompassID {
				return false
			}

			if val, ok := attestationMapping[claim.GetSkywayNonce()]; !ok {
				attestationMapping[claim.GetSkywayNonce()] = []types.Attestation{att}
			} else {
				attestationMapping[claim.GetSkywayNonce()] = append(val, att)
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
func (k Keeper) IterateAttestations(ctx context.Context, chainReferenceID string, reverse bool, cb func(key []byte, att types.Attestation) (stop bool)) error {
	store := k.GetStore(ctx, chainReferenceID)
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
func (k Keeper) IterateClaims(ctx context.Context, chainReferenceID string, reverse bool, claimType types.ClaimType, cb func(key []byte, att types.Attestation, claim types.EthereumClaim) (stop bool)) error {
	typeUrl := types.ClaimTypeToTypeUrl(claimType) // Used to avoid unpacking undesired attestations

	var g whoops.Group
	g.Add(
		k.IterateAttestations(ctx, chainReferenceID, reverse, func(key []byte, att types.Attestation) bool {
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
func (k Keeper) GetMostRecentAttestations(ctx context.Context, chainReferenceID string, limit uint64) ([]types.Attestation, error) {
	attestationMapping, keys, err := k.GetAttestationMapping(ctx, chainReferenceID)
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

// GetLastObservedSkywayNonce returns the latest observed event nonce
func (k Keeper) GetLastObservedSkywayNonce(ctx context.Context, chainReferenceID string) (uint64, error) {
	store := k.GetStore(ctx, chainReferenceID)
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
func (k Keeper) GetLastObservedEthereumBlockHeight(ctx context.Context, chainReferenceID string) types.LastObservedEthereumBlockHeight {
	store := k.GetStore(ctx, chainReferenceID)
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
func (k Keeper) SetLastObservedEthereumBlockHeight(ctx context.Context, chainReferenceID string, ethereumHeight uint64) error {
	store := k.GetStore(ctx, chainReferenceID)
	previous := k.GetLastObservedEthereumBlockHeight(ctx, chainReferenceID)
	if previous.EthereumBlockHeight > ethereumHeight {
		return fmt.Errorf("attempt to roll back Ethereum block height")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := types.LastObservedEthereumBlockHeight{
		EthereumBlockHeight: ethereumHeight,
		PalomaBlockHeight:   uint64(sdkCtx.BlockHeight()),
	}
	store.Set(types.LastObservedEthereumBlockHeightKey, k.cdc.MustMarshal(&height))
	return nil
}

// setLastObservedSkywayNonce sets the latest observed event nonce
func (k Keeper) setLastObservedSkywayNonce(ctx context.Context, chainReferenceID string, nonce uint64) error {
	store := k.GetStore(ctx, chainReferenceID)
	last, err := k.GetLastObservedSkywayNonce(ctx, chainReferenceID)
	if err != nil {
		return err
	}

	// event nonce must increase, unless it's zero at which point allow zero to be set
	// as many times as needed (genesis test setup etc)
	if nonce != 0 && nonce < last {
		return fmt.Errorf("event nonce going backwards or replay")
	}

	store.Set(types.LastObservedEventNonceKey, types.UInt64Bytes(nonce))
	return nil
}

// GetLastSkywayNonceByValidator returns the latest event nonce for a given validator
func (k Keeper) GetLastSkywayNonceByValidator(ctx context.Context, validator sdk.ValAddress, chainReferenceID string) (uint64, error) {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		return 0, sdkerrors.Wrap(err, "invalid validator address")
	}
	store := k.GetStore(ctx, chainReferenceID)
	key, err := types.GetLastEventNonceByValidatorKey(validator)
	if err != nil {
		return 0, err
	}
	bytes := store.Get(key)

	if len(bytes) == 0 {
		// in the case that we have no existing value this is the first
		// time a validator is submitting a claim. Since we don't want to force
		// them to replay the entire history of all events ever we can't start
		// at zero
		lastEventNonce, err := k.GetLastObservedSkywayNonce(ctx, chainReferenceID)
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

// SetLastSkywayNonceByValidator sets the latest event nonce for a give validator
func (k Keeper) SetLastSkywayNonceByValidator(ctx context.Context, validator sdk.ValAddress, chainReferenceID string, nonce uint64) error {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		return sdkerrors.Wrap(err, "invalid validator address")
	}
	store := k.GetStore(ctx, chainReferenceID)
	key, err := types.GetLastEventNonceByValidatorKey(validator)
	if err != nil {
		return err
	}
	store.Set(key, types.UInt64Bytes(nonce))
	return nil
}

// IterateValidatorLastEventNonces iterates through all batch confirmations
func (k Keeper) IterateValidatorLastEventNonces(ctx context.Context, chainReferenceID string, cb func(key []byte, nonce uint64) (stop bool)) error {
	store := k.GetStore(ctx, chainReferenceID)
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

// GetLatestCompassID returns the latest compass ID on record after
// message attestation
func (k Keeper) GetLatestCompassID(
	ctx context.Context,
	chainReferenceID string,
) string {
	store := k.GetStore(ctx, chainReferenceID)
	bytes := store.Get(types.LatestCompassIDKey)

	return string(bytes)
}

// setLatestCompassID stores the latest compass ID received
func (k Keeper) setLatestCompassID(
	ctx context.Context,
	chainReferenceID string,
	compassID string,
) {
	store := k.GetStore(ctx, chainReferenceID)
	store.Set(types.LatestCompassIDKey, []byte(compassID))
}

// UnobservedBlocksByAddr returns the blocks with reported events that
// are still unobserved and are not yet voted by `valAddr`
func (k Keeper) UnobservedBlocksByAddr(
	ctx context.Context,
	chainReferenceID string,
	valAddr string,
) ([]uint64, error) {
	lastCompassID := k.GetLatestCompassID(ctx, chainReferenceID)

	lastObservedNonce, err := k.GetLastObservedSkywayNonce(ctx, chainReferenceID)
	if err != nil {
		return nil, err
	}

	var blocks []uint64

	iterErr := k.IterateAttestations(ctx, chainReferenceID, false, func(_ []byte, att types.Attestation) bool {
		if att.Observed {
			return false
		}

		if slices.Contains(att.Votes, valAddr) {
			return false
		}

		var claim types.EthereumClaim
		claim, err = k.UnpackAttestationClaim(&att)
		if err != nil {
			return true
		}

		// Only include claims from current compass deployment
		if lastCompassID != "" && claim.GetCompassID() != lastCompassID {
			return false
		}

		// If we ever reset the latest skyway nonce ahead of any observed claim,
		// we don't need nor want pigeons trying to attest to them. So, we can
		// filter claims that are behind the latest attested claim.
		if claim.GetSkywayNonce() <= lastObservedNonce {
			return false
		}

		blocks = append(blocks, claim.GetEthBlockHeight())

		return false
	})
	if err != nil {
		return nil, err
	}

	if iterErr != nil {
		return nil, iterErr
	}

	// Return the blocks in ascending order
	slices.Sort(blocks)

	return blocks, nil
}
