package keeper

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/palomachain/paloma/x/gravity/types"
)

//////////////////////////
/// VALSET REQUESTS //////
//////////////////////////

// SetValsetRequest returns a new instance of the Gravity BridgeValidatorSet
// by taking a snapshot of the current set, this validator set is also placed
// into the store to be signed by validators and submitted to Ethereum. This
// is the only function to call when you want to create a validator set that
// is signed by consensus. If you want to peek at the present state of the set
// and perhaps take action based on that use k.GetCurrentValset
// i.e. {"nonce": 1, "members": [{"eth_addr": "foo", "power": 11223}]}
func (k Keeper) SetValsetRequest(ctx sdk.Context) (types.Valset, error) {
	valset, err := k.GetCurrentValset(ctx)
	if err != nil {
		return valset, err
	}
	k.StoreValset(ctx, valset)
	k.SetLatestValsetNonce(ctx, valset.Nonce)

	// Store the checkpoint as a legit past valset, this is only for evidence
	// based slashing. We are storing the checkpoint that will be signed with
	// the validators Ethereum keys so that we know not to slash them if someone
	// attempts to submit the signature of this validator set as evidence of bad behavior
	checkpoint := valset.GetCheckpoint(k.GetGravityID(ctx))
	k.SetPastEthSignatureCheckpoint(ctx, checkpoint)

	err = ctx.EventManager().EmitTypedEvent(
		&types.EventMultisigUpdateRequest{
			BridgeContract: k.GetBridgeContractAddress(ctx).GetAddress().Hex(),
			BridgeChainId:  strconv.Itoa(int(k.GetBridgeChainID(ctx))),
			MultisigId:     fmt.Sprint(valset.Nonce),
			Nonce:          fmt.Sprint(valset.Nonce),
		},
	)
	return valset, err
}

// StoreValset is for storing a valiator set at a given height, once this function is called
// the validator set will be available to the Ethereum Signers (orchestrators) to submit signatures
// therefore this function will panic if you attempt to overwrite an existing key. Any changes to
// historical valsets can not possibly be correct, as it would invalidate the signatures. The only
// valid operation on the same index is store followed by delete when it is time to prune state
func (k Keeper) StoreValset(ctx sdk.Context, valset types.Valset) {
	key := types.GetValsetKey(valset.Nonce)
	store := ctx.KVStore(k.storeKey)

	if store.Has(key) {
		panic("Trying to overwrite existing valset!")
	}

	store.Set((key), k.cdc.MustMarshal(&valset))
}

// HasValsetRequest returns true if a valset defined by a nonce exists
func (k Keeper) HasValsetRequest(ctx sdk.Context, nonce uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetValsetKey(nonce))
}

// DeleteValset deletes the valset at a given nonce from state
func (k Keeper) DeleteValset(ctx sdk.Context, nonce uint64) {
	ctx.KVStore(k.storeKey).Delete(types.GetValsetKey(nonce))
}

// CheckLatestValsetNonce returns true if the latest valset nonce
// is declared in the store and false if it has not been initialized
func (k Keeper) CheckLatestValsetNonce(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	has := store.Has(types.LatestValsetNonce)
	return has
}

// GetLatestValsetNonce returns the latest valset nonce
func (k Keeper) GetLatestValsetNonce(ctx sdk.Context) uint64 {
	if !k.CheckLatestValsetNonce(ctx) {
		panic("Valset nonce not initialized from genesis")
	}

	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LatestValsetNonce)
	return types.UInt64FromBytesUnsafe(bytes)
}

// SetLatestValsetNonce sets the latest valset nonce, since it's
// expected that this value will only increase it panics on an attempt
// to decrement
func (k Keeper) SetLatestValsetNonce(ctx sdk.Context, nonce uint64) {
	// this is purely an increasing counter and should never decrease
	if k.CheckLatestValsetNonce(ctx) && k.GetLatestValsetNonce(ctx) > nonce {
		panic("Decrementing valset nonce!")
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.LatestValsetNonce, types.UInt64Bytes(nonce))
}

// GetValset returns a valset by nonce
func (k Keeper) GetValset(ctx sdk.Context, nonce uint64) *types.Valset {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValsetKey(nonce))
	if bz == nil {
		return nil
	}
	var valset types.Valset
	k.cdc.MustUnmarshal(bz, &valset)
	return &valset
}

// IterateValsets returns all valsetRequests in reverse order, aka most recent first
func (k Keeper) IterateValsets(ctx sdk.Context, cb func(key []byte, val *types.Valset) (stop bool)) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ValsetRequestKey)
	iter := prefixStore.ReverseIterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var valset types.Valset
		k.cdc.MustUnmarshal(iter.Value(), &valset)
		// cb returns true to stop early
		if cb(iter.Key(), &valset) {
			break
		}
	}
}

