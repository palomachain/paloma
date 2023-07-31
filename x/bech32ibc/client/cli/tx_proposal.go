package cli

import (
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/x/bech32ibc/types"
	"github.com/spf13/cobra"
)

var (
	IcsToHeightOffset  = "ics-to-height-offset"
	IcsToTimeoutOffset = "ics-to-timeout-offset"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "bech32ibc transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	return txCmd
}

func CmdBech32IBCProposalHandler() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bech32ibc",
		Short: "Bech32 IBC proposals",
	}

	cmd.AddCommand(NewCmdSubmitUpdateHrpIbcRecordProposal())

	return cmd
}

func NewCmdSubmitUpdateHrpIbcRecordProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-hrp-ibc-record [human-readable-prefix] [channel-id]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit an update to map a bech32 prefix to a channel id",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hrp := args[0]
			channelId := args[1]

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
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

			heightOffset, err := cmd.Flags().GetUint64(IcsToHeightOffset)
			if err != nil {
				return err
			}

			durationOffsetText, err := cmd.Flags().GetString(IcsToTimeoutOffset)
			if err != nil {
				return err
			}
			durationOffset, err := time.ParseDuration(durationOffsetText)

			content := types.NewUpdateHrpIBCRecordProposal(title, description, hrp, channelId, heightOffset, durationOffset)

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().Uint64(IcsToHeightOffset, 0, "timeout for IBC routed packets through this channel, in blocks. A value of X here, means that if a packet is attempted to get relayed at counter-party chain height of N, and fails to be ack'd by height N+X, the packet will bounce back to the source chain.")
	cmd.Flags().String(IcsToTimeoutOffset, "", "the offset of timeout to expire on target chain")
	flags.AddTxFlagsToCmd(cmd)
	cmd.MarkFlagRequired(cli.FlagTitle)
	cmd.MarkFlagRequired(cli.FlagDescription)

	return cmd
}
