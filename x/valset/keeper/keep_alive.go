package keeper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	keeperutil "github.com/palomachain/paloma/v2/util/keeper"
	"github.com/palomachain/paloma/v2/util/liblog"
	"github.com/palomachain/paloma/v2/util/libvalid"
	"github.com/palomachain/paloma/v2/util/slice"
	"github.com/palomachain/paloma/v2/x/valset/types"
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

func (k Keeper) KeepValidatorAlive(ctx context.Context, valAddr sdk.ValAddress, pigeonVersion string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := k.CanAcceptKeepAlive(ctx, valAddr, pigeonVersion); err != nil {
		return err
	}
	store := k.keepAliveStore(ctx)
	data := types.KeepAliveData{
		ValAddr:               valAddr,
		ContactedAt:           sdkCtx.BlockTime(),
		AliveUntilBlockHeight: sdkCtx.BlockHeader().Height + cJailingDefaultKeepAliveBlockHeight,
		PigeonVersion:         pigeonVersion,
	}
	bz, err := json.Marshal(data)
	if err != nil {
		return err
	}
	store.Set(valAddr, bz)
	return nil
}

func (k Keeper) IsValidatorAlive(ctx context.Context, valAddr sdk.ValAddress) (bool, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	data, err := k.ValidatorKeepAliveData(ctx, valAddr)
	if err != nil {
		return false, err
	}
	return sdkCtx.BlockHeight() < data.AliveUntilBlockHeight, nil
}

func (k Keeper) ValidatorKeepAliveData(
	ctx context.Context,
	valAddr sdk.ValAddress,
) (types.KeepAliveData, error) {
	var data types.KeepAliveData

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.keepAliveStore(ctx)
	if !store.Has(valAddr) {
		return data, ErrValidatorNotInKeepAlive.Format(valAddr)
	}
	dataBz := store.Get(valAddr)
	err := json.Unmarshal(dataBz, &data)
	if err != nil {
		return data, err
	}
	if data.AliveUntilBlockHeight-sdkCtx.BlockHeight() <= cJailingImminentThresholdBlockHeight {
		liblog.FromSDKLogger(k.Logger(ctx)).WithFields("validator-address", data.ValAddr).Info("Validator TTL is about to run out. Jailing is imminent.")
	}

	return data, nil
}

func (k Keeper) CanAcceptKeepAlive(ctx context.Context, valAddr sdk.ValAddress, pigeonVersion string) error {
	stakingVal, err := k.staking.Validator(ctx, valAddr)
	if err != nil {
		return err
	}

	if stakingVal == nil {
		return ErrValidatorWithAddrNotFound.Format(valAddr.String())
	}

	req, err := k.PigeonRequirements(ctx)
	if err != nil {
		return err
	}

	if semver.Compare(pigeonVersion, req.MinVersion) < 0 {
		return ErrValidatorPigeonOutOfDate.Format(valAddr.String(), pigeonVersion, req.MinVersion)
	}

	return nil
}

func (k Keeper) CanAcceptValidator(ctx context.Context, valAddr sdk.ValAddress) error {
	stakingVal, err := k.staking.Validator(ctx, valAddr)
	if err != nil {
		return err
	}
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
func (k Keeper) UpdateGracePeriod(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	us := k.unjailedSnapshotStore(ctx)
	gs := k.gracePeriodStore(ctx)

	// Retrieve active validators from last block
	snapshot := bytes.Split(us.Get([]byte(cUnjailedSnapshotStoreKey)), []byte(","))
	lookup := make(map[string]struct{})
	for _, v := range snapshot {
		lookup[string(v)] = struct{}{}
	}

	vals, err := slice.MapErr(k.GetUnjailedValidators(ctx), func(i stakingtypes.ValidatorI) ([]byte, error) {
		bz, err := keeperutil.ValAddressFromBech32(k.AddressCodec, i.GetOperator())
		return bz, err
	})
	if err != nil {
		return err
	}
	for _, v := range vals {
		if _, found := lookup[string(v)]; !found {
			// Looks like there's a new unjailed validator. Let's give them
			// some time before considering jailing them again.
			gs.Set(v, sdk.Uint64ToBigEndian(uint64(sdkCtx.BlockHeight())))
		}
	}

	// Record current snapshot of unjailed validators
	us.Set([]byte(cUnjailedSnapshotStoreKey), bytes.Join(vals, []byte(",")))
	return nil
}

func (k Keeper) JailInactiveValidators(ctx context.Context) error {
	var g whoops.Group
	for _, val := range k.GetUnjailedValidators(ctx) {
		if !(val.GetStatus() == stakingtypes.Bonded || val.GetStatus() == stakingtypes.Unbonding) {
			continue
		}
		valAddr, err := keeperutil.ValAddressFromBech32(k.AddressCodec, val.GetOperator())
		if err != nil {
			return err
		}
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

		jailed, err := k.IsJailed(ctx, valAddr)
		if err != nil {
			return err
		}
		if !jailed {
			g.Add(
				k.Jail(ctx, valAddr, types.JailReasonPigeonInactive),
			)
		}
	}
	return g.Return()
}

func (k Keeper) isValidatorInGracePeriod(ctx context.Context, valAddr sdk.ValAddress) bool {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.gracePeriodStore(ctx)
	bytes := store.Get(valAddr)
	return libvalid.NotNil(bytes) && sdkCtx.BlockHeight()-int64(sdk.BigEndianToUint64(bytes)) <= cJailingGracePeriodBlockHeight
}

func (k Keeper) keepAliveStore(ctx context.Context) storetypes.KVStore {
	store := runtime.KVStoreAdapter(k.storeKey.OpenKVStore(ctx))
	return prefix.NewStore(store, []byte("keep-alive/"))
}
