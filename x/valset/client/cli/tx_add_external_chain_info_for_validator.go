package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdAddExternalChainInfoForValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-external-chain-info-for-validator [chain-id] [address]",
		Short: "Broadcast message AddExternalChainInfoForValidator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChainReferenceID := args[0]
			argAddress := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			creator := clientCtx.GetFromAddress().String()
			msg := &types.MsgAddExternalChainInfoForValidator{
				ChainInfos: []*types.ExternalChainInfo{
					{
						ChainReferenceID: argChainReferenceID,
						Address:          argAddress,
					},
				},
				Metadata: types.MsgMetadata{
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

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
