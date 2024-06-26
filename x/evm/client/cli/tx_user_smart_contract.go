package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/palomachain/paloma/x/evm/types"
	vtypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/spf13/cobra"
)

func getParamOrFile(param string) (string, error) {
	if _, err := os.Stat(param); err == nil {
		// File exists, read it
		data, err := os.ReadFile(param)
		if err != nil {
			return "", err
		}

		return string(data), nil
	}

	// If file does not exist, return the parameter directly
	return param, nil
}

func CmdUploadUserSmartContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload-user-smart-contract [title] [abi-json] [bytecode] [constructor-input]",
		Short: "Upload a user-defined smart contract to paloma",
		Long: `Upload a new user-defined smart contract. The smart contract will stay in paloma and can afterwads be deployed to any external chain.
The contract will be associated with the validator address used to sign the message.
The [abi-json], [bytecode] and [constructor-input] parameters can either be passed as strings directly or as filenames.`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to retrieve ctx: %w", err)
			}

			title := args[0]

			abi, err := getParamOrFile(args[1])
			if err != nil {
				return err
			}

			bytecode, err := getParamOrFile(args[2])
			if err != nil {
				return err
			}

			input, err := getParamOrFile(args[3])
			if err != nil {
				return err
			}

			creator := clientCtx.GetFromAddress().String()

			msg := &types.MsgUploadUserSmartContractRequest{
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
		Long:  `Remove a user-defined smart contract. The contract will no longer be available to be deploy in external chains.`,
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

			creator := clientCtx.GetFromAddress().String()

			msg := &types.MsgRemoveUserSmartContractRequest{
				Id: id,
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

func CmdDeployUserSmartContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-user-smart-contract [id] [target-chain]",
		Short: "Deploy a user-defined smart contract to a target chain",
		Long:  `Deploy a previously uploaded smart contract to a target chain. After issuing the request, the deployment will be listed as "IN_FLIGHT" for the specified chain. When deployed successfully, the status changes to "ACTIVE" and the "address" field will contain the contract address on the target chain.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to retrieve ctx: %w", err)
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			targetChain := args[1]

			creator := clientCtx.GetFromAddress().String()

			msg := &types.MsgDeployUserSmartContractRequest{
				Id:          id,
				TargetChain: targetChain,
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
