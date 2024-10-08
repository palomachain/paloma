package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	keeperutil "github.com/palomachain/paloma/v2/util/keeper"
	"github.com/palomachain/paloma/v2/util/libcons"
	"github.com/palomachain/paloma/v2/util/liblog"
	"github.com/palomachain/paloma/v2/x/consensus/types"
	metrixtypes "github.com/palomachain/paloma/v2/x/metrix/types"
)

type FeeProvider interface {
	GetCombinedFeesForRelay(ctx context.Context, valAddress sdk.ValAddress, chainReferenceID string) (*types.MessageFeeSettings, error)
}

type (
	Keeper struct {
		cdc        codec.Codec
		storeKey   store.KVStoreService
		paramstore paramtypes.Subspace

		ider keeperutil.IDGenerator

		valset types.ValsetKeeper

		registry                   *registry
		evmKeeper                  types.EvmKeeper
		consensusChecker           *libcons.ConsensusChecker
		feeProvider                FeeProvider
		onMessageAttestedListeners []metrixtypes.OnConsensusMessageAttestedListener
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey store.KVStoreService,
	ps paramtypes.Subspace,
	valsetKeeper types.ValsetKeeper,
	reg *registry,
	fp FeeProvider,
) *Keeper {
	k := &Keeper{
		cdc:                        cdc,
		storeKey:                   storeKey,
		paramstore:                 ps,
		valset:                     valsetKeeper,
		registry:                   reg,
		feeProvider:                fp,
		onMessageAttestedListeners: make([]metrixtypes.OnConsensusMessageAttestedListener, 0),
	}
	ider := keeperutil.NewIDGenerator(k, nil)
	k.ider = ider
	k.consensusChecker = libcons.New(k.valset.GetCurrentSnapshot, k.cdc)

	return k
}

func (k *Keeper) AddMessageConsensusAttestedListener(l metrixtypes.OnConsensusMessageAttestedListener) {
	k.onMessageAttestedListeners = append(k.onMessageAttestedListeners, l)
}

func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return liblog.FromSDKLogger(sdkCtx.Logger()).With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) Store(ctx context.Context) storetypes.KVStore {
	return runtime.KVStoreAdapter(k.storeKey.OpenKVStore(ctx))
}

func (k Keeper) ModuleName() string {
	return types.ModuleName
}

func (k *Keeper) LateInject(evm types.EvmKeeper) {
	k.evmKeeper = evm
}
