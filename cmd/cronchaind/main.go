package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/spf13/cobra"
	"github.com/tendermint/starport/starport/pkg/cosmoscmd"
	"github.com/volumefi/cronchain/app"
)

func main() {
	rootCmd, _ := cosmoscmd.NewRootCmd(
		app.Name,
		app.AccountAddressPrefix,
		app.DefaultNodeHome,
		app.Name,
		app.ModuleBasics,
		app.New,
		cosmoscmd.CustomizeStartCmd(func(startCmd *cobra.Command) {
			oldRunE := startCmd.RunE
			customRunE := func(cmd *cobra.Command, args []string) error {

				// start runners
				fmt.Println("STARTING RUNNERS")

				return oldRunE(cmd, args)
			}

			startCmd.RunE = customRunE
		}),
		// this line is used by starport scaffolding # root/arguments
	)
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
