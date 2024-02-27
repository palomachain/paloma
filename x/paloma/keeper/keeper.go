package keeper

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"cosmossdk.io/core/address"
	cosmosstore "cosmossdk.io/core/store"
	cosmoslog "cosmossdk.io/log"
	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	utilkeeper "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/paloma/types"
	"golang.org/x/mod/semver"
)

type (
	Keeper struct {
		cdc            codec.BinaryCodec
		storeKey       cosmosstore.KVStoreService
		paramstore     paramtypes.Subspace
		Valset         types.ValsetKeeper
		Upgrade        types.UpgradeKeeper
		AppVersion     string
		AddressCodec   address.Codec
		ExternalChains []types.ExternalChainSupporterKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey cosmosstore.KVStoreService,
	ps paramtypes.Subspace,
	appVersion string,
	valset types.ValsetKeeper,
	upgrade types.UpgradeKeeper,
	addressCodec address.Codec,
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
		cdc:          cdc,
		storeKey:     storeKey,
		paramstore:   ps,
		Valset:       valset,
		Upgrade:      upgrade,
		AddressCodec: addressCodec,
		AppVersion:   appVersion,
	}
}

func (k Keeper) Logger(ctx context.Context) cosmoslog.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return liblog.FromSDKLogger(sdkCtx.Logger()).With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) JailValidatorsWithMissingExternalChainInfos(ctx context.Context) error {
	liblog.FromSDKLogger(k.Logger(ctx)).Info("start jailing validators with invalid external chain infos")
	vals := k.Valset.GetUnjailedValidators(ctx)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// making a map of chain types and their external chains
	type mapkey [2]string
	mmap := make(map[mapkey]struct{})
	for _, supported := range k.ExternalChains {
		chainType := supported.XChainType()
		for _, cri := range supported.XChainReferenceIDs(sdkCtx) {
			mmap[mapkey{string(chainType), string(cri)}] = struct{}{}
		}
	}

	var g whoops.Group
	for _, val := range vals {
		valAddr, err := utilkeeper.ValAddressFromBech32(k.AddressCodec, val.GetOperator())
		if err != nil {
			g.Add(err)
			continue
		}
		exts, err := k.Valset.GetValidatorChainInfos(ctx, valAddr)
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
				notSupported = append(notSupported, fmt.Sprintf("[%s, %s]", mustExistKey[0], mustExistKey[1]))
			}
		}

		sort.Strings(notSupported)
		if len(notSupported) > 0 {
			g.Add(k.Valset.Jail(ctx, valAddr, fmt.Sprintf(types.JailReasonNotSupportingTheseExternalChains, strings.Join(notSupported, ", "))))
		}
	}

	return g.Return()
}

// CheckChainVersion will exit if the app version and the government proposed
// versions do not match.
func (k Keeper) CheckChainVersion(ctx context.Context) {
	// app needs to be in the [major].[minor] space,
	// but needs to be running at least [major].[minor].[patch]
	govVer, govHeight, err := k.Upgrade.GetLastCompletedUpgrade(ctx)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err != nil {
		liblog.FromSDKLogger(sdkCtx.Logger()).WithError(err)
	}
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
