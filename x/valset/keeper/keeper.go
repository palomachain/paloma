package keeper

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sort"

	sdkmath "cosmossdk.io/math"
	"github.com/VolumeFi/whoops"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/valset/types"
)

const (
	snapshotIDKey                   = "snapshot-id"
	maxNumOfAllowedExternalAccounts = 100
	cJailingNetworkShareProtection  = 0.25
)

type Keeper struct {
	EvmKeeper         types.EvmKeeper
	SnapshotListeners []types.OnSnapshotBuiltListener

	cdc                  codec.BinaryCodec
	ider                 keeperutil.IDGenerator
	memKey               storetypes.StoreKey
	minimumPigeonVersion string
	paramstore           paramtypes.Subspace
	powerReduction       sdkmath.Int
	staking              types.StakingKeeper
	storeKey             storetypes.StoreKey
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	staking types.StakingKeeper,
	minimumPigeonVersion string,
	powerReduction sdkmath.Int,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	k := &Keeper{
		cdc:                  cdc,
		storeKey:             storeKey,
		memKey:               memKey,
		paramstore:           ps,
		staking:              staking,
		minimumPigeonVersion: minimumPigeonVersion,
		powerReduction:       powerReduction,
	}
	k.ider = keeperutil.NewIDGenerator(keeperutil.StoreGetterFn(func(ctx sdk.Context) sdk.KVStore {
		return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("IDs"))
	}), nil)

	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// TODO: not required now
func (k Keeper) PunishValidator(ctx sdk.Context) {}

// TODO: not required now
func (k Keeper) Heartbeat(ctx sdk.Context) {}

// addExternalChainInfo adds external chain info, such as this conductor's address on outside chains so that
// we can attribute rewards for running the jobs.
func (k Keeper) AddExternalChainInfo(ctx sdk.Context, valAddr sdk.ValAddress, newChainInfo []*types.ExternalChainInfo) error {
	return k.SetExternalChainInfoState(ctx, valAddr, newChainInfo)
}

func (k Keeper) SetValidatorBalance(ctx sdk.Context, valAddr sdk.ValAddress, chainType string, chainReferenceID string, externalAddress string, balance *big.Int) error {
	chainInfos, err := k.GetValidatorChainInfos(ctx, valAddr)
	if err != nil {
		return err
	}
	found := false
	for _, ci := range chainInfos {
		if ci.GetChainReferenceID() == chainReferenceID && ci.GetChainType() == chainType && ci.GetAddress() == externalAddress {
			ci.Balance = balance.Text(10)
			found = true
			break
		}
	}
	if !found {
		return ErrValidatorWithAddrNotFound.Format(chainType, chainReferenceID, externalAddress, valAddr)
	}

	return k.SetExternalChainInfoState(ctx, valAddr, chainInfos)
}

func (k Keeper) SetExternalChainInfoState(ctx sdk.Context, valAddr sdk.ValAddress, chainInfos []*types.ExternalChainInfo) error {
	if len(chainInfos) > maxNumOfAllowedExternalAccounts {
		return ErrMaxNumberOfExternalAccounts.Format(
			len(chainInfos),
			maxNumOfAllowedExternalAccounts,
		)
	}

	if err := k.CanAcceptValidator(ctx, valAddr); err != nil {
		return err
	}

	allExistingChainAccounts, err := k.GetAllChainInfos(ctx)
	if err != nil {
		return err
	}

	var collisionErrors whoops.Group
	// O(n^2) to find if new one is already registered
	for _, existingVal := range allExistingChainAccounts {
		// we don't want to compare current validator's existing account
		// because it would most likely come up with a collision detection
		// error because most of the time this will be a noop, thus we skip
		// them
		if existingVal.Address.Equals(valAddr) {
			continue
		}
		for _, existingChainInfo := range existingVal.ExternalChainInfo {
			for _, newChainInfo := range chainInfos {
				if newChainInfo.GetChainType() != existingChainInfo.GetChainType() {
					continue
				}
				// this implies that pigeon can have only one(!) account per chain info.
				// this is an issue for compass-evm because compass-evm can't work with
				// multiple accounts existing for a single validator.
				if newChainInfo.GetChainReferenceID() != existingChainInfo.GetChainReferenceID() {
					continue
				}

				if newChainInfo.GetAddress() == existingChainInfo.GetAddress() || bytes.Equal(newChainInfo.GetPubkey(), existingChainInfo.GetPubkey()) {
					collisionErrors.Add(
						ErrExternalChainAlreadyRegistered.Format(
							newChainInfo.GetChainType(),
							newChainInfo.GetChainReferenceID(),
							newChainInfo.GetAddress(),
							existingVal.GetAddress().String(),
							valAddr.String(),
						),
					)
				}
			}
		}
	}

	if collisionErrors.Err() {
		return collisionErrors
	}

	store := k.externalChainInfoStore(ctx, valAddr)

	return keeperutil.Save(store, k.cdc, []byte(valAddr.String()), &types.ValidatorExternalAccounts{
		Address:           valAddr,
		ExternalChainInfo: chainInfos,
	})
}

