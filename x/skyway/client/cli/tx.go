package cli

import (
	"strconv"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/skyway/types"
	vtypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cobra"
)

// GetTxCmd bundles all the subcmds together so they appear under `skyway tx`
func GetTxCmd(storeKey string) *cobra.Command {
	// nolint: exhaustruct
	skywayTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Skyway transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	skywayTxCmd.AddCommand([]*cobra.Command{
		CmdSendToRemote(),
		CmdCancelSendToRemote(),
	}...)

	return skywayTxCmd
}

// CmdSendToRemote sends tokens to Ethereum. Locks Cosmos-side tokens into the Transaction pool for batching.
func CmdSendToRemote() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "send-tx [remote-chain-dest-address] [amount] [chain-reference-id]",
		Short: "Adds a new entry to the transaction pool to withdraw an amount from the remote chain bridge contract. This will not execute until a batch is requested and then actually relayed. Your funds can be reclaimed using cancel-tx so long as they remain in the pool",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cosmosAddr := cliCtx.GetFromAddress()

			ethDest := args[0]

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "amount")
			}

			chainReferenceID := args[2]

			// Make the message
			msg := types.MsgSendToRemote{
				EthDest:          ethDest,
				Amount:           amount[0],
				ChainReferenceId: chainReferenceID,
				Metadata: vtypes.MsgMetadata{
					Creator: cosmosAddr.String(),
					Signers: []string{cosmosAddr.String()},
				},
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			// Send it
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCancelSendToRemote enables users to take their Transaction out of the pool. Note that this cannot be done if it is
// locked up in a pending batch or if it has already been executed on Ethereum
func CmdCancelSendToRemote() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "cancel-tx [transaction id]",
		Short: "Removes an entry from the transaction pool, preventing your tokens from going to the remote chain and refunding the send.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cosmosAddr := cliCtx.GetFromAddress()

			txId, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse transaction id")
			}

			// Make the message
			msg := types.MsgCancelSendToRemote{
				TransactionId: txId,
				Metadata: vtypes.MsgMetadata{
					Creator: cosmosAddr.String(),
					Signers: []string{cosmosAddr.String()},
				},
			}
			// Send it
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
