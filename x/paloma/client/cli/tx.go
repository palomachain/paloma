package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/paloma/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
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

	cmd.AddCommand(CmdRegisterLightNodeClient())
	cmd.AddCommand(CmdAddLightNodeClientFunds())

	return cmd
}

func CmdRegisterLightNodeClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-light-node-client [client-address]",
		Short: "Registers a new light node client",
		Long: `Registers a new light node client, if the client has pre-paid for it, and it has not been activated yet.
The creator key is used to determine the available funds, which are transferred to the new client, in a vesting account.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientAddress := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			creator := clientCtx.GetFromAddress().String()
			msg := &types.MsgRegisterLightNodeClient{
				ClientAddress: clientAddress,
				Metadata: valsettypes.MsgMetadata{
					Creator: creator,
					Signers: []string{creator},
				},
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdAddLightNodeClientFunds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-light-node-client-funds [auth-address] [amount]",
		Short: "Manually add funds to a new light node client",
		Long: `Manually register a light node sale by adding funds to the authorization account. Funds will be added to any existing funds on the same account.
The [auth-address] field should contain the address of the client that's going to register the new light node client.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			authAddress := args[0]

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return errors.New("invalid amount")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			acct, err := sdk.AccAddressFromBech32(authAddress)
			if err != nil {
				return err
			}

			if sdk.VerifyAddressFormat(acct) != nil {
				return errors.New("invalid client_address")
			}

			creator := clientCtx.GetFromAddress().String()
			msg := &types.MsgAddLightNodeClientFunds{
				AuthAddress: authAddress,
				Amount:      amount[0],
				Metadata: valsettypes.MsgMetadata{
					Creator: creator,
					Signers: []string{creator},
				},
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
