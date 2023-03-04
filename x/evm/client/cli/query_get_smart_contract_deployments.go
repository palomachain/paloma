package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdQueryGetSmartContractDeployments() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "smart-contract-deployments",
		Short: "Query QueryGetSmartContractDeployments",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetSmartContractDeploymentsRequest{}

			res, err := queryClient.QueryGetSmartContractDeployments(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
