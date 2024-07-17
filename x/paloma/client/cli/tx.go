package cli

import (
	"context"
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
	"google.golang.org/protobuf/types/known/emptypb"
)

const flagFeegrant = "feegrant"

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
The client must have enough tokens to pay for fees, or have a feegrant set.`,
		Args: cobra.ExactArgs(0),
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
		Short: "Manually add a license to a new light node client",
		Long: `Manually register a light node license by locking funds to the registered account. Funds will be added to any existing funds on the same account.
The [client-address] field should contain the address of the new light node client.
The [vesting-months] parameter determines how long the funds will take to fully vest.
The --feegrant flag adds a feegrant to the license, so that transaction fees are covered for this account. The account that pays for the fees is determined by governance vote.`,
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

			feegrant, err := cmd.Flags().GetBool(flagFeegrant)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if feegrant {
				// If we have the feegrant flag, we first check if the
				// feegranter is defined
				qryCtx := client.GetClientContextFromCmd(cmd)

				queryClient := types.NewQueryClient(qryCtx)
				_, err := queryClient.GetLightNodeClientFeegranter(
					context.Background(), &emptypb.Empty{})
				if err != nil {
					return err
				}
			}

			creator := clientCtx.GetFromAddress().String()
			msg := &types.MsgAddLightNodeClientLicense{
				ClientAddress: clientAddress,
				Amount:        amount[0],
				VestingMonths: uint32(vestingMonths),
				Feegrant:      feegrant,
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

	cmd.Flags().Bool(flagFeegrant, false, "Grant fee usage to new account")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
