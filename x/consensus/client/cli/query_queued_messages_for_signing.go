package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/volumefi/cronchain/x/consensus/types"
)

var _ = strconv.Itoa(0)

func CmdQueuedMessagesForSigning() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queued-messages-for-signing [valaddress] [queueTypeName]",
		Short: "Query QueuedMessagesForSigning",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryQueuedMessagesForSigningRequest{
				ValAddress:    args[0],
				QueueTypeName: args[1],
			}

			res, err := queryClient.QueuedMessagesForSigning(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
