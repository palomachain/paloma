package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdGetAllQueueNames() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-all-queue-names",
		Short: "Query GetAllQueueNames",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetAllQueueNamesRequest{}

			res, err := queryClient.GetAllQueueNames(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
