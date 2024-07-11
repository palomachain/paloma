package keeper

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"cosmossdk.io/core/address"
	cosmosstore "cosmossdk.io/core/store"
	cosmoslog "cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	utilkeeper "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/paloma/types"
	"golang.org/x/mod/semver"
)

type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       cosmosstore.KVStoreService
	paramstore     paramtypes.Subspace
	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	feegrantKeeper types.FeegrantKeeper
	Valset         types.ValsetKeeper
	Upgrade        types.UpgradeKeeper
	AppVersion     string
	AddressCodec   address.Codec
	ExternalChains []types.ExternalChainSupporterKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey cosmosstore.KVStoreService,
	ps paramtypes.Subspace,
	appVersion string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	feegrantKeeper types.FeegrantKeeper,
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
		cdc:            cdc,
		storeKey:       storeKey,
		paramstore:     ps,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		feegrantKeeper: feegrantKeeper,
		Valset:         valset,
		Upgrade:        upgrade,
		AddressCodec:   addressCodec,
		AppVersion:     appVersion,
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

func (k Keeper) lightNodeClientStore(ctx context.Context) storetypes.KVStore {
	s := runtime.KVStoreAdapter(k.storeKey.OpenKVStore(ctx))
	return prefix.NewStore(s, types.LightNodeClientKeyPrefix)
}

func (k Keeper) AllLightNodeClientFunds(
	ctx context.Context,
) ([]*types.LightNodeClientFunds, error) {
	st := k.lightNodeClientStore(ctx)
	_, all, err := keeperutil.IterAll[*types.LightNodeClientFunds](st, k.cdc)
	return all, err
}

func (k Keeper) getLightNodeClientFunds(
	ctx context.Context,
	addr string,
) (*types.LightNodeClientFunds, error) {
	st := k.lightNodeClientStore(ctx)
	return keeperutil.Load[*types.LightNodeClientFunds](st, k.cdc, []byte(addr))
}

func (k Keeper) CreateLightNodeClientAccount(
	ctx context.Context,
	authAddr, acctAddr string,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	beginTime := sdkCtx.BlockTime()
	endTime := beginTime.AddDate(2, 0, 0) // Vesting for two years

	acct, err := k.accountKeeper.AddressCodec().StringToBytes(acctAddr)
	if err != nil {
		return err
	}

	// First, check if the vesting account already exists
	if acc := k.accountKeeper.GetAccount(ctx, acct); acc != nil {
		return errors.New("account already exists")
	}

	// Get the funds allocated to the authorization account
	funds, err := k.getLightNodeClientFunds(ctx, authAddr)
	if err != nil {
		return err
	}
	amount := sdk.Coins{funds.Amount}

	// Create a basic vesting account
	// It will vest until `endTime`, which is set two years from now
	baseAccount := authtypes.NewBaseAccountWithAddress(acct)
	baseAccount = k.accountKeeper.NewAccount(ctx, baseAccount).(*authtypes.BaseAccount)
	baseVestingAccount, err := vestingtypes.NewBaseVestingAccount(baseAccount, amount, endTime.Unix())
	if err != nil {
		return err
	}

	// Create a continuous vesting account, starting from now
	vestingAccount := vestingtypes.NewContinuousVestingAccountRaw(baseVestingAccount, beginTime.Unix())

	k.accountKeeper.SetAccount(ctx, vestingAccount)

	// Seed the account from the module fund
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, acct, amount)
	if err != nil {
		return err
	}

	// Finally, set the module as fee granter for the new account
	moduleAccount := k.accountKeeper.GetModuleAddress(types.ModuleName)

	// TODO - we might want to set the spend limit to a safe value
	allowance := &feegrant.BasicAllowance{
		SpendLimit: nil,      // Unlimited spend
		Expiration: &endTime, // Expires when it finishes vesting
	}

	return k.feegrantKeeper.GrantAllowance(ctx, moduleAccount, acct, allowance)
}
