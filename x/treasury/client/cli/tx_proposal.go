package cli

import (
	"fmt"
	"regexp"

	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/x/treasury/types"
	"github.com/spf13/cobra"
)

func applyFlags(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
}

func CmdTreasuryProposalHandler() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "treasury",
		Short: "Treasury proposals",
	}

	cmd.AddCommand(CmdTreasuryProposeSecurityFee())
	cmd.AddCommand(CmdTreasuryProposeCommunityFundFee())

	return cmd
}

func checkFeeValid(fee string) bool {
	if fee == "" {
		return false
	}

	re := regexp.MustCompile(`0\.[0-9]{1,8}`)
	found := re.FindString(fee)
	return found == fee
}

func getDeposit(cmd *cobra.Command) (sdk.Coins, error) {
	depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
	whoops.Assert(err)
	return sdk.ParseCoinsNormalized(depositStr)
}

func CmdTreasuryProposeSecurityFee() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-security-fee [fee]",
		Short: "Proposal to change the security fee",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return whoops.Try(func() {
				clientCtx, err := client.GetClientTxContext(cmd)
				whoops.Assert(err)

				fee := args[0]

				if !checkFeeValid(fee) {
					whoops.Assert(fmt.Errorf("invalid fee: %s", fee))
				}

				feeProposal := &types.SecurityFeeProposal{
					Title:       whoops.Must(cmd.Flags().GetString(cli.FlagTitle)),
					Description: whoops.Must(cmd.Flags().GetString(cli.FlagDescription)),
					Fee:         fee,
				}

				from := clientCtx.GetFromAddress()

				deposit, err := getDeposit(cmd)
				whoops.Assert(err)

				msg, err := govv1beta1types.NewMsgSubmitProposal(feeProposal, deposit, from)
				whoops.Assert(err)

				err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
				whoops.Assert(err)
			})
		},
	}
	applyFlags(cmd)

	return cmd
}

func CmdTreasuryProposeCommunityFundFee() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-community-fund-fee [fee]",
		Short: "Proposal to change the community fund fee",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return whoops.Try(func() {
				clientCtx, err := client.GetClientTxContext(cmd)
				whoops.Assert(err)

				fee := args[0]

				if !checkFeeValid(fee) {
					whoops.Assert(fmt.Errorf("invalid fee: %s", fee))
				}

				feeProposal := &types.CommunityFundFeeProposal{
					Title:       whoops.Must(cmd.Flags().GetString(cli.FlagTitle)),
					Description: whoops.Must(cmd.Flags().GetString(cli.FlagDescription)),
					Fee:         fee,
				}

				from := clientCtx.GetFromAddress()

				deposit, err := getDeposit(cmd)
				whoops.Assert(err)

				msg, err := govv1beta1types.NewMsgSubmitProposal(feeProposal, deposit, from)
				whoops.Assert(err)

				err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
				whoops.Assert(err)
			})
		},
	}
	applyFlags(cmd)

	return cmd
}
