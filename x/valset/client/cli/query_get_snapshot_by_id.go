package cli

import (
	"strconv"

	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdGetSnapshotByID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-snapshot-by-id [snapshot-id]",
		Short: "Query GetSnapshotByID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqSnapshotId := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			id := uint64(whoops.Must(strconv.Atoi(reqSnapshotId)))
			params := &types.QueryGetSnapshotByIDRequest{
				SnapshotId: id,
			}

			res, err := queryClient.GetSnapshotByID(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
