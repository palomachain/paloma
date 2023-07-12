package keeper

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/palomachain/paloma/util/libvalid"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/valset/types"
)

const (
	// TODO: make this a param
	defaultKeepAliveDuration        = 5 * time.Minute
	cValidatorJailedErrorMessage    = "validator is jailed"
	cValidatorNotBondedErrorMessage = "validator is not bonded"
	cJailingImminentThreshold       = 2 * time.Minute
	cGracePeriodBlockHeight         = 10
	cUnjailedSnapshotStoreKey       = "unjailed-validators-snapshot"
)

type keepAliveData struct {
	ValAddr     sdk.ValAddress
	ContactedAt time.Time
	AliveUntil  time.Time
}

func (k Keeper) KeepValidatorAlive(ctx sdk.Context, valAddr sdk.ValAddress) error {
	if err := k.CanAcceptValidator(ctx, valAddr); err != nil {
		// Make sure we allow validators to keep alive even if jailed
		// or not bonded.
		// https://github.com/VolumeFi/paloma/issues/254
		if whoops.Is(err, ErrValidatorWithAddrNotFound) {
			return err
		}
	}

	store := k.keepAliveStore(ctx)
	data := keepAliveData{
		ValAddr:     valAddr,
		ContactedAt: ctx.BlockTime(),
		AliveUntil:  ctx.BlockTime().Add(defaultKeepAliveDuration),
	}
	bz, err := json.Marshal(data)
	if err != nil {
		return err
	}
	store.Set(valAddr, bz)
	return nil
}

func (k Keeper) IsValidatorAlive(ctx sdk.Context, valAddr sdk.ValAddress) (bool, error) {
	aliveUntil, err := k.ValidatorAliveUntil(ctx, valAddr)
	if err != nil {
		return false, err
	}
	return ctx.BlockTime().Before(aliveUntil), nil
}

func (k Keeper) ValidatorAliveUntil(ctx sdk.Context, valAddr sdk.ValAddress) (time.Time, error) {
	store := k.keepAliveStore(ctx)
	if !store.Has(valAddr) {
		return time.Time{}, ErrValidatorNotInKeepAlive.Format(valAddr)
	}

	dataBz := store.Get(valAddr)
	var data keepAliveData
	err := json.Unmarshal(dataBz, &data)
	if err != nil {
		return time.Time{}, err
	}

	if data.AliveUntil.UTC().Sub(ctx.BlockTime().UTC()) < cJailingImminentThreshold {
		k.Logger(ctx).Info("Validator TTL is about to run out. Jailing is imminent.", "validator-address", data.ValAddr)
	}

	return data.AliveUntil, nil
}

func (k Keeper) CanAcceptValidator(ctx sdk.Context, valAddr sdk.ValAddress) error {
	stakingVal := k.staking.Validator(ctx, valAddr)

	if stakingVal == nil {
		return ErrValidatorWithAddrNotFound.Format(valAddr.String())
	}

	if stakingVal.IsJailed() {
		return ErrValidatorCannotBePigeon.Format(valAddr.String()).WrapS(cValidatorJailedErrorMessage)
	}

	if !stakingVal.IsBonded() {
		return ErrValidatorCannotBePigeon.Format(valAddr.String()).WrapS(cValidatorNotBondedErrorMessage)
	}

	return nil
}

// UpdateGracePeriod will compare the list of active validators against the
// snapshot taken during the last block. Any new members will receive a grace
// period of n blocks.
// Active validators are stored as one flattened entry to relief pressure on
// reads every block.
// Call this during the EndBlock logic.
func (k Keeper) UpdateGracePeriod(ctx sdk.Context) {
	us := k.unjailedSnapshotStore(ctx)
	gs := k.gracePeriodStore(ctx)

	// Retrieve active validators from last block
	snapshot := bytes.Split(us.Get([]byte(cUnjailedSnapshotStoreKey)), []byte(","))
	lookup := make(map[string]struct{})
	for _, v := range snapshot {
		lookup[string(v)] = struct{}{}
	}

	vals := slice.Map(k.GetUnjailedValidators(ctx), func(i stakingtypes.ValidatorI) []byte {
		return i.GetOperator()
	})
	for _, v := range vals {
		if _, found := lookup[string(v)]; !found {
			// Looks like there's a new unjailed validator. Let's give them
			// some time before considering jailing them again.
			gs.Set(v, sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight())))
		}
	}

	// Record current snapshot of unjailed validators
	us.Set([]byte(cUnjailedSnapshotStoreKey), bytes.Join(vals, []byte(",")))
}

func (k Keeper) JailInactiveValidators(ctx sdk.Context) error {
	var g whoops.Group
	for _, val := range k.GetUnjailedValidators(ctx) {
		if !(val.GetStatus() == stakingtypes.Bonded || val.GetStatus() == stakingtypes.Unbonding) {
			continue
		}
		valAddr := val.GetOperator()
		alive, err := k.IsValidatorAlive(ctx, valAddr)
		switch {
		case err == nil:
			// does nothing
		case errors.Is(err, ErrValidatorNotInKeepAlive):
			// well...sucks to be you
		default:
			g.Add(err)
			continue
		}
		if alive {
			continue
		}

		if k.isValidatorInGracePeriod(ctx, valAddr) {
			// Pigeon is not alive, but validator still covered by grace period
			continue
		}

		if !k.IsJailed(ctx, valAddr) {
			g.Add(
				k.Jail(ctx, valAddr, types.JailReasonPigeonInactive),
			)
		}
	}
	return g.Return()
}

func (k Keeper) isValidatorInGracePeriod(ctx sdk.Context, valAddr sdk.ValAddress) bool {
	store := k.gracePeriodStore(ctx)
	bytes := store.Get(valAddr)
	return libvalid.NotNil(bytes) && ctx.BlockHeight()-int64(sdk.BigEndianToUint64(bytes)) <= cGracePeriodBlockHeight

}

func (k Keeper) keepAliveStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("keep-alive/"))
}
