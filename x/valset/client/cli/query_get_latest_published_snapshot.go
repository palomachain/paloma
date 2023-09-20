package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cobra"
)

func CmdGetLatestPublishedSnapshot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-latest-published-snapshot [chain-reference-id]",
		Short: "Query GetLatestPublishedSnapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetLatestPublishedSnapshotRequest{
				ChainReferenceID: args[0],
			}

			res, err := queryClient.GetLatestPublishedSnapshot(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