// GetValsets returns all the validator sets in state
func (k Keeper) GetValsets(ctx sdk.Context) (out []types.Valset) {
	k.IterateValsets(ctx, func(_ []byte, val *types.Valset) bool {
		out = append(out, *val)
		return false
	})
	sort.Sort(types.Valsets(out))
	return
}

// GetLatestValset returns the latest validator set in store. This is different
// from the CurrentValset because this one has been saved and is therefore *the* valset
// for this nonce. GetCurrentValset shows you what could be, if you chose to save it, this function
// shows you what is the latest valset that was saved.
func (k Keeper) GetLatestValset(ctx sdk.Context) (out *types.Valset) {
	latestValsetNonce := k.GetLatestValsetNonce(ctx)
	out = k.GetValset(ctx, latestValsetNonce)
	return
}

// setLastSlashedValsetNonce sets the latest slashed valset nonce
func (k Keeper) SetLastSlashedValsetNonce(ctx sdk.Context, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastSlashedValsetNonce, types.UInt64Bytes(nonce))
}

// GetLastSlashedValsetNonce returns the latest slashed valset nonce
func (k Keeper) GetLastSlashedValsetNonce(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastSlashedValsetNonce)

	if len(bytes) == 0 {
		return 0
	}
	return types.UInt64FromBytesUnsafe(bytes)
}

// SetLastUnBondingBlockHeight sets the last unbonding block height. Note this value is not saved and loaded in genesis
// and is reset to zero on chain upgrade.
func (k Keeper) SetLastUnBondingBlockHeight(ctx sdk.Context, unbondingBlockHeight uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastUnBondingBlockHeight, types.UInt64Bytes(unbondingBlockHeight))
}

// GetLastUnBondingBlockHeight returns the last unbonding block height, returns zero if not set, this is not
// saved or loaded ing enesis and is reset to zero on chain upgrade
func (k Keeper) GetLastUnBondingBlockHeight(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastUnBondingBlockHeight)

	if len(bytes) == 0 {
		return 0
	}
	return types.UInt64FromBytesUnsafe(bytes)
}

// GetUnSlashedValsets returns all the "ready-to-slash" unslashed validator sets in state (valsets at least signedValsetsWindow blocks old)
func (k Keeper) GetUnSlashedValsets(ctx sdk.Context, signedValsetsWindow uint64) (out []*types.Valset) {
	lastSlashedValsetNonce := k.GetLastSlashedValsetNonce(ctx)
	blockHeight := uint64(ctx.BlockHeight())
	k.IterateValsetBySlashedValsetNonce(ctx, lastSlashedValsetNonce, func(_ []byte, valset *types.Valset) bool {
		// Implicitly the unslashed valsets appear after the last slashed valset,
		// however not all valsets are ready-to-slash since validators have a window
		if valset.Nonce > lastSlashedValsetNonce && !(blockHeight < valset.Height+signedValsetsWindow) {
			out = append(out, valset)
		}
		return false
	})
	return
}

// IterateValsetBySlashedValsetNonce iterates through all valset by last slashed valset nonce in ASC order
func (k Keeper) IterateValsetBySlashedValsetNonce(ctx sdk.Context, lastSlashedValsetNonce uint64, cb func([]byte, *types.Valset) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ValsetRequestKey)
	// Consider all valsets, including the most recent one
	cutoffNonce := k.GetLatestValsetNonce(ctx) + 1
	iter := prefixStore.Iterator(types.UInt64Bytes(lastSlashedValsetNonce), types.UInt64Bytes(cutoffNonce))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var valset types.Valset
		k.cdc.MustUnmarshal(iter.Value(), &valset)
		// cb returns true to stop early
		if cb(iter.Key(), &valset) {
			break
		}
	}
}

