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
	"golang.org/x/mod/semver"
)

const (
	// Deprecated. Remove after https://github.com/VolumeFi/paloma/issues/707 is deployed on all nets
	defaultKeepAliveDuration = 5 * time.Minute
	// Deprecated. Remove after https://github.com/VolumeFi/paloma/issues/707 is deployed on all nets
	cJailingImminentThreshold = 2 * time.Minute

	cValidatorJailedErrorMessage         = "validator is jailed"
	cValidatorNotBondedErrorMessage      = "validator is not bonded"
	cJailingDefaultKeepAliveBlockHeight  = 185 // calculated against current block speed of 1.612 seconds
	cJailingImminentThresholdBlockHeight = 60  // publish warning if less than 60 blocks worth of TTL remaining
	cJailingGracePeriodBlockHeight       = 30  // don't jail a validator during the first 30 blocks after unjailing
	cUnjailedSnapshotStoreKey            = "unjailed-validators-snapshot"
)

type keepAliveData struct {
	ValAddr     sdk.ValAddress
	ContactedAt time.Time
	// Deprecated. Remove after https://github.com/VolumeFi/paloma/issues/707 is deployed on all nets
	AliveUntil            time.Time
	AliveUntilBlockHeight int64
}

func (k Keeper) KeepValidatorAlive(ctx sdk.Context, valAddr sdk.ValAddress, pigeonVersion string) error {
	if err := k.CanAcceptKeepAlive(ctx, valAddr, pigeonVersion); err != nil {
		return err
	}

	store := k.keepAliveStore(ctx)
	data := keepAliveData{
		ValAddr:               valAddr,
		ContactedAt:           ctx.BlockTime(),
		AliveUntilBlockHeight: ctx.BlockHeader().Height + cJailingDefaultKeepAliveBlockHeight,
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

	return ctx.BlockHeight() < aliveUntil, nil
}

func (k Keeper) ValidatorAliveUntil(ctx sdk.Context, valAddr sdk.ValAddress) (int64, error) {
	store := k.keepAliveStore(ctx)
	if !store.Has(valAddr) {
		return 0, ErrValidatorNotInKeepAlive.Format(valAddr)
	}

	dataBz := store.Get(valAddr)
	var data keepAliveData
	err := json.Unmarshal(dataBz, &data)
	if err != nil {
		return 0, err
	}

	// hack hack hack
	// To ensure a smooth migration to block based Pigeon TTL, we're going to pretend we received a valid
	// future block height as long as the AliveUntil time is still valid.
	// remove this once https://github.com/VolumeFi/paloma/issues/709 has been deployed to all nets
	if data.AliveUntilBlockHeight == 0 {
		if ctx.BlockTime().Before(data.AliveUntil) {
			data.AliveUntilBlockHeight = ctx.BlockHeight() + cJailingDefaultKeepAliveBlockHeight
		}
	}

	if data.AliveUntilBlockHeight-ctx.BlockHeight() <= cJailingImminentThresholdBlockHeight {
		k.Logger(ctx).Info("Validator TTL is about to run out. Jailing is imminent.", "validator-address", data.ValAddr)
	}

	return data.AliveUntilBlockHeight, nil
}

func (k Keeper) CanAcceptKeepAlive(ctx sdk.Context, valAddr sdk.ValAddress, pigeonVersion string) error {
	stakingVal := k.staking.Validator(ctx, valAddr)

	if stakingVal == nil {
		return ErrValidatorWithAddrNotFound.Format(valAddr.String())
	}

	if semver.Compare(pigeonVersion, k.minimumPigeonVersion) < 0 {
		return ErrValidatorPigeonOutOfDate.Format(valAddr.String(), pigeonVersion, k.minimumPigeonVersion)
	}

	return nil
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
	return libvalid.NotNil(bytes) && ctx.BlockHeight()-int64(sdk.BigEndianToUint64(bytes)) <= cJailingGracePeriodBlockHeight
}

func (k Keeper) keepAliveStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("keep-alive/"))
}
