package main

import (
	"errors"
	"io"

	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/cosmos/cosmos-sdk/client/debug"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	palomaapp "github.com/palomachain/paloma/app"

	cosmoslog "cosmossdk.io/log"
	"cosmossdk.io/tools/confix/cmd"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/app/params"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// appExport creates a new simapp (optionally at a given height) and exports state.
func appExport(
	logger cosmoslog.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	// this check is necessary as we use the flag in x/upgrade.
	// we can exit more gracefully by checking the flag here.
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

	var App *app.App
	if height != -1 {
		App = app.New(logger, db, traceStore, false, params.EncodingConfig{}, nil)

		if err := App.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		App = app.New(logger, db, traceStore, true, params.EncodingConfig{}, nil)
	}

	return App.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// newApp creates the application
func newApp(
	logger cosmoslog.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	//baseappOptions := server.DefaultBaseappOptions(appOpts)

	return app.New(
		logger, db, traceStore, true,
		params.EncodingConfig{}, nil,
	)
}

func initRootCmd(
	rootCmd *cobra.Command,
	txConfig params.EncodingConfig,
	interfaceRegistry codectypes.InterfaceRegistry,
	appCodec params.EncodingConfig,
	basicManager module.BasicManager,
) {

	rootCmd.AddCommand(
		genutilcli.InitCmd(basicManager, palomaapp.DefaultNodeHome),
		genesisCommand(txConfig, basicManager, appExport),
		tmcli.NewCompletionCmd(rootCmd, true),
		debug.Cmd(),
		cmd.ConfigCommand(),
	)

	server.AddCommands(rootCmd, palomaapp.DefaultNodeHome, newApp, appExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		queryCommand(basicManager),
		txCommand(basicManager),
		keys.Commands(),
	)
	rootCmd.AddCommand(
		snapshot.Cmd(newApp),
	)
}

// genesisCommand builds genesis-related `palomad genesis` command. Users may provide application specific commands as a parameter
func genesisCommand(encodingConfig params.EncodingConfig, basicManager module.BasicManager, appExport servertypes.AppExporter, cmds ...*cobra.Command) *cobra.Command {
	cmd := genutilcli.GenesisCoreCommand(encodingConfig.TxConfig, basicManager, palomaapp.DefaultNodeHome)

	for _, sub_cmd := range cmds {
		cmd.AddCommand(sub_cmd)
	}
	return cmd
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

func queryCommand(mm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.QueryEventForTxCmd(),
		rpc.ValidatorCommand(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	mm.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func txCommand(mm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions sub-commands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
	)

	mm.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}