// TriggerSnapshotBuild creates the snapshot of currently active validators that are
// active and registered as conductors.
func (k Keeper) TriggerSnapshotBuild(ctx sdk.Context) (*types.Snapshot, error) {
	snapshot, err := k.createNewSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("create new snapshot", "snapshot-id", snapshot.GetId())

	current, err := k.GetCurrentSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("get current triggered snapshot", "id", current.GetId(), "validator-length", len(current.GetValidators()))

	worthy := k.isNewSnapshotWorthy(ctx, current, snapshot)
	if !worthy {
		return nil, nil
	}

	k.Logger(ctx).Info("is worthy", "snapshot-id", current.GetId())

	err = k.setSnapshotAsCurrent(ctx, snapshot)
	if err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("set snapshot as current", "snapshot-id", snapshot.GetId())

	// remove jail reasons for all active validators.
	// given that a validator is in snapshot, they can't be jailed.
	for _, val := range snapshot.GetValidators() {
		k.jailReasonStore(ctx).Delete(val.GetAddress())
	}

	for _, listener := range k.SnapshotListeners {
		listener.OnSnapshotBuilt(ctx, snapshot)
	}

	return snapshot, err
}

func (k Keeper) isNewSnapshotWorthy(ctx sdk.Context, currentSnapshot, newSnapshot *types.Snapshot) bool {
	log := func(reason string) {
		k.Logger(ctx).Info("new snapshot is worthy", "reason", reason)
	}
	// if there is no current snapshot, that this new one is worthy
	if currentSnapshot == nil {
		log("this is the first snapshot")
		return true
	}

	// if there is a different in sizes of validators in snapshots, then we
	// need to build it
	if len(currentSnapshot.GetValidators()) != len(newSnapshot.GetValidators()) {
		log("number of validators in old and new snapshots differ")
		return true
	}

	// now that those sets are of the same size, we need to check if all new
	// validators are existing in the new current valset

	mapKeyFn := func(val types.Validator) string { return val.GetAddress().String() }
	currentMap := slice.MakeMapKeys(currentSnapshot.GetValidators(), mapKeyFn)

	// given that they are the same length we can only verify if one exists in another.
	// We don't need to check if A exists in B and if B exists in A.
	for _, val := range newSnapshot.GetValidators() {
		if _, ok := currentMap[val.GetAddress().String()]; !ok {
			log("snapshots differ in validators they hold")
			return true
		}
	}

	// given that both sets contains the same validators, we need to check if
	// their relative powers are still the same. To do that, we can simply
	// order them by their powers and if they are in the same order then
	// this new set is not worthy.
	returnSortedValidators := func(val []types.Validator) []types.Validator {
		ret := make([]types.Validator, len(val))
		copy(ret, val)
		sort.SliceStable(ret, func(i, j int) bool {
			return ret[i].ShareCount.LT(ret[j].ShareCount)
		})
		return ret
	}

	sortedCurrent, sortedNew := returnSortedValidators(currentSnapshot.GetValidators()), returnSortedValidators(newSnapshot.GetValidators())

	for i := 0; i < len(sortedCurrent); i++ {
		if !sortedCurrent[i].GetAddress().Equals(sortedNew[i].GetAddress()) {
			log("their relative powers are different")
			return true
		}
	}

	// and for the final check we want to see if their absolute powers were
	// changed by more than 1%.  What could happen is that the validator that
	// was previously the biggest one and owned les say 20% of the network, now
	// could own 60% of the network. And all other validators stayed the
	// (relatively) same.
	for i := 0; i < len(sortedCurrent); i++ {
		percentageCurrent := sdkmath.LegacyNewDecFromInt(sortedCurrent[i].ShareCount).QuoInt(currentSnapshot.TotalShares)
		percentageNow := sdkmath.LegacyNewDecFromInt(sortedNew[i].ShareCount).QuoInt(newSnapshot.TotalShares)

		if percentageCurrent.Sub(percentageNow).Abs().MustFloat64() >= 0.01 {
			log("validator's power was increased for more than 1%")
			return true
		}
	}

	// we also need to see if validators added or removed any external chain info.
	// If they did, then this change is also considered to be worthy.
	for i := 0; i < len(sortedCurrent); i++ {
		currentVal, newVal := sortedCurrent[i], sortedNew[i]
		if len(currentVal.ExternalChainInfos) != len(newVal.ExternalChainInfos) {
			log("validator's external chain info sets have changed")
			return true
		}

		keyFnc := func(acc *types.ExternalChainInfo) string {
			return fmt.Sprintf("%s-%s-%s", acc.GetChainReferenceID(), acc.GetChainType(), acc.GetAddress())
		}

		currentMap := slice.MakeMapKeys(currentVal.ExternalChainInfos, keyFnc)
		newMap := slice.MakeMapKeys(newVal.ExternalChainInfos, keyFnc)

		for _, currv := range currentMap {
			newv, ok := newMap[keyFnc(currv)]
			if !ok {
				log("validator changed some of the external address")
				return true
			}

			if len(newv.Traits) != len(currv.Traits) {
				log("validator changed some of its traits")
				return true
			}

			currentTraitMap := make(map[string]struct{})
			newTraitMap := make(map[string]struct{})
			for _, v := range currv.Traits {
				currentTraitMap[v] = struct{}{}
			}
			for _, v := range newv.Traits {
				newTraitMap[v] = struct{}{}
			}

			for k := range currentTraitMap {
				if _, fnd := newTraitMap[k]; !fnd {
					log("validator changed some of its traits")
					return true
				}
			}
		}
	}

	return false
}

