package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/spf13/cobra"
	"github.com/vizualni/whoops"
)

func applyFlags(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")

	cmd.MarkFlagRequired(cli.FlagTitle)
	cmd.MarkFlagRequired(cli.FlagDescription)
}

func CmdEvmChainProposalHandler() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evm",
		Short: "EVM proposals",
	}

	cmd.AddCommand(CmdEvmProposeNewChain())
	cmd.AddCommand(CmdEvmProposeChainRemoval())
	cmd.AddCommand(CmdEvmProposalDeployNewSmartContract())

	return cmd
}

// isHexCharacter returns bool of c being a valid hexadecimal.
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}

func CmdEvmProposeNewChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-new-chain [chain-id] [block-height] [block-hash-at-height]",
		Short: "Proposal to add a new EVM chain",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return whoops.Try(func() {
				clientCtx, err := client.GetClientTxContext(cmd)
				whoops.Assert(err)

				chainID, blockHeightStr, blockHashAtHeight := args[0], args[1], args[2]

				blockHeight, err := strconv.ParseInt(blockHeightStr, 10, 64)
				whoops.Assert(err)

				if len(blockHashAtHeight) <= len("0x") || !isHex(blockHashAtHeight[2:]) {
					whoops.Assert(whoops.String("invalid block hash"))
				}

				addChainProposal := &types.AddChainProposal{
					Title:             whoops.Must(cmd.Flags().GetString(cli.FlagTitle)),
					Description:       whoops.Must(cmd.Flags().GetString(cli.FlagDescription)),
					ChainID:           chainID,
					BlockHeight:       uint64(blockHeight),
					BlockHashAtHeight: blockHashAtHeight,
				}

				from := clientCtx.GetFromAddress()

				depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
				whoops.Assert(err)

				deposit, err := sdk.ParseCoinsNormalized(depositStr)
				whoops.Assert(err)

				msg, err := gov.NewMsgSubmitProposal(addChainProposal, deposit, from)
				whoops.Assert(err)

				err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
				whoops.Assert(err)
			})
		},
	}
	applyFlags(cmd)

	return cmd
}

func CmdEvmProposalDeployNewSmartContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-new-smart-contract [chain-id] [abi-json] [bytecode-hey]",
		Short: "Proposal to add a new EVM chain",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return whoops.Try(func() {
				clientCtx, err := client.GetClientTxContext(cmd)
				whoops.Assert(err)

				chainID, abiJSON, bytecodeHex := args[0], args[1], args[2]

				deployNewSmartContractProposal := &types.DeployNewSmartContractProposal{
					Title:       whoops.Must(cmd.Flags().GetString(cli.FlagTitle)),
					Description: whoops.Must(cmd.Flags().GetString(cli.FlagDescription)),
					ChainID:     chainID,
					AbiJSON:     abiJSON,
					BytecodeHex: bytecodeHex,
				}

				from := clientCtx.GetFromAddress()

				depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
				whoops.Assert(err)

				deposit, err := sdk.ParseCoinsNormalized(depositStr)
				whoops.Assert(err)

				msg, err := gov.NewMsgSubmitProposal(deployNewSmartContractProposal, deposit, from)
				whoops.Assert(err)

				err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
				whoops.Assert(err)
			})
		},
	}
	applyFlags(cmd)

	return cmd
}

func CmdEvmProposeChainRemoval() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-chain-removal [chain-id]",
		Short: "Proposal to remove an existing EVM chain from Paloma",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return whoops.Try(func() {
				clientCtx, err := client.GetClientTxContext(cmd)
				whoops.Assert(err)

				chainID := args[0]

				addChainProposal := &types.RemoveChainProposal{
					Title:       whoops.Must(cmd.Flags().GetString(cli.FlagTitle)),
					Description: whoops.Must(cmd.Flags().GetString(cli.FlagDescription)),
					ChainID:     chainID,
				}

				from := clientCtx.GetFromAddress()

				depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
				whoops.Assert(err)

				deposit, err := sdk.ParseCoinsNormalized(depositStr)
				whoops.Assert(err)

				msg, err := gov.NewMsgSubmitProposal(addChainProposal, deposit, from)
				whoops.Assert(err)

				err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
				whoops.Assert(err)
			})
		},
	}
	applyFlags(cmd)

	return cmd
}
