package keeper

import (
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/vizualni/whoops"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/x/paloma/types"
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

func (k Keeper) JailValidatorsWithMissingExternalChainInfos(ctx sdk.Context) error {
	k.Logger(ctx).Info("start jailing validators with invalid external chain infos")
	vals := k.Valset.UnjailedValidators(ctx)

	// making a map of chain types and their external chains
	type mapkey [2]string
	mmap := make(map[mapkey]struct{})
	for _, supported := range k.ExternalChains {
		chainType := supported.XChainType()
		for _, cri := range supported.XChainReferenceIDs(ctx) {
			mmap[mapkey{string(chainType), string(cri)}] = struct{}{}
		}
	}

	var g whoops.Group
	for _, val := range vals {
		exts, err := k.Valset.GetValidatorChainInfos(ctx, val.GetOperator())
		if err != nil {
			g.Add(err)
			continue
		}
		valmap := make(map[mapkey]struct{})
		for _, ext := range exts {
			key := mapkey{ext.GetChainType(), ext.GetChainReferenceID()}
			valmap[key] = struct{}{}
		}

		notSupported := []string{}
		for mustExistKey := range mmap {
			if _, ok := valmap[mustExistKey]; !ok {
				// well well well
				notSupported = append(notSupported, fmt.Sprintf("[%s, %s]", mustExistKey[0], mustExistKey[1]))
			}
		}

		if len(notSupported) > 0 {
			g.Add(k.Valset.Jail(ctx, val.GetOperator(), fmt.Sprintf(types.JailReasonNotSupportingTheseExternalChains, strings.Join(notSupported, ", "))))
		}
	}

	return g.Return()
}