func (k Keeper) GetUnjailedValidators(ctx sdk.Context) []stakingtypes.ValidatorI {
	validators := []stakingtypes.ValidatorI{}
	k.staking.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) bool {
		if !val.IsJailed() {
			validators = append(validators, val)
		}
		return false
	})

	return validators
}

// ValidatorSupportsAllChains returns true if the validator supports all chains in the keeper
func (k Keeper) ValidatorSupportsAllChains(ctx sdk.Context, validatorAddress sdk.ValAddress) bool {
	valSupportedChains, err := k.GetValidatorChainInfos(ctx, validatorAddress)
	if err != nil {
		k.Logger(ctx).Error("Unable to get supported chains for validator",
			"validator-address",
			validatorAddress.String(),
		)
		return false
	}

	valSupportedChainReferenceIDs := make([]string, len(valSupportedChains))
	for i, v := range valSupportedChains {
		valSupportedChainReferenceIDs[i] = v.GetChainReferenceID()
	}

	missingChains, err := k.EvmKeeper.MissingChains(ctx, valSupportedChainReferenceIDs)
	if err != nil {
		k.Logger(ctx).With("error", err).Error("error checking missing chains for validator",
			"validator-address",
			validatorAddress.String())
	}
	return len(missingChains) == 0
}

// createNewSnapshot builds a current snapshot of validators.
func (k Keeper) createNewSnapshot(ctx sdk.Context) (*types.Snapshot, error) {
	validators := []stakingtypes.ValidatorI{}
	k.staking.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) bool {
		if val.IsBonded() && !val.IsJailed() && k.ValidatorSupportsAllChains(ctx, val.GetOperator()) {
			validators = append(validators, val)
		}
		return false
	})

	snapshot := &types.Snapshot{
		Height:      ctx.BlockHeight(),
		CreatedAt:   ctx.BlockTime(),
		TotalShares: sdk.ZeroInt(),
	}

	for _, val := range validators {
		chainInfo, err := k.GetValidatorChainInfos(ctx, val.GetOperator())
		if err != nil {
			return nil, err
		}
		snapshot.TotalShares = snapshot.TotalShares.Add(val.GetBondedTokens())
		snapshot.Validators = append(snapshot.Validators, types.Validator{
			Address:            val.GetOperator(),
			ShareCount:         val.GetBondedTokens(),
			State:              types.ValidatorState_ACTIVE,
			ExternalChainInfos: chainInfo,
		})
	}

	return snapshot, nil
}

