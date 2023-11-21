package cli

import (
	"fmt"

	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/spf13/cobra"
)

func CmdDeploySmartContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-smart-contract [abi-json] [hex-payload]",
		Short: "Deploy a smart contract to any target chain.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to retrieve ctx: %w", err)
			}

			creator := clientCtx.GetFromAddress()
			title := whoops.Must(cmd.Flags().GetString(cli.FlagTitle))
			description := whoops.Must(cmd.Flags().GetString(cli.FlagDescription))
			abiJSON, bytecodeHex := args[0], args[1]

			msg := types.NewMsgDeployNewSmartContractRequest(creator, title, description, abiJSON, bytecodeHex)

			err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			whoops.Assert(err)
			return nil
		},
	}
	applyFlags(cmd)

	return cmd
}
