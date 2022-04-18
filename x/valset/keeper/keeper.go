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
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,

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

func (k Keeper) Register(ctx sdk.Context, valAddr sdk.ValAddress, chainInfos []*types.ExternalChainInfo) error {
	store := k.validatorStore(ctx)

	// check if is already registered! if yes, then error
	if store.Has(valAddr) {
		return ErrValidatorAlreadyRegistered
	}

	val := &types.Validator{}

	// TODO: more logic here
	val.State = types.ValidatorState_ACTIVE

	// do a bit of logic
	if len(chainInfos) == 0 {
		val.State = types.ValidatorState_NONE
	}

	// save val
	return keeperutil.Save(store, k.cdc, valAddr, val)
}

func (k Keeper) CreateSnapshot(ctx sdk.Context) error {
	valStore := k.validatorStore(ctx)

	validators, err := keeperutil.IterAll[*types.Validator](valStore, k.cdc)
	if err != nil {
		return err
	}

	snapshot := &types.Snapshot{
		Height:    ctx.BlockHeight(),
		CreatedAt: ctx.BlockTime(),
	}
	for _, val := range validators {
		if val.State != types.ValidatorState_ACTIVE {
			continue
		}
		snapshot.TotalShares = snapshot.TotalShares.Add(val.ShareCount)
		snapshot.Validators = append(snapshot.Validators, *val)
	}

	return k.setCurrentSnapshot(ctx, snapshot)
}

func (k Keeper) setCurrentSnapshot(ctx sdk.Context, snapshot *types.Snapshot) error {
	snapStore := k.snapshotStore(ctx)
	bytez, err := k.cdc.Marshal(snapshot)
	if err != nil {
		return err
	}
	snapStore.Set([]byte("snapshot"), bytez)
	return nil
}

func (k Keeper) GetCurrentSnapshot(ctx sdk.Context) (*types.Snapshot, error) {
	snapStore := k.snapshotStore(ctx)
	bytez := snapStore.Get([]byte("snapshot"))
	var snapshot *types.Snapshot
	err := k.cdc.Unmarshal(bytez, snapshot)
	if err != nil {
		return nil, err
	}
	return snapshot, nil

}

func (k Keeper) GetValidatorPubKey(ctx sdk.Context) cryptotypes.PubKey {
	return nil
}

func (k Keeper) validatorStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("validators"))
}

func (k Keeper) snapshotStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("snapshot"))
}
