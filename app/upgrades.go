package app

import (
	"context"

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
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	consensusmoduletypes "github.com/palomachain/paloma/x/consensus/types"
	metrixmoduletypes "github.com/palomachain/paloma/x/metrix/types"
	skywaymoduletypes "github.com/palomachain/paloma/x/skyway/types"
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
	_ = app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

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
				"gravity",
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
				skywaymoduletypes.ModuleName,
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