// GetCurrentValset gets powers from the store and normalizes them
// into an integer percentage with a resolution of uint32 Max meaning
// a given validators 'gravity power' is computed as
// Cosmos power for that validator / total cosmos power = x / uint32 Max
// where x is the voting power on the gravity contract. This allows us
// to only use integer division which produces a known rounding error
// from truncation equal to the ratio of the validators
// Cosmos power / total cosmos power ratio, leaving us at uint32 Max - 1
// total voting power. This is an acceptable rounding error since floating
// point may cause consensus problems if different floating point unit
// implementations are involved.
//
// 'total cosmos power' has an edge case, if a validator has not set their
// Ethereum key they are not included in the total. If they were control
// of the bridge could be lost in the following situation.
//
// If we have 100 total power, and 100 total power joins the validator set
// the new validators hold more than 33% of the bridge power, if we generate
// and submit a valset and they don't have their eth keys set they can never
// update the validator set again and the bridge and all its' funds are lost.
// For this reason we exclude validators with unset eth keys from validator sets
//
// The function is intended to return what the valset would look like if you made one now
// you should call this function, evaluate if you want to save this new valset, and discard
// it or save
func (k Keeper) GetCurrentValset(ctx sdk.Context) (types.Valset, error) {
	validators := k.StakingKeeper.GetBondedValidatorsByPower(ctx)
	if len(validators) == 0 {
		// nolint: exhaustruct
		return types.Valset{}, types.ErrNoValidators
	}
	// allocate enough space for all validators, but len zero, we then append
	// so that we have an array with extra capacity but the correct length depending
	// on how many validators have keys set.
	bridgeValidators := make([]*types.InternalBridgeValidator, 0, len(validators))
	totalPower := sdk.NewInt(0)
	// TODO someone with in depth info on Cosmos staking should determine
	// if this is doing what I think it's doing
	for _, validator := range validators {
		val := validator.GetOperator()
		if err := sdk.VerifyAddressFormat(val); err != nil {
			return types.Valset{}, sdkerrors.Wrap(err, types.ErrInvalidValAddress.Error())
		}

		p := sdk.NewInt(k.StakingKeeper.GetLastValidatorPower(ctx, val))

		if ethAddr, found := k.GetEthAddressByValidator(ctx, val); found {
			bv := types.BridgeValidator{Power: p.Uint64(), EthereumAddress: ethAddr.GetAddress().Hex()}
			ibv, err := types.NewInternalBridgeValidator(bv)
			if err != nil {
				return types.Valset{}, sdkerrors.Wrapf(err, types.ErrInvalidEthAddress.Error(), val)
			}
			bridgeValidators = append(bridgeValidators, ibv)
			totalPower = totalPower.Add(p)
		}
	}
	// normalize power values to the maximum bridge power which is 2^32
	for i := range bridgeValidators {
		bridgeValidators[i].Power = normalizeValidatorPower(bridgeValidators[i].Power, totalPower)
	}

	// get the reward from the params store
	reward := k.GetParams(ctx).ValsetReward
	var rewardToken *types.EthAddress
	var rewardAmount sdk.Int
	if !reward.IsValid() || reward.IsZero() {
		// the case where a validator has 'no reward'. The 'no reward' value is interpreted as having a zero
		// address for the ERC20 token and a zero value for the reward amount. Since we store a coin with the
		// params, a coin with a blank denom and/or zero amount is interpreted in this way.
		za := types.ZeroAddress()
		rewardToken = &za
		rewardAmount = sdk.NewIntFromUint64(0)

	} else {
		rewardToken, rewardAmount = k.RewardToERC20Lookup(ctx, reward)
	}

	// increment the nonce, since this potential future valset should be after the current valset
	valsetNonce := k.GetLatestValsetNonce(ctx) + 1

	valset, err := types.NewValset(valsetNonce, uint64(ctx.BlockHeight()), bridgeValidators, rewardAmount, *rewardToken)
	if err != nil {
		return types.Valset{}, (sdkerrors.Wrap(err, types.ErrInvalidValset.Error()))
	}
	return *valset, nil
}

// normalizeValidatorPower scales rawPower with respect to totalValidatorPower to take a value between 0 and 2^32
// Uses BigInt operations to avoid overflow errors
// Example: rawPower = max (2^63 - 1), totalValidatorPower = 1 validator: (2^63 - 1)
// result: (2^63 - 1) * 2^32 / (2^63 - 1) = 2^32 = 4294967296 [this is the multiplier value below, our max output]
// Example: rawPower = max (2^63 - 1), totalValidatorPower = 1000 validators with the same power: 1000*(2^63 - 1)
// result: (2^63 - 1) * 2^32 / (1000(2^63 - 1)) = 2^32 / 1000 = 4294967
func normalizeValidatorPower(rawPower uint64, totalValidatorPower sdk.Int) uint64 {
	// Compute rawPower * multiplier / quotient
	// Set the upper limit to 2^32, which would happen if there is a single validator with all the power
	multiplier := new(big.Int).SetUint64(4294967296)
	// Scale by current validator powers, a particularly low-power validator (1 out of over 2^32) would have 0 power
	quotient := new(big.Int).Set(totalValidatorPower.BigInt())
	power := new(big.Int).SetUint64(rawPower)
	power.Mul(power, multiplier)
	power.Quo(power, quotient)
	return power.Uint64()
}

