package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/paloma/types"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

func CmdQueryLightNodeClientFunder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "light-node-client-funder",
		Short: "Shows the current light node client funder settings, set by governance vote",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.GetLightNodeClientFunder(
				context.Background(), &emptypb.Empty{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
