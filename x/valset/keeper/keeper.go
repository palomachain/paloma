package keeper

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	keeperutil "github.com/volumefi/cronchain/util/keeper"
	"github.com/volumefi/cronchain/x/valset/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		paramstore paramtypes.Subspace
		staking    types.StakingKeeper
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

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		staking:    staking,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// TODO: not required now
func (k Keeper) PunishValidator(ctx sdk.Context) {}

// TODO: not required now
func (k Keeper) Heartbeat(ctx sdk.Context) {}

// TODO: not required now
// TODO: break this into add, remove
func (k Keeper) updateExternalChainInfo(ctx sdk.Context) {}

func (k Keeper) Register(ctx sdk.Context, valAddr sdk.ValAddress, signingKey cryptotypes.PubKey) error {
	sval := k.staking.Validator(ctx, valAddr)
	if sval == nil {
		return ErrValidatorWithAddrNotFound.Format(valAddr)
	}
	store := k.validatorStore(ctx)

	// check if is already registered! if yes, then error
	if store.Has(valAddr) {
		return ErrValidatorAlreadyRegistered
	}

	val := &types.Validator{
		Address: sval.GetOperator().String(),
		// TODO: add the rest
	}

	// TODO: more logic here
	val.State = types.ValidatorState_ACTIVE

	// save val
	return keeperutil.Save(store, k.cdc, valAddr, val)
}

func (k Keeper) CreateSnapshot(ctx sdk.Context) error {
	valStore := k.validatorStore(ctx)

	// get all registered validators
	validators, err := keeperutil.IterAll[*types.Validator](valStore, k.cdc)
	if err != nil {
		return err
	}

	snapshot := &types.Snapshot{
		Height:      ctx.BlockHeight(),
		CreatedAt:   ctx.BlockTime(),
		TotalShares: sdk.ZeroInt(),
	}

	for _, val := range validators {
		// if val.State != types.ValidatorState_ACTIVE {
		// 	continue
		// }
		snapshot.TotalShares = snapshot.TotalShares.Add(val.ShareCount)
		snapshot.Validators = append(snapshot.Validators, *val)
	}

	return k.setSnapshotAsCurrent(ctx, snapshot)
}

func (k Keeper) setSnapshotAsCurrent(ctx sdk.Context, snapshot *types.Snapshot) error {
	snapStore := k.snapshotStore(ctx)
	return keeperutil.Save(snapStore, k.cdc, []byte("snapshot"), snapshot)
}

func (k Keeper) GetCurrentSnapshot(ctx sdk.Context) (*types.Snapshot, error) {
	snapStore := k.snapshotStore(ctx)
	return keeperutil.Load[*types.Snapshot](snapStore, k.cdc, []byte("snapshot"))
}

func (k Keeper) GetSigningKey(ctx sdk.Context, valAddr sdk.ValAddress) cryptotypes.PubKey {
	val := k.staking.Validator(ctx, valAddr)
	if val == nil {
		return nil
	}
	pk, _ := val.ConsPubKey()
	return pk
}

func (k Keeper) validatorStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("validators"))
}

func (k Keeper) snapshotStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("snapshot"))
}
