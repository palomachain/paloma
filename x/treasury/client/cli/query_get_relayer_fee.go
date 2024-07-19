package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/treasury/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdGetRelayerFee() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "relayer-fee [validator-address]",
		Short:   "Retrieve realyer fee values",
		Long:    "Query the current relayer fee settings for a given validator",
		Example: "relayer-fee palomavaloper...",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			validator, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.RelayerFee(cmd.Context(), &types.QueryRelayerFeeRequest{
				ValAddress: validator.String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetRelayerFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "relayer-fees [chain-reference-id]",
		Short:   "Retrieve all set realyer fees values for a given chain",
		Long:    "Query all currently set relayer fee settings for a given chain",
		Example: "relayer-fees eth-main",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.RelayerFees(cmd.Context(), &types.QueryRelayerFeesRequest{
				ChainReferenceId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