func (k Keeper) setSnapshotAsCurrent(ctx sdk.Context, snapshot *types.Snapshot) error {
	snapStore := k.snapshotStore(ctx)
	newID := k.ider.IncrementNextID(ctx, snapshotIDKey)
	snapshot.Id = newID
	return keeperutil.Save(snapStore, k.cdc, keeperutil.Uint64ToByte(newID), snapshot)
}

func (k Keeper) SetSnapshotOnChain(ctx sdk.Context, snapshotID uint64, chainReferenceID string) error {
	snapStore := k.snapshotStore(ctx)
	snapshot, err := k.FindSnapshotByID(ctx, snapshotID)
	if err != nil {
		return err
	}
	snapshot.Chains = append(snapshot.Chains, chainReferenceID)
	return keeperutil.Save(snapStore, k.cdc, keeperutil.Uint64ToByte(snapshot.Id), snapshot)
}

func (k Keeper) GetLatestSnapshotOnChain(ctx sdk.Context, chainReferenceID string) (*types.Snapshot, error) {
	snapshotId := k.ider.GetLastID(ctx, snapshotIDKey)

	// Walk backwards from the most recent snapshot until we find one for this chainReferenceID
	for {
		snapshot, err := k.FindSnapshotByID(ctx, snapshotId)
		if err != nil {
			return nil, err
		}

		// See if this snapshot is active on this chain
		for _, chain := range snapshot.Chains {
			if chain == chainReferenceID {
				return snapshot, nil
			}
		}

		snapshotId = snapshot.GetId() - 1
		if snapshotId == 0 {
			break
		}
	}

	// If we made it here, we didn't find a snapshot for this chain
	return nil, keeperutil.ErrNotFound.Format(&types.Snapshot{}, snapshotIDKey)
}

// GetCurrentSnapshot returns the currently active snapshot.
func (k Keeper) GetCurrentSnapshot(ctx sdk.Context) (*types.Snapshot, error) {
	snapStore := k.snapshotStore(ctx)
	lastID := k.ider.GetLastID(ctx, snapshotIDKey)
	snapshot, err := keeperutil.Load[*types.Snapshot](snapStore, k.cdc, keeperutil.Uint64ToByte(lastID))
	k.Logger(ctx).Debug("get current snapshot", "last-id", lastID, "snapshot-validator-size", len(snapshot.Validators))
	if errors.Is(err, keeperutil.ErrNotFound) {
		return nil, nil
	}
	return snapshot, err
}

func (k Keeper) FindSnapshotByID(ctx sdk.Context, id uint64) (*types.Snapshot, error) {
	snapStore := k.snapshotStore(ctx)
	return keeperutil.Load[*types.Snapshot](snapStore, k.cdc, keeperutil.Uint64ToByte(id))
}

