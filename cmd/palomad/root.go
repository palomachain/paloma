package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	cosmoslog "cosmossdk.io/log"
	"cosmossdk.io/tools/confix/cmd"
	dbm "github.com/cosmos/cosmos-db"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/viper"

	tmcfg "github.com/cometbft/cometbft/config"
	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/palomachain/paloma/app"
	palomaapp "github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/app/params"
	"github.com/spf13/cobra"
)

// NewRootCmd returns the root command handler for the Paloma daemon.
func NewRootCmd() *cobra.Command {
	// set Bech32 address configuration
	params.SetAddressConfig()

	encCfg := palomaapp.MakeEncodingConfig()
	initClientCtx := client.Context{}.
		WithCodec(encCfg.Codec).
		WithInterfaceRegistry(encCfg.InterfaceRegistry).
		WithTxConfig(encCfg.TxConfig).
		WithLegacyAmino(encCfg.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithHomeDir(palomaapp.DefaultNodeHome).
		WithViper("PALOMA")

	rootCmd := &cobra.Command{
		Use:     palomaapp.Name + "d",
		Short:   "Paloma application network daemon and client",
		PreRunE: rootPreRunE,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customTMConfig := initTendermintConfig()

			if err := server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customTMConfig); err != nil {
				return err
			}

			return applyForcedConfigOptions(cmd)
		},
	}

	ac := appCreator{
		encCfg:        encCfg,
		moduleManager: nil, //TODO
	}
	initRootCmd(rootCmd, ac)

	if err := overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        palomaapp.Name,
		flags.FlagKeyringBackend: keyring.BackendOS,
	}); err != nil {
		// seems to be non-fatal issue
		log.Printf("error overriding flag defaults: %s", err.Error())
	}

	stakingCmd := findCommand(rootCmd, "tx", "staking")

	// all children of tx staking command must check if the pigeon is running
	for _, child := range stakingCmd.Commands() {
		oldPreRun := child.PreRunE
		child.PreRunE = func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				clientCtx, err := client.GetClientTxContext(cmd)
				if err != nil {
					return fmt.Errorf("failed to get client tx: %w", err)
				}

				// Only perform the pigeon liveness check in case the client is running against a local
				// Paloma node.
				if isRunningAgainstLocalNode(clientCtx.NodeURI) {
					// the process will die if pigeon is not running
					app.PigeonMustRun(cmd.Context(), app.PigeonHTTPClient())
				}
			}

			if oldPreRun != nil {
				return oldPreRun(cmd, args)
			}

			return nil
		}
	}

	return rootCmd
}

func isRunningAgainstLocalNode(nodeURI string) bool {
	for _, v := range []string{
		"localhost",
		"127.0.0.1",
		"::1",
	} {
		if strings.Contains(nodeURI, v) {
			return true
		}
	}

	return false
}

// applyForcedConfigOptions reads in the serverContext, applies config to it, and then applies it
func applyForcedConfigOptions(cmd *cobra.Command) error {
	serverCtx := server.GetServerContextFromCmd(cmd)
	serverCtx.Config.Consensus.TimeoutCommit = 1 * time.Second
	return server.SetCmdServerContext(cmd, serverCtx)
}

func initTendermintConfig() *tmcfg.Config {
	cfg := tmcfg.DefaultConfig()

	// these values put a higher strain on node memory
	// cfg.P2P.MaxNumInboundPeers = 100
	// cfg.P2P.MaxNumOutboundPeers = 40

	return cfg
}

func rootPreRunE(cmd *cobra.Command, args []string) error {
	go func() {
		pprofListen := os.Getenv("PPROF_LISTEN")
		if len(pprofListen) > 0 {
			fmt.Println("enabling PPROF")
			mux := http.NewServeMux()

			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

			if err := http.ListenAndServe(pprofListen, mux); err != nil {
				panic(err)
			}
		}
	}()

	// check if pigeon is running
	go func() {
		app.CheckPigeonRunningLooper(cmd.Context(), app.PigeonHTTPClient())
	}()

	return nil
}

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

func initRootCmd(rootCmd *cobra.Command, ac appCreator) {
	rootCmd.AddCommand(
		genutilcli.InitCmd(ac.moduleManager, palomaapp.DefaultNodeHome),
		genesisCommand(ac.encCfg),
		tmcli.NewCompletionCmd(rootCmd, true),
		debug.Cmd(),
		cmd.ConfigCommand(),
	)

	server.AddCommands(rootCmd, palomaapp.DefaultNodeHome, newApp, appExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		queryCommand(ac.moduleManager),
		txCommand(ac.moduleManager),
		keys.Commands(),
	)

	rootCmd.AddCommand(
		snapshot.Cmd(newApp),
	)
}

// genesisCommand builds genesis-related `palomad genesis` command. Users may provide application specific commands as a parameter
func genesisCommand(encodingConfig params.EncodingConfig, cmds ...*cobra.Command) *cobra.Command {
	cmd := genutilcli.GenesisCoreCommand(encodingConfig.TxConfig, palomaapp.ModuleBasics, palomaapp.DefaultNodeHome)

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
