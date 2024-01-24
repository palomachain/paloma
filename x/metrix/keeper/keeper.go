package keeper

import (
	"context"
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/palomachain/paloma/x/metrix/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

var _ valsettypes.OnSnapshotBuiltListener = &Keeper{}

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		paramstore paramtypes.Subspace
		slashing   types.SlashingKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	slashing types.SlashingKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:        cdc,
		paramstore: ps,
		slashing:   slashing,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// OnConsensusMessageAttested implements types.OnConsensusMessageAttestedListener.
func (Keeper) OnConsensusMessageAttested(context.Context, types.MessageAttestedEvent) {
	// 4. success rate
	// 5. runtime
}

// OnSnapshotBuilt implements types.OnSnapshotBuiltListener.
func (*Keeper) OnSnapshotBuilt(ctx sdk.Context, snapshot *valsettypes.Snapshot) {
	// 1. Fee?
	// 2. Feature sets
}

func (k *Keeper) UpdateUptime(ctx sdk.Context) {
	// use slashing keeper
	// 3. uptime

	k.slashing.IterateValidatorSigningInfos(ctx, func(consAddr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		return false
	})
}
