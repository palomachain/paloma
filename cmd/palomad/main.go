package main

import (
	"fmt"
	"os"

	"net/http"
	"net/http/pprof"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/palomachain/paloma/app"
	"github.com/spf13/cobra"
	"github.com/tendermint/starport/starport/pkg/cosmoscmd"
)

func main() {
	rootCmd, _ := cosmoscmd.NewRootCmd(
		app.Name,
		app.AccountAddressPrefix,
		app.DefaultNodeHome,
		app.Name,
		app.ModuleBasics,
		app.New,
		cosmoscmd.CustomizeStartCmd(func(cmd *cobra.Command) {
			oldPreRunE := cmd.PreRunE
			cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
				go func() {
					listen := os.Getenv("PPROF_LISTEN")
					if len(listen) > 0 {
						fmt.Println("enabling PPROF")
						mux := http.NewServeMux()

						mux.HandleFunc("/debug/pprof/", pprof.Index)
						mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
						mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
						mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
						mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

						if err := http.ListenAndServe(listen, mux); err != nil {
							panic(err)
						}
					}
				}()

				// check if pigeon is running
				go func() {
					app.CheckPigeonRunningLooper(cmd.Context(), app.PigeonHTTPClient())
				}()

				if oldPreRunE != nil {
					return oldPreRunE(cmd, args)
				}
				return nil
			}
		}),
	)

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

	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}

func findCommand(root *cobra.Command, path ...string) *cobra.Command {
	cmd, _, err := root.Traverse(path)
	if err != nil {
		panic(err)
	}
	return cmd
}
