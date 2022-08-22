package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/vizualni/whoops"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/x/paloma/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		paramstore paramtypes.Subspace
		Valset     types.ValsetKeeper

		ExternalChains []types.ExternalChainSupporterKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	valset types.ValsetKeeper,
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
		Valset:     valset,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) JailValidatorsWithInvalidExternalChainInfos(ctx sdk.Context) error {
	k.Logger(ctx).Info("start jailing validators with invalid external chain infos")
	vals := k.Valset.UnjailedValidators(ctx)
	jail := func(val stakingtypes.ValidatorI, reason string) {
		k.Valset.Jail(ctx, val.GetOperator(), reason)
	}

	type mapkey [2]string
	mmap := make(map[mapkey]struct{})
	for _, supported := range k.ExternalChains {
		for _, cri := range supported.ChainReferenceIDs(ctx) {
			mmap[mapkey{supported.ChainType(ctx), cri}] = struct{}{}
		}
	}

	var g whoops.Group
	for _, val := range vals {
		exts, err := k.Valset.GetValidatorChainInfos(ctx, val.GetOperator())
		if err != nil {
			g.Add(err)
			continue
		}
		if exts == nil {
			jail(val, "not supporting any external chain")
			continue
		}

		notSupported := []string{}

		for _, ext := range exts {
			key := mapkey{ext.GetChainType(), ext.GetChainReferenceID()}
			_, ok := mmap[key]
			if !ok {
				// well well well
				notSupported = append(notSupported, fmt.Sprintf("[%s, %s]", ext.GetChainType(), ext.GetChainReferenceID()))
			}
		}

		if len(notSupported) > 0 {
			jail(val, fmt.Sprintf("not supporthing these external chains: %s", notSupported))
		}
	}

	return g.Return()
}
