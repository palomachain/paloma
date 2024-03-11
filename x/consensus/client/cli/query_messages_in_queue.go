package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdMessagesInQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages-in-queue [queue-type-name]",
		Short: "Query MessagesInQueue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqQueueTypeName := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryMessagesInQueueRequest{
				QueueTypeName: reqQueueTypeName,
			}

			res, err := queryClient.MessagesInQueue(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdMessageByID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "message-by-id [queue-type-name] [message-id]",
		Short: "Query a message by queue name and ID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqQueueTypeName := args[0]
			reqID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryMessageByIDRequest{
				QueueTypeName: reqQueueTypeName,
				Id:            reqID,
			}

			res, err := queryClient.MessageByID(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
