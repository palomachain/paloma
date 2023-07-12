package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"

	"github.com/palomachain/paloma/x/gravity/types"
)

func GetTxCmd(storeKey string) *cobra.Command {
	gravityTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Gravity transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	gravityTxCmd.AddCommand(
		CmdSendToEthereum(),
		CmdCancelSendToEthereum(),
		CmdSetDelegateKeys(),
	)

	return gravityTxCmd
}

func CmdSendToEthereum() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "send-to-ethereum [ethereum-reciever] [send-coins] [fee-coins]",
		Aliases: []string{"send", "transfer"},
		Args:    cobra.ExactArgs(3),
		Short:   "Send tokens from cosmos chain to connected ethereum chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			if from == nil {
				return fmt.Errorf("must pass from flag")
			}

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("must be a valid ethereum address got %s", args[0])
			}

			// Get amount of coins
			sendCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			feeCoin, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgSendToEthereum(from, common.HexToAddress(args[0]).Hex(), sendCoin, feeCoin)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdCancelSendToEthereum() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-send-to-ethereum [id]",
		Args:  cobra.ExactArgs(1),
		Short: "Cancel ethereum send by id",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			if from == nil {
				return fmt.Errorf("must pass from flag")
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgCancelSendToEthereum(id, from)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSetDelegateKeys() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-delegate-keys [validator-address] [orchestrator-address] [ethereum-address] [ethereum-signature]",
		Args:  cobra.ExactArgs(4),
		Short: "Set gravity delegate keys",
		Long: `Set a validator's Ethereum and orchestrator addresses. The validator must
sign over a binary Proto-encoded DelegateKeysSignMsg message. The message contains
the validator's address and operator account current nonce.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			orcAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			ethAddr, err := parseContractAddress(args[2])
			if err != nil {
				return err
			}

			ethSig, err := hexutil.Decode(args[3])
			if err != nil {
				return err
			}

			msg := types.NewMsgDelegateKeys(valAddr, orcAddr, ethAddr, ethSig)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSubmitCommunityPoolEthereumSpendProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community-pool-ethereum-spend [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a community pool Ethereum spend proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a community pool Ethereum spend proposal along with an initial deposit.
The proposal details must be supplied via a JSON file. The funds from the community pool
will be bridged to Ethereum to the supplied recipient Ethereum address. Only one denomination
of Cosmos token can be sent, and the bridge fee supplied along with the amount must be of the
same denomination.

Example:
$ %s tx gov submit-proposal community-pool-ethereum-spend <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
	"title": "Community Pool Ethereum Spend",
	"description": "Bridge me some tokens to Ethereum!",
	"recipient": "0x0000000000000000000000000000000000000000",
	"amount": "20000ugrain",
	"bridge_fee": "1000ugrain",
	"deposit": "1000ugrain"
}
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := ParseCommunityPoolEthereumSpendProposal(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			if len(proposal.Title) == 0 {
				return fmt.Errorf("title is empty")
			}

			if len(proposal.Description) == 0 {
				return fmt.Errorf("description is empty")
			}

			if !common.IsHexAddress(proposal.Recipient) {
				return fmt.Errorf("recipient is not a valid Ethereum address")
			}

			amount, err := sdk.ParseCoinNormalized(proposal.Amount)
			if err != nil {
				return err
			}

			bridgeFee, err := sdk.ParseCoinNormalized(proposal.BridgeFee)
			if err != nil {
				return err
			}

			if amount.Denom != bridgeFee.Denom {
				return fmt.Errorf("amount and bridge fee denominations must match")
			}

			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			content := types.NewCommunityPoolEthereumSpendProposal(proposal.Title, proposal.Description, proposal.Recipient, amount, bridgeFee)

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
