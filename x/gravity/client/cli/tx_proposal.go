package cli

import (
	"fmt"
	"strconv"

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

const (
	flagExcludedTokens  = "excluded-tokens"
	flagExemptAddresses = "exempt-addresses"
)

func CmdGravityProposalHandler() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gravity",
		Short: "Gravity proposals",
	}
	cmd.AddCommand([]*cobra.Command{
		CmdSetErc20ToDenom(),
		CmdSetBridgeTax(),
	}...)

	return cmd
}

func applyFlags(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagSummary, "", "description of proposal")
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
					Description:      whoops.Must(cmd.Flags().GetString(cli.FlagSummary)),
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

func CmdSetBridgeTax() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-bridge-tax [tax-value]",
		Short: "Sets the bridge tax value, and optionally token exceptions and exempt addresses",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			rateRaw := args[0]

			rate, err := strconv.ParseFloat(rateRaw, 32)
			if err != nil {
				return err
			}

			if rate < 0 || rate > 1 {
				return fmt.Errorf("invalid tax rate: %v", rate)
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			excludedTokens, err := cmd.Flags().GetStringSlice(flagExcludedTokens)
			if err != nil {
				return err
			}

			exemptAddresses, err := cmd.Flags().GetStringSlice(flagExemptAddresses)
			if err != nil {
				return err
			}

			prop := &types.SetBridgeTaxProposal{
				Title:           title,
				Description:     description,
				Rate:            rateRaw,
				ExcludedTokens:  excludedTokens,
				ExemptAddresses: exemptAddresses,
			}

			from := cliCtx.GetFromAddress()

			deposit, err := getDeposit(cmd)
			if err != nil {
				return err
			}

			msg, err := govv1beta1types.NewMsgSubmitProposal(prop, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringSlice(flagExcludedTokens, []string{},
		"list of tokens excluded from the bridge tax")
	cmd.Flags().StringSlice(flagExemptAddresses, []string{},
		"list of addresses exempt from the bridge tax")

	applyFlags(cmd)
	return cmd
}
