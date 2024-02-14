package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	cosmoslog "cosmossdk.io/log"
	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/palomachain/paloma/app"
	palomaapp "github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/app/params"
	"github.com/spf13/cobra"
)

// NewRootCmd returns the root command handler for the Paloma daemon.
func NewRootCmd() *cobra.Command {
	// set Bech32 address configuration
	params.SetAddressConfig()

	tempApp := palomaapp.New(cosmoslog.NewNopLogger(), db.NewMemDB(), io.MultiWriter(), true, db.OptionsMap{})
	encCfg := params.EncodingConfig{
		InterfaceRegistry: tempApp.InterfaceRegistry(),
		Codec:             tempApp.AppCodec(),
		TxConfig:          tempApp.TxConfig(),
		Amino:             tempApp.LegacyAmino(),
	}

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
			customTMConfig := initCometBFTConfig()

			if err := server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customTMConfig); err != nil {
				return err
			}

			return applyForcedConfigOptions(cmd)
		},
	}
	initRootCmd(rootCmd, encCfg, encCfg.InterfaceRegistry, encCfg, tempApp.BasicModuleManager)

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

	// add keyring to autocli opts
	autoCliOpts := tempApp.AutoCliOpts()
	initClientCtx, _ = config.ReadFromClientConfig(initClientCtx)
	autoCliOpts.Keyring, _ = keyring.NewAutoCLIKeyring(initClientCtx.Keyring)
	autoCliOpts.ClientCtx = initClientCtx
	autoCliOpts.AddressCodec = address.Bech32Codec{
		Bech32Prefix: params.AccountAddressPrefix,
	}
	autoCliOpts.ValidatorAddressCodec = address.Bech32Codec{
		Bech32Prefix: params.ValidatorAddressPrefix,
	}
	autoCliOpts.ConsensusAddressCodec = address.Bech32Codec{
		Bech32Prefix: params.ConsNodeAddressPrefix,
	}
	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
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
