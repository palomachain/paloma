package cli

import (
	"errors"
	"fmt"
	"strconv"
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
	cmd.AddCommand(CmdAddLightNodeClientLicense())

	return cmd
}

func CmdRegisterLightNodeClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-light-node-client",
		Short: "Registers a new light node client",
		Long: `Registers a new light node client, if the client has pre-paid for it, and it has not been activated yet.
The creator key is used to determine the available funds, which are transferred to the new client, in a vesting account.
The creator should already have a feegrant allowance from the paloma feegranter.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			creator := clientCtx.GetFromAddress().String()
			msg := &types.MsgRegisterLightNodeClient{
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

func CmdAddLightNodeClientLicense() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-light-node-client-license [client-address] [amount] [vesting-months]",
		Short: "Manually adds a license for a new light node client",
		Long: `Manually adds a light node license by locking funds to the registered account.
The [client-address] field should contain the address of the light node client to be registered later.
The [vesting-months] parameter determines how long the funds will take to fully vest.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientAddress := args[0]

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return errors.New("invalid amount")
			}

			acct, err := sdk.AccAddressFromBech32(clientAddress)
			if err != nil {
				return err
			}

			if sdk.VerifyAddressFormat(acct) != nil {
				return errors.New("invalid client_address")
			}

			vestingMonths, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			creator := clientCtx.GetFromAddress().String()
			msg := &types.MsgAddLightNodeClientLicense{
				ClientAddress: clientAddress,
				Amount:        amount[0],
				VestingMonths: uint32(vestingMonths),
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
