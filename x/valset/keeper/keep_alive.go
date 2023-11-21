package keeper

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"cosmossdk.io/store/prefix"
	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/palomachain/paloma/util/libvalid"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/valset/types"
	"golang.org/x/mod/semver"
)

const (
	cValidatorJailedErrorMessage         = "validator is jailed"
	cValidatorNotBondedErrorMessage      = "validator is not bonded"
	cJailingDefaultKeepAliveBlockHeight  = 2000 // ca. 50 minutes at 1.62s block speed
	cJailingImminentThresholdBlockHeight = 60   // publish warning if less than 60 blocks worth of TTL remaining
	cJailingGracePeriodBlockHeight       = 30   // don't jail a validator during the first 30 blocks after unjailing
	cUnjailedSnapshotStoreKey            = "unjailed-validators-snapshot"
)

type keepAliveData struct {
	ContactedAt           time.Time
	ValAddr               sdk.ValAddress
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
