package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/spf13/cobra"
)

func CmdQueryUserSmartContracts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-smart-contracts [validator-address]",
		Short: "Query user-uploaded smart contracts",
		Long:  `List all user-upload smart contracts associated with "validator-address". Together with the smart contract fields, each contract has a list of deployments containing status and address, if applicable.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to retrieve ctx: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryUserSmartContractsRequest{
				ValAddress: args[0],
			}

			res, err := queryClient.GetUserSmartContracts(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
