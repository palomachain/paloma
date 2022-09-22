package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/scheduler/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdGetJobByID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Finds job by provoding job's ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			jobID := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetJobByIDRequest{
				JobID: jobID,
			}

			res, err := queryClient.QueryGetJobByID(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
