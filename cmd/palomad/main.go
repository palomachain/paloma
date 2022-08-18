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
				if oldPreRunE != nil {
					return oldPreRunE(cmd, args)
				}
				return nil
			}
		}),
		// this line is used by starport scaffolding # root/arguments
	)
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
