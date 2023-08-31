package cli

import (
	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/spf13/cobra"
)

func CmdGravityProposalHandler() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gravity",
		Short: "Gravity proposals",
	}
	cmd.AddCommand([]*cobra.Command{
		CmdSetErc20ToDenom(),
	}...)

	return cmd
}

func applyFlags(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
}

func getDeposit(cmd *cobra.Command) (sdk.Coins, error) {
	depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
	whoops.Assert(err)
	return sdk.ParseCoinsNormalized(depositStr)
}

func CmdSetErc20ToDenom() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "set-erc20-to-denom [chain-reference-id] [denom] [erc20]",
		Short: "Sets an association between a denom and an erc20 token for a given chain",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return whoops.Try(func() {
				cliCtx, err := client.GetClientTxContext(cmd)
				whoops.Assert(err)

				chainReferenceID := args[0]
				denom := args[1]
				erc20 := args[2]

				setERC20ToDenomProposal := &types.SetERC20ToDenomProposal{
					Title:            whoops.Must(cmd.Flags().GetString(cli.FlagTitle)),
					Description:      whoops.Must(cmd.Flags().GetString(cli.FlagDescription)),
					ChainReferenceId: chainReferenceID,
					Erc20:            erc20,
					Denom:            denom,
				}

				from := cliCtx.GetFromAddress()

				deposit, err := getDeposit(cmd)
				whoops.Assert(err)

				msg, err := govv1beta1types.NewMsgSubmitProposal(setERC20ToDenomProposal, deposit, from)
				whoops.Assert(err)

				err = tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
				whoops.Assert(err)
			})
		},
	}
	applyFlags(cmd)
	return cmd
}
