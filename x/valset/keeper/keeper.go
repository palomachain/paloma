package keeper

import (
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

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		paramstore paramtypes.Subspace
		staking    types.StakingKeeper
		ider       keeperutil.IDGenerator
	}
)

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
	// verify that the acc that actually sent the message is a validator

	if len(newChainInfo) > maxNumOfAllowedExternalAccounts {
		return ErrMaxNumberOfExternalAccounts.Format(
			len(newChainInfo),
			maxNumOfAllowedExternalAccounts,
		)
	}

	// TODO: do some logic on the stakingVal's status
	stakingVal := k.staking.Validator(ctx, valAddr)

	if stakingVal == nil {
		return ErrValidatorWithAddrNotFound.Format(valAddr.String())
	}

	store := k.externalChainInfoStore(ctx, valAddr)

	externalAccounts, err := k.getValidatorChainInfos(ctx, valAddr)
	if err != nil {
		return err
	}
	if externalAccounts == nil {
		externalAccounts = []*types.ExternalChainInfo{}
	}

	eq := func(a, b *types.ExternalChainInfo) bool {
		return a.Address == b.Address &&
			a.ChainID == b.ChainID &&
			a.ChainType == b.ChainType
	}

	allExistingChainAccounts, err := k.getAllChainInfos(ctx)
	if err != nil {
		return err
	}

	// TODO: limit the number of external chain accounts (per chain or globally?)
	// O(n^2) to find if new one is already registered

	for _, existingVal := range allExistingChainAccounts {
		for _, existingChainInfo := range existingVal.ExternalChainInfo {
			for _, chain := range newChainInfo {
				if eq(existingChainInfo, chain) {
					return ErrExternalChainAlreadyRegistered.Format(chain.ChainID, chain.Address)
				}
			}
		}
	}

	externalAccounts = append(externalAccounts, newChainInfo...)

	if len(externalAccounts) > maxNumOfAllowedExternalAccounts {
		return ErrMaxNumberOfExternalAccounts.Format(
			len(externalAccounts),
			maxNumOfAllowedExternalAccounts,
		)
	}

	return keeperutil.Save(store, k.cdc, []byte(valAddr.String()), &types.ValidatorExternalAccounts{
		Address:           valAddr,
		ExternalChainInfo: externalAccounts,
	})
}

// TODO: this is not required for the private alpha
func (k Keeper) RemoveExternalChainInfo(ctx sdk.Context) {}

// TriggerSnapshotBuild creates the snapshot of currently active validators that are
// active and registered as conductors.
func (k Keeper) TriggerSnapshotBuild(ctx sdk.Context) error {
	return k.createSnapshot(ctx)
}

// createSnapshot builds a current snapshot of validators.
func (k Keeper) createSnapshot(ctx sdk.Context) error {
	validators := []stakingtypes.ValidatorI{}
	k.staking.IterateBondedValidatorsByPower(ctx, func(_ int64, val stakingtypes.ValidatorI) bool {
		validators = append(validators, val)
		return true
	})

	snapshot := &types.Snapshot{
		Height:      ctx.BlockHeight(),
		CreatedAt:   ctx.BlockTime(),
		TotalShares: sdk.ZeroInt(),
	}

	for _, val := range validators {
		chainInfo, err := k.getValidatorChainInfos(ctx, val.GetOperator())
		if err != nil {
			return err
		}
		snapshot.TotalShares = snapshot.TotalShares.Add(val.GetBondedTokens())
		snapshot.Validators = append(snapshot.Validators, types.Validator{
			Address:            val.GetOperator(),
			ShareCount:         val.GetBondedTokens(),
			State:              types.ValidatorState_ACTIVE,
			ExternalChainInfos: chainInfo,
		})
	}

	return k.setSnapshotAsCurrent(ctx, snapshot)
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

func (k Keeper) getSnapshotByID(ctx sdk.Context, id uint64) (*types.Snapshot, error) {
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
func (k Keeper) GetSigningKey(ctx sdk.Context, valAddr sdk.ValAddress, chainType, chainID, signedByAddress string) ([]byte, error) {
	externalAccounts, err := k.getValidatorChainInfos(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	for _, acc := range externalAccounts {
		if acc.ChainID == chainID && acc.ChainType == chainType && acc.Address == signedByAddress {
			return acc.Pubkey, nil
		}
	}

	return nil, ErrSigningKeyNotFound.Format(valAddr.String(), chainType, chainID)
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
