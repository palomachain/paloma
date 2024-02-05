package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/metrix/types"
	"github.com/spf13/cobra"
)

func CmdQueryHistoricRelayData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "historic-relay-data [validator-address]",
		Short: "Gets the historic relay performance data used for the given validator address.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			valAddr := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryHistoricRelayDataRequest{
				ValAddress: valAddr,
			}

			res, err := queryClient.HistoricRelayData(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
