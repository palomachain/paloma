package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/vizualni/whoops"
)

const (
	snapshotIDKey                   = "snapshot-id"
	maxNumOfAllowedExternalAccounts = 100
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	memKey     sdk.StoreKey
	paramstore paramtypes.Subspace
	staking    types.StakingKeeper
	ider       keeperutil.IDGenerator

	SnapshotListeners []types.OnSnapshotBuiltListener
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	staking types.StakingKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	k := &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		staking:    staking,
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

func (k Keeper) CanAcceptValidator(ctx sdk.Context, valAddr sdk.ValAddress) error {
	stakingVal := k.staking.Validator(ctx, valAddr)

	if stakingVal == nil {
		return ErrValidatorWithAddrNotFound.Format(valAddr.String())
	}

	if stakingVal.IsJailed() {
		return ErrValidatorCannotBePigeon.Format(valAddr.String()).WrapS("validator is jailed")
	}

	if !stakingVal.IsBonded() {
		return ErrValidatorCannotBePigeon.Format(valAddr.String()).WrapS("validator is not bonded")
	}

	return nil
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

	allExistingChainAccounts, err := k.getAllChainInfos(ctx)
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
	snapshot, err := k.createSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	for _, listener := range k.SnapshotListeners {
		listener.OnSnapshotBuilt(ctx, snapshot)
	}

	return snapshot, err
}

// createSnapshot builds a current snapshot of validators.
func (k Keeper) createSnapshot(ctx sdk.Context) (*types.Snapshot, error) {
	validators := []stakingtypes.ValidatorI{}
	k.staking.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) bool {
		if val.IsBonded() {
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
		chainInfo, err := k.getValidatorChainInfos(ctx, val.GetOperator())
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

	err := k.setSnapshotAsCurrent(ctx, snapshot)
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (k Keeper) setSnapshotAsCurrent(ctx sdk.Context, snapshot *types.Snapshot) error {
	snapStore := k.snapshotStore(ctx)
	newID := k.ider.IncrementNextID(ctx, snapshotIDKey)
	snapshot.Id = newID
	return keeperutil.Save(snapStore, k.cdc, keeperutil.Uint64ToByte(newID), snapshot)
}

// GetCurrentSnapshot returns the currently active snapshot.
func (k Keeper) GetCurrentSnapshot(ctx sdk.Context) (*types.Snapshot, error) {
	snapStore := k.snapshotStore(ctx)
	lastID := k.ider.GetLastID(ctx, snapshotIDKey)
	return keeperutil.Load[*types.Snapshot](snapStore, k.cdc, keeperutil.Uint64ToByte(lastID))
}

func (k Keeper) FindSnapshotByID(ctx sdk.Context, id uint64) (*types.Snapshot, error) {
	snapStore := k.snapshotStore(ctx)
	return keeperutil.Load[*types.Snapshot](snapStore, k.cdc, keeperutil.Uint64ToByte(id))
}

func (k Keeper) getValidatorChainInfos(ctx sdk.Context, valAddr sdk.ValAddress) ([]*types.ExternalChainInfo, error) {
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

func (k Keeper) getAllChainInfos(ctx sdk.Context) ([]*types.ValidatorExternalAccounts, error) {
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
	externalAccounts, err := k.getValidatorChainInfos(ctx, valAddr)
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

func (k Keeper) validatorStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("validators"))
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
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("snapshot"))
}
