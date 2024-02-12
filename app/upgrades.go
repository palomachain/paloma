package app

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	ibctmmigrations "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint/migrations"
	consensusmoduletypes "github.com/palomachain/paloma/x/consensus/types"
	evmmoduletypes "github.com/palomachain/paloma/x/evm/types"
	gravitymoduletypes "github.com/palomachain/paloma/x/gravity/types"
	metrixmoduletypes "github.com/palomachain/paloma/x/metrix/types"
	palomamoduletypes "github.com/palomachain/paloma/x/paloma/types"
	schedulermoduletypes "github.com/palomachain/paloma/x/scheduler/types"
	treasurymoduletypes "github.com/palomachain/paloma/x/treasury/types"
	valsetmoduletypes "github.com/palomachain/paloma/x/valset/types"
)

var minCommissionRate = sdk.MustNewDecFromStr("0.05")

// UpdateMinCommissionRate update minimum commission rate param.
func UpdateMinCommissionRate(ctx sdk.Context, keeper stakingkeeper.Keeper) (sdk.Dec, error) {
	params := keeper.GetParams(ctx)
	params.MinCommissionRate = minCommissionRate

	if err := keeper.SetParams(ctx, params); err != nil {
		return sdk.Dec{}, err
	}

	return minCommissionRate, nil
}

// SetMinimumCommissionRate updates the commission rate for validators
// whose current commission rate is lower than the new minimum commission rate.
func SetMinimumCommissionRate(ctx sdk.Context, keeper stakingkeeper.Keeper, minCommissionRate sdk.Dec) error {
	validators := keeper.GetAllValidators(ctx)

	for _, validator := range validators {
		if validator.Commission.Rate.IsNil() || validator.Commission.Rate.LT(minCommissionRate) {
			if err := keeper.Hooks().BeforeValidatorModified(ctx, validator.GetOperator()); err != nil {
				return err
			}

			validator.Commission.Rate = minCommissionRate
			validator.Commission.UpdateTime = ctx.BlockTime()

			keeper.SetValidator(ctx, validator)
		}
	}

	return nil
}

func (app *App) RegisterUpgradeHandlers(semverVersion string) {
	// Set param key table for params module migration
	for _, subspace := range app.ParamsKeeper.GetSubspaces() {
		subspace := subspace

		var keyTable paramstypes.KeyTable
		switch subspace.Name() {
		case authtypes.ModuleName:
			keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
		case banktypes.ModuleName:
			keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
		case stakingtypes.ModuleName:
			keyTable = stakingtypes.ParamKeyTable() //nolint:staticcheck
		case minttypes.ModuleName:
			keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
		case distrtypes.ModuleName:
			keyTable = distrtypes.ParamKeyTable() //nolint:staticcheck
		case slashingtypes.ModuleName:
			keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
		case govtypes.ModuleName:
			keyTable = govv1.ParamKeyTable() //nolint:staticcheck
		case crisistypes.ModuleName:
			keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck
		case consensusmoduletypes.ModuleName:
			keyTable = consensusmoduletypes.ParamKeyTable() //nolint:staticcheck
		case evmmoduletypes.ModuleName:
			keyTable = evmmoduletypes.ParamKeyTable() //nolint:staticcheck
		case gravitymoduletypes.ModuleName:
			keyTable = gravitymoduletypes.ParamKeyTable() //nolint:staticcheck
		case palomamoduletypes.ModuleName:
			keyTable = palomamoduletypes.ParamKeyTable() //nolint:staticcheck
		case schedulermoduletypes.ModuleName:
			keyTable = schedulermoduletypes.ParamKeyTable() //nolint:staticcheck
		case treasurymoduletypes.ModuleName:
			keyTable = treasurymoduletypes.ParamKeyTable() //nolint:staticcheck
		case valsetmoduletypes.ModuleName:
			keyTable = valsetmoduletypes.ParamKeyTable() //nolint:staticcheck
		case metrixmoduletypes.ModuleName:
			keyTable = metrixmoduletypes.ParamKeyTable() //nolint:staticcheck
		case wasmtypes.ModuleName:
			keyTable = wasmtypes.ParamKeyTable() //nolint:staticcheck
		}

		if !subspace.HasKeyTable() {
			subspace.WithKeyTable(keyTable)
		}
	}

	baseAppLegacySS := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

	app.UpgradeKeeper.SetUpgradeHandler(
		semverVersion,
		func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			// Migrate CometBFT consensus parameters from x/params module to a
			// dedicated x/consensus module.
			baseapp.MigrateParams(ctx, baseAppLegacySS, &app.ConsensusParamsKeeper)

			// TODO: We may need to execute ibc-go v6 migrations but importing ibc-go
			// v6 will fail using Cosmos SDK v0.47.x.
			//
			// if err := v6.MigrateICS27ChannelCapability(ctx, app.cdc, app.keys[capabilitytypes.StoreKey], app.CapabilityKeeper, ""); err != nil {
			// 	return nil, err
			// }

			// OPTIONAL: prune expired tendermint consensus states to save storage space
			if _, err := ibctmmigrations.PruneExpiredConsensusStates(ctx, app.appCodec, app.IBCKeeper.ClientKeeper); err != nil {
				return nil, err
			}

			// save oldIcaVersion, so we can skip icahost.InitModule in longer term tests.
			oldIcaVersion := fromVM[icatypes.ModuleName]

			// Add Interchain Accounts host module
			// set the ICS27 consensus version so InitGenesis is not run
			fromVM[icatypes.ModuleName] = app.mm.Modules[icatypes.ModuleName].(module.HasConsensusVersion).ConsensusVersion()

			// create ICS27 Host submodule params, host module not enabled.
			hostParams := icahosttypes.Params{
				HostEnabled:   false,
				AllowMessages: []string{},
			}

			mod, found := app.mm.Modules[icatypes.ModuleName]
			if !found {
				return nil, fmt.Errorf("module %s is not in the module manager", icatypes.ModuleName)
			}

			icaMod, ok := mod.(ica.AppModule)
			if !ok {
				return nil, fmt.Errorf("expected module %s to be type %T, got %T", icatypes.ModuleName, ica.AppModule{}, mod)
			}

			// skip InitModule in upgrade tests after the upgrade has gone through.
			if oldIcaVersion != fromVM[icatypes.ModuleName] {
				icaMod.InitModule(ctx, icacontrollertypes.DefaultParams(), hostParams)
			}

			vm, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
			if err != nil {
				return vm, err
			}

			minCommissionRate, err := UpdateMinCommissionRate(ctx, *app.StakingKeeper)
			if err != nil {
				return vm, err
			}

			err = SetMinimumCommissionRate(ctx, *app.StakingKeeper, minCommissionRate)
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
}
