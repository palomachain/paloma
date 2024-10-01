package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/v2/x/paloma/types"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

func CmdQueryLightNodeClientFunders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "light-node-client-funders",
		Short: "Shows the current light node client funders settings, set by governance vote",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.GetLightNodeClientFunders(
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
