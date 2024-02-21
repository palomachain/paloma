package cli

import (
	"fmt"
	"strconv"

	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/palomachain/paloma/x/evm/types"
	vtypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cobra"
)

func CmdRemoveSmartContractDeployment() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-smart-contract-deployment smart-contract-id chain-reference-id",
		Short: "Used to remove a stuck smart contract deployment.  The smart contract will attempt to redeploy automatically",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to retrieve ctx: %w", err)
			}

			smartContractID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("unable to parse smart-contract-id %s as uint64", args[0])
			}
			chainReferenceID := args[1]

			creator := clientCtx.GetFromAddress().String()
			msg := &types.MsgRemoveSmartContractDeploymentRequest{
				SmartContractID:  uint64(smartContractID),
				ChainReferenceID: chainReferenceID,
				Metadata: vtypes.MsgMetadata{
					Creator: creator,
					Signers: []string{creator},
				},
			}

			err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			whoops.Assert(err)
			return nil
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
