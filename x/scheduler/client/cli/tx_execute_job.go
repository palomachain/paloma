package cli

import (
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/palomachain/paloma/x/scheduler/types"
	vtypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdExecuteJob() *cobra.Command {
	var ff struct {
		payloadPath string
	}

	cmd := &cobra.Command{
		Use:   "execute-job [job-id] [--payload payload.json]",
		Short: "Executes a job",
		Long:  "Given a job and and an optional payload file it will execute the job. Payload format must match the job's payload format. E.g. you can't put payload in the CosmWasm type while the job is for the EVM chain. If job doesn't support payload modification, TX would fail. If the message creator is not allowed to execute the job, the TX would fail.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argJobID := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var payload []byte
			if len(ff.payloadPath) > 0 {
				bz, err := os.ReadFile(ff.payloadPath)
				if err != nil {
					return err
				}
				payload = bz
			}

			creator := clientCtx.GetFromAddress().String()
			msg := &types.MsgExecuteJob{
				JobID:   argJobID,
				Payload: payload,
				Metadata: vtypes.MsgMetadata{
					Creator: creator,
					Signers: []string{creator},
				},
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().StringVar(&ff.payloadPath, "payload", "", "json formatted payload for the job")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
