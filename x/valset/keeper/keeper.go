package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
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

func (k Keeper) PunishValidator(ctx sdk.Context) {}

func (k Keeper) Heartbeat(ctx sdk.Context) {}

func (k Keeper) Register(ctx sdk.Context) {}

// TODO: break this into add, remove
func (k Keeper) UpdateExternalChainInfo(ctx sdk.Context) {}

func (k Keeper) CreateSnapshot(ctx sdk.Context) {}

func (k Keeper) GetCurrentSnapshot(ctx sdk.Context) {}

func (k Keeper) GetValPubKey(ctx sdk.Context) {}

func (k Keeper) validatorStore(ctx sdk.Context) {

}

func (k Keeper) snapshotStore(ctx sdk.Context) {

}
