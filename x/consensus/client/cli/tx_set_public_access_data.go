package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdSetPublicAccessData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-public-access-data [message-id] [queue-type-name] [data]",
		Short: "Broadcast message SetPublicAccessData",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			panic("we dont need this")
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
