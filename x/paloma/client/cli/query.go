package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/palomachain/paloma/v2/x/paloma/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group paloma queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdQueryLightNodeClientFeegranter())
	cmd.AddCommand(CmdQueryLightNodeClientLicenses())
	cmd.AddCommand(CmdQueryLightNodeClientFunders())
	cmd.AddCommand(CmdQueryLightNodeClients())

	return cmd
}
