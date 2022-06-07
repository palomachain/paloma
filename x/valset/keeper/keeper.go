package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/vizualni/whoops"
)

const (
	externalChainInfoIDKey = "external-chain-info"
	snapshotIDKey          = "snapshot-id"
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
func (k Keeper) addExternalChainInfo(ctx sdk.Context, msg *types.MsgAddExternalChainInfoForValidator) error {
	// verify that the acc that actually sent the message is a validator
	validatorStore := k.validatorStore(ctx)

	accAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return err
	}

	valAddr := sdk.ValAddress(accAddr)

	validator, err := keeperutil.Load[*types.Validator](
		validatorStore,
		k.cdc,
		[]byte(valAddr.String()),
	)
	if err != nil {
		if whoops.Is(err, keeperutil.ErrNotFound) {
			return ErrMustUseAccAssociatedWithValidators
		}
		return err
	}
	// TODO: limit the number of external chain accounts (per chain or globally?)

	// O(n^2) to find if new one is already registered
	for _, newChainInfo := range msg.ChainInfos {
		for _, existingChainInfo := range validator.ExternalChainInfos {
			if existingChainInfo.ChainID == newChainInfo.ChainID && existingChainInfo.Address == newChainInfo.Address {
				return ErrExternalChainAlreadyRegistered.Format(newChainInfo.ChainID, newChainInfo.Address)
			}
		}
	}
	for _, newChainInfo := range msg.ChainInfos {
		id := k.ider.IncrementNextID(ctx, externalChainInfoIDKey)
		validator.ExternalChainInfos = append(validator.ExternalChainInfos, &types.ExternalChainInfo{
			ID:      id,
			ChainID: newChainInfo.ChainID,
			Address: newChainInfo.Address,
			Pubkey:  newChainInfo.GetPubKey(),
		})
	}

	err = keeperutil.Save(validatorStore, k.cdc, []byte(msg.Creator), validator)

	if err != nil {
		return err
	}

	return nil
}

// TODO: this is not required for the private alpha
func (k Keeper) RemoveExternalChainInfo(ctx sdk.Context) {}

// Register registers the validator as being a part of a conductor's network.
func (k Keeper) Register(ctx sdk.Context, msg *types.MsgRegisterConductor) error {

	valAddr, err := sdk.ValAddressFromBech32(msg.ValAddr)
	if err != nil {
		return err
	}

	sval := k.staking.Validator(ctx, valAddr)
	if sval == nil {
		return ErrValidatorWithAddrNotFound.Format(valAddr)
	}

	// TODO: making the assumption that the pub key is of secp256k1 type.
	pk := secp256k1.PubKey(msg.PubKey)

	if !pk.VerifySignature(msg.PubKey, msg.SignedPubKey) {
		return ErrPublicKeyOrSignatureIsInvalid
	}

	store := k.validatorStore(ctx)

	// check if is already registered! if yes, then error
	if store.Has(valAddr) {
		return ErrValidatorAlreadyRegistered
	}

	val := &types.Validator{
		Address:       sval.GetOperator(),
		SignerAddress: whoops.Must(sdk.AccAddressFromBech32(msg.Creator)),
		// TODO: add the rest
	}

	// TODO: more logic here
	val.State = types.ValidatorState_ACTIVE

	// save val
	return keeperutil.Save(store, k.cdc, valAddr, val)
}

// TriggerSnapshotBuild creates the snapshot of currently active validators that are
// active and registered as conductors.
func (k Keeper) TriggerSnapshotBuild(ctx sdk.Context) error {
	return k.createSnapshot(ctx)
}

// createSnapshot builds a current snapshot of validators.
func (k Keeper) createSnapshot(ctx sdk.Context) error {
	// TODO: check if there is a need for snapshots being incremental and keeping the historical versions.
	valStore := k.validatorStore(ctx)

	// get all registered validators
	_, validators, err := keeperutil.IterAll[*types.Validator](valStore, k.cdc)
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
	newID := k.ider.IncrementNextID(ctx, snapshotIDKey)
	return keeperutil.Save(snapStore, k.cdc, keeperutil.Uint64ToByte(newID), snapshot)
}

// GetCurrentSnapshot returns the currently active snapshot.
func (k Keeper) GetCurrentSnapshot(ctx sdk.Context) (*types.Snapshot, error) {
	snapStore := k.snapshotStore(ctx)
	lastID := k.ider.GetLastID(ctx, snapshotIDKey)
	return keeperutil.Load[*types.Snapshot](snapStore, k.cdc, keeperutil.Uint64ToByte(lastID))
}

func (k Keeper) getValidator(ctx sdk.Context, valAddr sdk.ValAddress) (*types.Validator, error) {
	return keeperutil.Load[*types.Validator](k.validatorStore(ctx), k.cdc, valAddr)
}

// GetSigningKey returns a signing key used by the conductor to sign arbitrary messages.
func (k Keeper) GetSigningKey(ctx sdk.Context, valAddr sdk.ValAddress, chainID string) crypto.PubKey {
	validator, err := k.getValidator(ctx, valAddr)
	if err != nil {
		return nil
	}

	return secp256k1.PubKey(validator.PubKey)
}

func (k Keeper) validatorStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("validators"))
}

func (k Keeper) snapshotStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("snapshot"))
}
