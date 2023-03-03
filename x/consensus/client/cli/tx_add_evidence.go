package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdAddEvidence() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-evidence [message-id] [signature] [public-key] [evidence]",
		Short: "Broadcast message AddEvidence",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			panic("remove! we don't need this")
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
