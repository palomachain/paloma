package main

import (
	"errors"
	"io"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	palomaapp "github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/app/params"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type appCreator struct {
	encCfg        params.EncodingConfig
	moduleManager module.BasicManager
}

func (ac appCreator) newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	return palomaapp.New(
		logger,
		db,
		traceStore,
		true,
		ac.encCfg,
		appOpts,
		server.DefaultBaseappOptions(appOpts)...,
	)
}

func (ac appCreator) appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var app *palomaapp.App

	// This check is necessary as we use the flag in x/upgrade, but we can exit
	// more gracefully by checking the flag here.
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}

	// overwrite the FlagInvCheckPeriod
	viperAppOpts.Set(server.FlagInvCheckPeriod, 1)
	appOpts = viperAppOpts

	if height != -1 {
		app = palomaapp.New(logger, db, traceStore, false, ac.encCfg, appOpts)

		if err := app.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		app = palomaapp.New(logger, db, traceStore, true, ac.encCfg, appOpts)
	}

	return app.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// GetAppDBBackend gets the backend type to use for the application DBs.
func GetAppDBBackend(opts servertypes.AppOptions) dbm.BackendType {
	rv := cast.ToString(opts.Get("app-db-backend"))
	if len(rv) == 0 {
		rv = cast.ToString(opts.Get("db-backend"))
	}
	if len(rv) != 0 {
		return dbm.BackendType(rv)
	}

	return dbm.GoLevelDBBackend
}
