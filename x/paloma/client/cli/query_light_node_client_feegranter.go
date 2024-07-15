package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/paloma/types"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

func CmdQueryLightNodeClientFeegranter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "light-node-client-feegranter",
		Short: "Shows the current light node client fee granter account, set by governance vote",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := &emptypb.Empty{}
			res, err := queryClient.GetLightNodeClientFeegranter(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
