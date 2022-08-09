package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdQueryGetSmartContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "smart-contract [smart-contract-id]",
		Short: "Query QueryGetSmartContract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqSmartContractID := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			id, err := strconv.ParseInt(reqSmartContractID, 10, 64)
			if err != nil {
				return err
			}

			params := &types.QueryGetSmartContractRequest{
				SmartContractID: uint64(id),
			}

			res, err := queryClient.QueryGetSmartContract(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
