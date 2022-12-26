package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/palomachain/paloma/app"
	palomaapp "github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/app/params"
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
		WithBroadcastMode(flags.BroadcastBlock).
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
			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig)
		},
	}

	ac := appCreator{
		encCfg:        encCfg,
		moduleManager: palomaapp.ModuleBasics,
	}
	initRootCmd(rootCmd, ac)

	overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        palomaapp.Name,
		flags.FlagKeyringBackend: keyring.BackendOS,
	})

	stakingCmd := findCommand(rootCmd, "tx", "staking")

	// all children of tx staking command must check if the pigeon is running
	for _, child := range stakingCmd.Commands() {
		oldPreRun := child.PreRunE
		child.PreRunE = func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// the process will die if pigeon is not running
				app.PigeonMustRun(cmd.Context(), app.PigeonHTTPClient())
			}

			if oldPreRun != nil {
				return oldPreRun(cmd, args)
			}

			return nil
		}
	}

	return rootCmd
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

func initRootCmd(rootCmd *cobra.Command, ac appCreator) {
	rootCmd.AddCommand(
		genutilcli.InitCmd(ac.moduleManager, palomaapp.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, palomaapp.DefaultNodeHome),
		genutilcli.MigrateGenesisCmd(),
		genutilcli.GenTxCmd(
			ac.moduleManager,
			ac.encCfg.TxConfig,
			banktypes.GenesisBalancesIterator{},
			palomaapp.DefaultNodeHome,
		),
		genutilcli.ValidateGenesisCmd(ac.moduleManager),
		AddGenesisAccountCmd(palomaapp.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		debug.Cmd(),
		config.Cmd(),
	)

	server.AddCommands(rootCmd, palomaapp.DefaultNodeHome, ac.newApp, ac.appExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(ac.moduleManager),
		txCommand(ac.moduleManager),
		keys.Commands(palomaapp.DefaultNodeHome),
	)
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
		authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
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
