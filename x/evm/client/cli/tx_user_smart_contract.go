package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/evm/types"
	vtypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cobra"
)

func CmdUploadUserSmartContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload-user-smart-contract [title] [abi-json] [bytecode] [constructor-input]",
		Short: "Upload a user-defined smart contract to paloma",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to retrieve ctx: %w", err)
			}

			title, abi, bytecode, input := args[0], args[1], args[2], args[3]

			validator := sdk.ValAddress(clientCtx.GetFromAddress().Bytes())
			creator := clientCtx.GetFromAddress().String()

			msg := &types.MsgUploadUserSmartContractRequest{
				ValAddress:       validator.String(),
				Title:            title,
				AbiJson:          abi,
				Bytecode:         bytecode,
				ConstructorInput: input,
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

func CmdRemoveUserSmartContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-user-smart-contract [id]",
		Short: "Remove a user-defined smart contract from paloma",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to retrieve ctx: %w", err)
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			validator := sdk.ValAddress(clientCtx.GetFromAddress().Bytes())
			creator := clientCtx.GetFromAddress().String()

			msg := &types.MsgRemoveUserSmartContractRequest{
				ValAddress: validator.String(),
				Id:         id,
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