/////////////////////////////
// VALSET CONFIRMS     //
/////////////////////////////

// GetValsetConfirm returns a valset confirmation by a nonce and validator address
func (k Keeper) GetValsetConfirm(ctx sdk.Context, nonce uint64, validator sdk.AccAddress) *types.MsgValsetConfirm {
	store := ctx.KVStore(k.storeKey)
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		ctx.Logger().Error("invalid validator address")
		return nil
	}
	entity := store.Get(types.GetValsetConfirmKey(nonce, validator))
	if entity == nil {
		return nil
	}
	confirm := types.MsgValsetConfirm{
		Nonce:        nonce,
		Orchestrator: "",
		EthAddress:   "",
		Signature:    "",
	}
	k.cdc.MustUnmarshal(entity, &confirm)
	return &confirm
}

// SetValsetConfirm sets a valset confirmation
func (k Keeper) SetValsetConfirm(ctx sdk.Context, valsetConf types.MsgValsetConfirm) []byte {
	store := ctx.KVStore(k.storeKey)
	addr, err := sdk.AccAddressFromBech32(valsetConf.Orchestrator)
	if err != nil {
		panic(err)
	}
	key := types.GetValsetConfirmKey(valsetConf.Nonce, addr)
	store.Set(key, k.cdc.MustMarshal(&valsetConf))
	return key
}

// GetValsetConfirms returns all validator set confirmations by nonce
func (k Keeper) GetValsetConfirms(ctx sdk.Context, nonce uint64) (confirms []types.MsgValsetConfirm) {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetValsetConfirmNoncePrefix(nonce)
	iterator := store.Iterator(prefixRange([]byte(prefix)))

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		confirm := types.MsgValsetConfirm{
			Nonce:        nonce,
			Orchestrator: "",
			EthAddress:   "",
			Signature:    "",
		}
		k.cdc.MustUnmarshal(iterator.Value(), &confirm)
		confirms = append(confirms, confirm)
	}

	return confirms
}

// IterateValsetConfirms returns all valset confirms in reverse order by nonce, aka most recent valset nonce first
func (k Keeper) IterateValsetConfirms(ctx sdk.Context, cb func(key []byte, confirms []types.MsgValsetConfirm, nonce uint64) (stop bool)) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ValsetConfirmKey)
	iter := prefixStore.ReverseIterator(nil, nil)
	defer iter.Close()

	var confirmsByNonce []types.MsgValsetConfirm
	var nonce uint64 = 0

	// Iterate through all stored valset confirms grouping them into confirmsByNonce.
	// When a new nonce is found: process the collected confirmsByNonce, clear the collection + update the current nonce
	var key []byte
	for ; iter.Valid(); iter.Next() {
		key = iter.Key()
		// The iterator guarantees us some nonempty value stored at the Value(), it better be a MsgValsetConfirm
		var confirm types.MsgValsetConfirm
		k.cdc.MustUnmarshal(iter.Value(), &confirm)

		if len(confirmsByNonce) == 0 { // First confirm found
			// Initialize current nonce, start the collection
			nonce = confirm.Nonce
			confirmsByNonce = append(confirmsByNonce, confirm)
		} else {
			if confirm.Nonce != nonce { // At nonce boundary
				// We are guaranteed to have some collection of confirms at this point
				// cb returns true to stop early
				if cb(key, confirmsByNonce, nonce) {
					return
				}
				// Now we are done with the old confirms, update the nonce and clear the collection
				nonce = confirm.Nonce
				confirmsByNonce = []types.MsgValsetConfirm{}
			}
			// Add the current confirm to the collection
			confirmsByNonce = append(confirmsByNonce, confirm)
		}
	}

	// Process the final nonce, or arrive here if no confirms are in the store
	if len(confirmsByNonce) != 0 {
		// This is the final callback execution, we stop regardless of their request to do so
		cb(key, confirmsByNonce, nonce)
	}
}

// DeleteValsetConfirms deletes the valset confirmations for the valset at a given nonce from state
func (k Keeper) DeleteValsetConfirms(ctx sdk.Context, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	for _, confirm := range k.GetValsetConfirms(ctx, nonce) {
		orchestrator, err := sdk.AccAddressFromBech32(confirm.Orchestrator)
		if err == nil {
			confirmKey := types.GetValsetConfirmKey(nonce, orchestrator)
			if store.Has(confirmKey) {
				store.Delete(confirmKey)
			}
		}
	}
}
