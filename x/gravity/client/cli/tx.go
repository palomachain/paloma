package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/palomachain/paloma/x/gravity/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/spf13/cobra"
)

// GetTxCmd bundles all the subcmds together so they appear under `gravity tx`
func GetTxCmd(storeKey string) *cobra.Command {
	// needed for governance proposal txs in cli case
	// internal check prevents double registration in node case
	keeper.RegisterProposalTypes()

	// nolint: exhaustruct
	gravityTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Gravity transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	gravityTxCmd.AddCommand([]*cobra.Command{
		CmdSendToEth(),
		CmdCancelSendToEth(),
		CmdRequestBatch(),
		CmdSetOrchestratorAddress(),
		CmdExecutePendingIbcAutoForwards(),
	}...)

	return gravityTxCmd
}

// CmdSendToEth sends tokens to Ethereum. Locks Cosmos-side tokens into the Transaction pool for batching.
func CmdSendToEth() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "send-to-eth [eth-dest] [amount] [bridge-fee] [chain-fee]",
		Short: "Adds a new entry to the transaction pool to withdraw an amount from the Ethereum bridge contract. This will not execute until a batch is requested and then actually relayed. Chain fee must be at least min_chain_fee_basis_points in `query gravity params`. Your funds can be reclaimed using cancel-send-to-eth so long as they remain in the pool",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cosmosAddr := cliCtx.GetFromAddress()

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "amount")
			}
			bridgeFee, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return sdkerrors.Wrap(err, "bridge fee")
			}
			chainFee, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return sdkerrors.Wrap(err, "chain fee")
			}

			ethAddr, err := types.NewEthAddress(args[0])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid eth address")
			}

			if len(amount) != 1 || len(bridgeFee) != 1 || len(chainFee) != 1 {
				return fmt.Errorf("unexpected coin amounts, expecting just 1 coin amount for both amount and bridgeFee")
			}

			// Make the message
			msg := types.MsgSendToEth{
				Sender:    cosmosAddr.String(),
				EthDest:   ethAddr.GetAddress().Hex(),
				Amount:    amount[0],
				BridgeFee: bridgeFee[0],
				ChainFee:  chainFee[0],
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

// CmdCancelSendToEth enables users to take their Transaction out of the pool. Note that this cannot be done if it is
// locked up in a pending batch or if it has already been executed on Ethereum
func CmdCancelSendToEth() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "cancel-send-to-eth [transaction id]",
		Short: "Removes an entry from the transaction pool, preventing your tokens from going to Ethereum and refunding the send.",
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
			msg := types.MsgCancelSendToEth{
				Sender:        cosmosAddr.String(),
				TransactionId: txId,
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

// CmdRequestBatch requests that the validators create and confirm a batch to be sent to Ethereum. This
// is a manual command which duplicates the efforts of the Ethereum Relayer, likely not to be used often
func CmdRequestBatch() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "request-batch [chain-reference-id]",
		Short: "Request a new batch on the cosmos side for pooled withdrawal transactions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cosmosAddr := cliCtx.GetFromAddress()

			msg := types.MsgRequestBatch{
				Sender:           cosmosAddr.String(),
				ChainReferenceId: args[0],
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

// CmdSetOrchestratorAddress registers delegate keys for a validator so that their Orchestrator has authority to perform
// its responsibility
func CmdSetOrchestratorAddress() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "set-orchestrator-address [validator-address] [orchestrator-address] [ethereum-address]",
		Short: "Allows validators to delegate their voting responsibilities to a given key.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.MsgSetOrchestratorAddress{
				Validator:    args[0],
				Orchestrator: args[1],
				EthAddress:   args[2],
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

// CmdExecutePendingIbcAutoForwards Executes a number of queued IBC Auto Forwards. When users perform a Send to Cosmos
// with a registered foreign address prefix (e.g. canto1... cre1...), their funds will be locked in the Gravity module
// until their pending forward is executed. This will send the funds to the equivalent gravity-prefixed account and then
// immediately create an IBC transfer to the destination chain to the original foreign account. If there is an IBC
// failure, the funds will be deposited on the gravity-prefixed account.
func CmdExecutePendingIbcAutoForwards() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "execute-pending-ibc-auto-forwards [forwards-to-execute]",
		Short: "Executes a given number of IBC Auto-Forwards",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := cliCtx.GetFromAddress()
			if sender.String() == "" {
				return fmt.Errorf("from address must be specified")
			}
			forwardsToClear, err := strconv.ParseUint(args[0], 10, 0)
			if err != nil {
				return sdkerrors.Wrap(err, "Unable to parse forwards-to-execute as an non-negative integer")
			}
			msg := types.MsgExecuteIbcAutoForwards{
				ForwardsToClear: forwardsToClear,
				Executor:        cliCtx.GetFromAddress().String(),
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
