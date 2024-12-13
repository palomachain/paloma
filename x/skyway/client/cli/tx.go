package cli

import (
	"fmt"
	"strconv"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/skyway/types"
	tokenfactorytypes "github.com/palomachain/paloma/v2/x/tokenfactory/types"
	vtypes "github.com/palomachain/paloma/v2/x/valset/types"
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
		CmdSetErc20ToTokenDenom(),
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

func CmdSetErc20ToTokenDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-erc20-to-token-denom [erc20-address] [chain-reference-id] [token-denom]",
		Short: "Creates or updates the ERC20 tracking address mapping for the given denom. Must have admin authority to do so.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()
			erc20, err := types.NewEthAddress(args[0])
			if err != nil {
				return err
			}
			if erc20 == nil {
				return fmt.Errorf("invalid eth address")
			}
			chainReferenceID := args[1]
			denom := args[2]
			if _, _, err := tokenfactorytypes.DeconstructDenom(denom); err != nil {
				return err
			}

			msg := types.NewMsgSetERC20ToTokenDenom(sender, *erc20, chainReferenceID, denom)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Send it
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