func (k Keeper) GetValidatorChainInfos(ctx sdk.Context, valAddr sdk.ValAddress) ([]*types.ExternalChainInfo, error) {
	info, err := keeperutil.Load[*types.ValidatorExternalAccounts](
		k.externalChainInfoStore(ctx, valAddr),
		k.cdc,
		[]byte(valAddr.String()),
	)
	if err != nil {
		if whoops.Is(err, keeperutil.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return info.ExternalChainInfo, nil
}

func (k Keeper) GetAllChainInfos(ctx sdk.Context) ([]*types.ValidatorExternalAccounts, error) {
	chainInfoStore := k._externalChainInfoStore(ctx)
	iter := chainInfoStore.Iterator(nil, nil)

	res := []*types.ValidatorExternalAccounts{}
	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		externalAccounts := &types.ValidatorExternalAccounts{}
		err := k.cdc.Unmarshal(bz, externalAccounts)
		if err != nil {
			return nil, err
		}
		res = append(res, externalAccounts)
	}
	return res, nil
}

// GetSigningKey returns a signing key used by the conductor to sign arbitrary messages.
func (k Keeper) GetSigningKey(ctx sdk.Context, valAddr sdk.ValAddress, chainType, chainReferenceID, signedByAddress string) ([]byte, error) {
	externalAccounts, err := k.GetValidatorChainInfos(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	for _, acc := range externalAccounts {
		if acc.ChainReferenceID == chainReferenceID && acc.ChainType == chainType && acc.Address == signedByAddress {
			return acc.Pubkey, nil
		}
	}

	return nil, ErrSigningKeyNotFound.Format(valAddr.String(), chainType, chainReferenceID)
}

// IsJailed returns if the current validator is jailed or not.
func (k Keeper) IsJailed(ctx sdk.Context, val sdk.ValAddress) bool {
	return k.staking.Validator(ctx, val).IsJailed()
}

func (k Keeper) Jail(ctx sdk.Context, valAddr sdk.ValAddress, reason string) error {
	val := k.staking.Validator(ctx, valAddr)
	if val == nil {
		return ErrValidatorWithAddrNotFound.Format(valAddr)
	}
	if val.IsJailed() {
		return ErrValidatorAlreadyJailed.Format(valAddr.String())
	}

	consensusPower := val.GetConsensusPower(k.powerReduction)
	totalConsensusPower := int64(0)
	count := 0
	k.staking.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) bool {
		if val.IsBonded() && !val.IsJailed() {
			totalConsensusPower += val.GetConsensusPower(k.powerReduction)
			count++
		}
		return false
	})

	if count == 1 {
		return ErrCannotJailValidator.Format(valAddr).WrapS("number of active validators would be zero then")
	}

	if float64(consensusPower)/float64(totalConsensusPower) > cJailingNetworkShareProtection {
		return ErrCannotJailValidator.Format(valAddr).WrapS("validator stake holds over %v percent of entire network", cJailingNetworkShareProtection)
	}

	cons, err := val.GetConsAddr()
	if err != nil {
		return err
	}

	err = func() (jailingErr error) {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			switch t := r.(type) {
			case error:
				jailingErr = t
			case string:
				jailingErr = whoops.String(t)
			default:
				panic(r)
			}
		}()
		k.staking.Jail(ctx, cons)
		return
	}()

	if err != nil {
		return err
	}

	k.Logger(ctx).Info("jailing a validator", "val-addr", valAddr, "reason", reason)
	k.jailReasonStore(ctx).Set(valAddr, []byte(reason))
	return nil
}

func (k Keeper) jailReasonStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("jail-reasons"))
}

func (k Keeper) gracePeriodStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("grace-period"))
}

func (k Keeper) unjailedSnapshotStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("unjailed-snapshot"))
}

func (k Keeper) externalChainInfoStore(ctx sdk.Context, val sdk.ValAddress) sdk.KVStore {
	return prefix.NewStore(
		k._externalChainInfoStore(ctx),
		[]byte(
			fmt.Sprintf("val-%s", val.String()),
		),
	)
}

func (k Keeper) _externalChainInfoStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(
		ctx.KVStore(k.storeKey),
		[]byte("external-chain-info"),
	)
}

func (k Keeper) snapshotStore(ctx sdk.Context) sdk.KVStore {
	k.Logger(ctx).Debug("snapshot store", "store-key-name", k.storeKey.Name(), "store-key-string", k.storeKey.String())
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("snapshot"))
}

// SaveModifiedSnapshot is needed for integration tests
func (k Keeper) SaveModifiedSnapshot(ctx sdk.Context, snapshot *types.Snapshot) error {
	snapStore := k.snapshotStore(ctx)
	return keeperutil.Save(snapStore, k.cdc, keeperutil.Uint64ToByte(snapshot.GetId()), snapshot)
}
