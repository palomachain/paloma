package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

type submitNewJobPayloadJson struct {
}

func CmdSubmitNewJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-new-job [chain-id] [smart contract address] [smart contract payload] [method] [smart contract JSON abi]",
		Short: "Broadcast message SubmitNewJob",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgSubmitNewJob{
				Creator:                 clientCtx.GetFromAddress().String(),
				ChainID:                 args[0],
				HexSmartContractAddress: args[1],
				HexPayload:              args[2],
				Method:                  args[3],
				Abi:                     args[4],
				ChainType:               "EVM",
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
