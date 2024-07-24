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
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	feegrantmodule "cosmossdk.io/x/feegrant"
	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
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
	bondDenom      string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey cosmosstore.KVStoreService,
	ps paramtypes.Subspace,
	appVersion string,
	bondDenom string,
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
		bondDenom:      bondDenom,
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
		valAddr, err := keeperutil.ValAddressFromBech32(k.AddressCodec, val.GetOperator())
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

func (k Keeper) Store(ctx context.Context) storetypes.KVStore {
	return runtime.KVStoreAdapter(k.storeKey.OpenKVStore(ctx))
}

func (k Keeper) lightNodeClientLicenseStore(
	ctx context.Context,
) storetypes.KVStore {
	s := runtime.KVStoreAdapter(k.storeKey.OpenKVStore(ctx))
	return prefix.NewStore(s, types.LightNodeClientLicenseKeyPrefix)
}

func (k Keeper) LightNodeClientFeegranter(
	ctx context.Context,
) (*types.LightNodeClientFeegranter, error) {
	st := k.Store(ctx)
	key := types.LightNodeClientFeegranterKey

	return keeperutil.Load[*types.LightNodeClientFeegranter](st, k.cdc, key)
}

func (k Keeper) SetLightNodeClientFeegranter(
	ctx context.Context,
	acct sdk.AccAddress,
) error {
	st := k.Store(ctx)
	key := types.LightNodeClientFeegranterKey

	feegranter := &types.LightNodeClientFeegranter{
		Account: acct,
	}

	return keeperutil.Save(st, k.cdc, key, feegranter)
}

func (k Keeper) LightNodeClientFunder(
	ctx context.Context,
) (*types.LightNodeClientFunder, error) {
	st := k.Store(ctx)
	key := types.LightNodeClientFunderKey

	return keeperutil.Load[*types.LightNodeClientFunder](st, k.cdc, key)
}

func (k Keeper) SetLightNodeClientFunder(
	ctx context.Context,
	acct sdk.AccAddress,
) error {
	st := k.Store(ctx)
	key := types.LightNodeClientFunderKey

	funder := &types.LightNodeClientFunder{
		Account: acct,
	}

	return keeperutil.Save(st, k.cdc, key, funder)
}

func (k Keeper) AllLightNodeClientLicenses(
	ctx context.Context,
) ([]*types.LightNodeClientLicense, error) {
	st := k.lightNodeClientLicenseStore(ctx)
	_, all, err := keeperutil.IterAll[*types.LightNodeClientLicense](st, k.cdc)
	return all, err
}

func (k Keeper) SetLightNodeClientLicense(
	ctx context.Context,
	addr string,
	license *types.LightNodeClientLicense,
) error {
	st := k.lightNodeClientLicenseStore(ctx)
	return keeperutil.Save(st, k.cdc, []byte(addr), license)
}

func (k Keeper) GetLightNodeClientLicense(
	ctx context.Context,
	addr string,
) (*types.LightNodeClientLicense, error) {
	st := k.lightNodeClientLicenseStore(ctx)
	return keeperutil.Load[*types.LightNodeClientLicense](st, k.cdc, []byte(addr))
}

func (k Keeper) CreateLightNodeClientLicense(
	ctx context.Context,
	creatorAddr, clientAddr string,
	amount sdk.Coin,
	vestingMonths uint32,
) error {
	creatorAcct, err := sdk.AccAddressFromBech32(creatorAddr)
	if err != nil {
		return err
	}

	if sdk.VerifyAddressFormat(creatorAcct) != nil || !amount.IsValid() {
		return types.ErrInvalidParameters
	}

	// Check if license exists already
	_, err = k.GetLightNodeClientLicense(ctx, clientAddr)
	if err == nil {
		return types.ErrLicenseExists
	} else if !errors.Is(err, keeperutil.ErrNotFound) {
		// Any errors other than not found and we bail
		return err
	}

	// The client account name, to create a new base account and/or to set up
	// the feegrant
	acct, err := k.accountKeeper.AddressCodec().StringToBytes(clientAddr)
	if err != nil {
		return err
	}

	// If the account already exists, we can't move on
	if k.accountKeeper.HasAccount(ctx, acct) {
		return types.ErrAccountExists
	}

	license := &types.LightNodeClientLicense{
		ClientAddress: clientAddr,
		Amount:        amount,
		VestingMonths: vestingMonths,
	}

	// We're here for the first time, so we create the base account now, as
	// it will be used to register the light node client and start vesting
	baseAccount := authtypes.NewBaseAccountWithAddress(acct)
	baseAccount = k.accountKeeper.NewAccount(ctx, baseAccount).(*authtypes.BaseAccount)
	k.accountKeeper.SetAccount(ctx, baseAccount)

	// Lock new coins in module
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAcct,
		types.ModuleName, sdk.Coins{amount})
	if err != nil {
		return err
	}

	return k.SetLightNodeClientLicense(ctx, clientAddr, license)
}

// CreateLightNodeClientLicenseWithFeegrant is used by the skyway module when
// processing a light node sale event to create a new license with feegrant
func (k Keeper) CreateLightNodeClientLicenseWithFeegrant(
	ctx context.Context,
	creatorAddr, clientAddr string,
	amount math.Int,
	vestingMonths uint32,
) error {
	feegranter, err := k.LightNodeClientFeegranter(ctx)
	if err != nil {
		if errors.Is(err, keeperutil.ErrNotFound) {
			return types.ErrNoFeegranter
		}

		return err
	}

	allowance := &feegrantmodule.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewCoin(k.bondDenom, math.NewInt(1_000_000))),
		Expiration: nil, // Unlimited time
	}

	acct, err := k.accountKeeper.AddressCodec().StringToBytes(clientAddr)
	if err != nil {
		return err
	}

	err = k.feegrantKeeper.GrantAllowance(ctx, feegranter.Account, acct, allowance)
	if err != nil {
		return err
	}

	coin := sdk.NewCoin(k.bondDenom, amount)

	return k.CreateLightNodeClientLicense(ctx, creatorAddr, clientAddr, coin,
		vestingMonths)
}

func (k Keeper) CreateLightNodeClientAccount(
	ctx context.Context,
	addr string,
) error {
	// Get the license allocated to the account
	license, err := k.GetLightNodeClientLicense(ctx, addr)
	if err != nil {
		if errors.Is(err, keeperutil.ErrNotFound) {
			return types.ErrNoLicense
		}

		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	beginTime := sdkCtx.BlockTime()
	endTime := beginTime.AddDate(0, int(license.VestingMonths), 0)

	acct, err := k.accountKeeper.AddressCodec().StringToBytes(addr)
	if err != nil {
		return err
	}

	amount := sdk.Coins{license.Amount}

	// Get the existing base account and turn it into a vesting account
	// It will vest until `endTime`, which is set to `vestingMonths` from now
	baseAccount, ok := k.accountKeeper.GetAccount(ctx, acct).(*authtypes.BaseAccount)
	if !ok {
		return types.ErrNoAccount
	}

	baseVestingAccount, err := vestingtypes.NewBaseVestingAccount(baseAccount,
		amount, endTime.Unix())
	if err != nil {
		return err
	}

	// Create a continuous vesting account, starting from now
	vestingAccount := vestingtypes.NewContinuousVestingAccountRaw(
		baseVestingAccount, beginTime.Unix())

	k.accountKeeper.SetAccount(ctx, vestingAccount)

	// Seed the account from the module account
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, acct, amount)
	if err != nil {
		return err
	}

	// Delete the light node client funds
	k.lightNodeClientLicenseStore(ctx).Delete([]byte(addr))

	return nil
}
