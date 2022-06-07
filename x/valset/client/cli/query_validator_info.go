package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdValidatorInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-info [val-addr]",
		Short: "Query ValidatorInfo",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqValAddr := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryValidatorInfoRequest{
				ValAddr: reqValAddr,
			}

			res, err := queryClient.ValidatorInfo(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
