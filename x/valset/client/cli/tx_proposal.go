package cli

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/v2/x/valset/types"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

const flagTargetBlockHeight = "target-block-height"

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

func getCurrentVersion(cmd *cobra.Command, clientCtx client.Context) (string, error) {
	queryClient := types.NewQueryClient(clientCtx)
	queryParams := &types.QueryGetPigeonRequirementsRequest{}
	res, err := queryClient.GetPigeonRequirements(cmd.Context(), queryParams)
	if err != nil {
		return "", err
	}

	return res.PigeonRequirements.MinVersion, nil
}

func CmdValsetChainProposalHandler() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "valset",
		Short: "Valset proposals",
	}

	cmd.AddCommand(CmdValsetProposePigeonRequirements())

	return cmd
}

func CmdValsetProposePigeonRequirements() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-pigeon-requirements [minimum-version]",
		Short: "Proposal to set new pigeon requirements",
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

			blockHeight, err := cmd.Flags().GetUint64(flagTargetBlockHeight)
			if err != nil {
				return err
			}

			curVersion, err := getCurrentVersion(cmd, clientCtx)
			if err != nil {
				return err
			}

			minVersion := args[0]

			if semver.Compare(minVersion, curVersion) < 0 {
				return errors.New("new version cannot be lower than current")
			}

			prop := &types.SetPigeonRequirementsProposal{
				MinVersion:        args[0],
				TargetBlockHeight: blockHeight,
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

	cmd.Flags().Uint64(flagTargetBlockHeight, 0, "block height when requirements will be applied")

	applyFlags(cmd)

	return cmd
}
