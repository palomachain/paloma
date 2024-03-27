package app

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	consensusmoduletypes "github.com/palomachain/paloma/x/consensus/types"
	gravitymoduletypes "github.com/palomachain/paloma/x/gravity/types"
	metrixmoduletypes "github.com/palomachain/paloma/x/metrix/types"
)

var minCommissionRate = math.LegacyMustNewDecFromStr("0.05")

// UpdateMinCommissionRate update minimum commission rate param.
func UpdateMinCommissionRate(ctx sdk.Context, keeper stakingkeeper.Keeper) (math.LegacyDec, error) {
	params, err := keeper.GetParams(ctx)
	if err != nil {
		return math.LegacyDec{}, err
	}
	params.MinCommissionRate = minCommissionRate

	if err := keeper.SetParams(ctx, params); err != nil {
		return math.LegacyDec{}, err
	}

	return minCommissionRate, nil
}

// SetMinimumCommissionRate updates the commission rate for validators
// whose current commission rate is lower than the new minimum commission rate.
func SetMinimumCommissionRate(ctx sdk.Context, keeper stakingkeeper.Keeper, minCommissionRate math.LegacyDec) error {
	validators, err := keeper.GetAllValidators(ctx)
	if err != nil {
		return err
	}
	for _, validator := range validators {
		if validator.Commission.Rate.IsNil() || validator.Commission.Rate.LT(minCommissionRate) {
			valAddr, err := keeperutil.ValAddressFromBech32(keeper.ValidatorAddressCodec(), validator.GetOperator())
			if err != nil {
				return err
			}
			if err := keeper.Hooks().BeforeValidatorModified(ctx, valAddr); err != nil {
				return err
			}

			validator.Commission.Rate = minCommissionRate
			validator.Commission.UpdateTime = ctx.BlockTime()

			err = keeper.SetValidator(ctx, validator)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (app *App) RegisterUpgradeHandlers(semverVersion string) {
	baseAppLegacySS := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

	app.UpgradeKeeper.SetUpgradeHandler(
		semverVersion,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			// Migrate CometBFT consensus parameters from x/params module to a
			// dedicated x/consensus module.
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			err := baseapp.MigrateParams(sdkCtx, baseAppLegacySS, &app.ConsensusParamsKeeper.ParamsStore)
			if err != nil {
				return nil, err
			}

			// TODO: We may need to execute ibc-go v6 migrations but importing ibc-go
			// v6 will fail using Cosmos SDK v0.47.x.
			//
			// if err := v6.MigrateICS27ChannelCapability(ctx, app.cdc, app.keys[capabilitytypes.StoreKey], app.CapabilityKeeper, ""); err != nil {
			// 	return nil, err
			// }

			// OPTIONAL: prune expired tendermint consensus states to save storage space
			if _, err := ibctmmigrations.PruneExpiredConsensusStates(sdkCtx, app.appCodec, app.IBCKeeper.ClientKeeper); err != nil {
				return nil, err
			}

			// save oldIcaVersion, so we can skip icahost.InitModule in longer term tests.
			oldIcaVersion := fromVM[icatypes.ModuleName]

			// Add Interchain Accounts host module
			// set the ICS27 consensus version so InitGenesis is not run
			fromVM[icatypes.ModuleName] = app.ModuleManager.Modules[icatypes.ModuleName].(module.HasConsensusVersion).ConsensusVersion()

			// create ICS27 Host submodule params, host module not enabled.
			hostParams := icahosttypes.Params{
				HostEnabled:   false,
				AllowMessages: []string{},
			}

			mod, found := app.ModuleManager.Modules[icatypes.ModuleName]
			if !found {
				return nil, fmt.Errorf("module %s is not in the module manager", icatypes.ModuleName)
			}

			icaMod, ok := mod.(ica.AppModule)
			if !ok {
				return nil, fmt.Errorf("expected module %s to be type %T, got %T", icatypes.ModuleName, ica.AppModule{}, mod)
			}

			// skip InitModule in upgrade tests after the upgrade has gone through.
			if oldIcaVersion != fromVM[icatypes.ModuleName] {
				icaMod.InitModule(sdkCtx, icacontrollertypes.DefaultParams(), hostParams)
			}

			vm, err := app.ModuleManager.RunMigrations(ctx, app.configurator, fromVM)
			if err != nil {
				return vm, err
			}

			minCommissionRate, err := UpdateMinCommissionRate(sdkCtx, *app.StakingKeeper)
			if err != nil {
				return vm, err
			}

			err = SetMinimumCommissionRate(sdkCtx, *app.StakingKeeper, minCommissionRate)
			if err != nil {
				return vm, err
			}

			// Try to build a snapshot.  Fails silently if unworthy
			_, err = app.ValsetKeeper.TriggerSnapshotBuild(ctx)
			if err != nil {
				return vm, err
			}

			snapshot, err := app.ValsetKeeper.GetCurrentSnapshot(ctx)
			if err != nil {
				return vm, err
			}

			// Publish latest snapshot to all chains
			err = app.EvmKeeper.PublishSnapshotToAllChains(ctx, snapshot, true)
			if err != nil {
				return vm, err
			}

			return vm, nil
		},
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == "v1.0.0" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Renamed: []storetypes.StoreRename{ // x/consensus module renamed to palomaconsensus
				{
					OldKey: consensustypes.ModuleName,
					NewKey: consensusmoduletypes.ModuleName,
				},
			},
			Added: []string{
				consensusmoduletypes.ModuleName,
				ibcfeetypes.ModuleName,
				icacontrollertypes.StoreKey,
				icahosttypes.StoreKey,
				crisistypes.ModuleName,
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if (upgradeInfo.Name == "v1.5.0" || upgradeInfo.Name == "v1.6.1") && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				gravitymoduletypes.ModuleName,
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if upgradeInfo.Name == "v1.8.0" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Deleted: []string{
				"gravity",
			},
			Added: []string{
				gravitymoduletypes.ModuleName,
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if (upgradeInfo.Name == "v1.12.0" || upgradeInfo.Name == "v1.12.1") && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				metrixmoduletypes.ModuleName,
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	app.UpgradeKeeper.SetUpgradeHandler(semverVersion, func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return app.ModuleManager.RunMigrations(ctx, app.configurator, fromVM)
	})
}
