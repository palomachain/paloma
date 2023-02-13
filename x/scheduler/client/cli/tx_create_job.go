package cli

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/palomachain/paloma/x/scheduler/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdCreateJob() *cobra.Command {
	var ff struct {
		id      string
		chType  string
		chRefId string

		def     string
		payload string

		payloadModifiable bool
	}

	cmd := &cobra.Command{
		Use:   "create-job [--job-id] [--chain-type] [--chain-ref-id] [--definition] [--payload] [--payload-modifiable]",
		Short: "Creates a new job",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			job := &types.Job{
				ID: ff.id,
				Routing: types.Routing{
					ChainType:        ff.chType,
					ChainReferenceID: ff.chRefId,
				},
				IsPayloadModifiable: ff.payloadModifiable,
			}

			job.Definition, err = ioutil.ReadFile(ff.def)
			if err != nil {
				return err
			}

			job.Payload, err = ioutil.ReadFile(ff.payload)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateJob{
				Creator: clientCtx.GetFromAddress().String(),
				Job:     job,
			}
			fmt.Printf("[msgcreatejob] UNPACK ARGS: %+v\n", msg)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringVar(&ff.id, "job-id", "", "job's id")
	cmd.Flags().StringVar(&ff.chType, "chain-type", "", "chain's type (e.g. evm)")
	cmd.Flags().StringVar(&ff.chRefId, "chain-ref-id", "", "chain's reference (e.g. eth-main, binance,...)")
	cmd.Flags().StringVar(&ff.def, "definition", "", "path to a job's definition json file")
	cmd.Flags().StringVar(&ff.payload, "payload", "", "path to a job's default json file. Can be empty, but then payload-modifiable flag must be set")
	cmd.Flags().BoolVar(&ff.payloadModifiable, "payload-modifiable", false, "")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
