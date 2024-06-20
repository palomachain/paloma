package cli

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/treasury/types"
	vtypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cobra"
)

var DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdUpsertRelayerFee())
	return cmd
}

func CmdUpsertRelayerFee() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upsert-relayer-fee [chain-reference-id] [fee-multiplicator]",
		Short: "Sets the relayer fee for the sender to the given multiplicator.",
		Long: `
    Sets the relayer fee for the sender to the given multiplicator value.
    The multiplicator determines the total fee a relayer may claim based 
    on the total cost of a relayed transaction.

    In order to aim for an even reimbursement, the multiplicator value 
    must be 1.
    In order to make a profit, the multiplicator value must be higher
    than 1.

    Example: Using a fee multiplicator value of 1.1 entitles the sender
    to a claimable reimbursement of 11uCOIN for a transaction relayed
    totalling in 10uCOIN of GAS.
    `,
		Example: "upsert-relayer-fee eth-main 1.05",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id := args[0]
			fee := math.LegacyMustNewDecFromStr(args[1])

			if fee.LT(math.LegacyMustNewDecFromStr("1")) {
				return fmt.Errorf("loss incurring fees not allowed")
			}

			validator := sdk.ValAddress(clientCtx.GetFromAddress().Bytes())
			creator := clientCtx.GetFromAddress().String()

			msg := &types.MsgUpsertRelayerFee{
				FeeSetting: &types.RelayerFeeSetting{
					ValAddress: validator.String(),
					Fees: []types.RelayerFeeSetting_FeeSetting{
						{
							ChainReferenceId: id,
							Multiplicator:    fee,
						},
					},
				},
				Metadata: vtypes.MsgMetadata{
					Creator: creator,
					Signers: []string{creator},
				},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
