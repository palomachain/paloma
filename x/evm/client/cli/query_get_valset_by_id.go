package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdGetValsetByID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-valset-by-id [valset-id] [chain-id]",
		Short: "Query GetValsetByID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqValsetID, err := cast.ToUint64E(args[0])
			chainReferenceID := args[1]
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetValsetByIDRequest{
				ValsetID:         reqValsetID,
				ChainReferenceID: chainReferenceID,
			}

			res, err := queryClient.GetValsetByID(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
