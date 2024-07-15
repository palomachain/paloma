package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/x/paloma/types"
	"github.com/spf13/cobra"
)

func CmdPalomaChainProposalHandler() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paloma",
		Short: "Paloma proposals",
	}

	cmd.AddCommand(CmdPalomaProposeLightNodeClientFeegranter())

	return cmd
}

func applyFlags(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagSummary, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")

	//nolint:errcheck
	cmd.MarkFlagRequired(cli.FlagTitle)
	//nolint:errcheck
	cmd.MarkFlagRequired(cli.FlagSummary)
}

func CmdPalomaProposeLightNodeClientFeegranter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-light-node-client-feegranter [account]",
		Short: "Proposal to set new feegranter account for light node clients",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			// Sanity-check account string
			if _, err := sdk.AccAddressFromBech32(args[0]); err != nil {
				return err
			}

			prop := &types.SetLightNodeClientFeegranterProposal{
				FeegranterAccount: args[0],
				Title:             title,
				Description:       description,
			}

			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			msg, err := govv1beta1types.NewMsgSubmitProposal(prop, deposit, from)
			if err != nil {
				return err
			}

			err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			if err != nil {
				return err
			}

			return nil
		},
	}

	applyFlags(cmd)

	return cmd
}
