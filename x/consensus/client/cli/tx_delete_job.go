package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdDeleteJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-job [queue-type-name] [message-id]",
		Short: "Broadcast message DeleteJob",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argQueueTypeName := args[0]
			argMessageID := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msgId, err := strconv.Atoi(argMessageID)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteJob(
				clientCtx.GetFromAddress().String(),
				argQueueTypeName,
				uint64(msgId),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
