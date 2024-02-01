package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/palomachain/paloma/x/metrix/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	// Group paloma queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdQueryValidator())
	cmd.AddCommand(CmdQueryValidators())
	cmd.AddCommand(CmdQueryHistoricRelayData())

	return cmd
}
