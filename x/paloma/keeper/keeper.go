package keeper

import (
	"fmt"
	"sort"
	"strings"

	storetypes "cosmossdk.io/store/types"
	paramtypes "cosmossdk.io/x/params/types"
	"github.com/VolumeFi/whoops"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/paloma/types"
	"golang.org/x/mod/semver"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace
		Valset     types.ValsetKeeper
		Upgrade    types.UpgradeKeeper
		AppVersion string

		ExternalChains []types.ExternalChainSupporterKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	appVersion string,
	valset types.ValsetKeeper,
	upgrade types.UpgradeKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}
	if !strings.HasPrefix(appVersion, "v") {
		appVersion = fmt.Sprintf("v%s", appVersion)
	}

	if !semver.IsValid(appVersion) {
		panic(fmt.Sprintf("provided app version: '%s' it not a valid semver", appVersion))
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		Valset:     valset,
		Upgrade:    upgrade,

		AppVersion: appVersion,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) JailValidatorsWithMissingExternalChainInfos(ctx sdk.Context) error {
	k.Logger(ctx).Info("start jailing validators with invalid external chain infos")
	vals := k.Valset.GetUnjailedValidators(ctx)

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

		sort.Strings(notSupported)

		if len(notSupported) > 0 {
			g.Add(k.Valset.Jail(ctx, val.GetOperator(), fmt.Sprintf(types.JailReasonNotSupportingTheseExternalChains, strings.Join(notSupported, ", "))))
		}
	}

	return g.Return()
}

// CheckChainVersion will exit if the app version and the government proposed
// versions do not match.
func (k Keeper) CheckChainVersion(ctx sdk.Context) {
	// app needs to be in the [major].[minor] space,
	// but needs to be running at least [major].[minor].[patch]
	govVer, govHeight := k.Upgrade.GetLastCompletedUpgrade(ctx)

	if len(govVer) == 0 || govHeight == 0 {
		return
	}

	if !strings.HasPrefix(govVer, "v") {
		govVer = fmt.Sprintf("v%s", govVer)
	}

	abandon := func() {
		panic(fmt.Sprintf("needs to be running at least %s palomad version, but is running with %s", govVer, k.AppVersion))
	}

	switch semver.Compare(semver.MajorMinor(k.AppVersion), semver.MajorMinor(govVer)) {
	case 0:
		// good!
		// still needs to check for the patch versions
	default:
		// if the app's [major].[minor] is too big or too smal, then we need to exit
		abandon()
	}

	// [major].[minor] are the same up until this point

	// let us now compare the patches
	switch semver.Compare(k.AppVersion, govVer) {
	case 0, 1: // appVersion == govVer or appVersion > govVer
		// good!
		return
	case -1: // appVersion < govVer
		// not good :(
		abandon()
		return
	}
}
