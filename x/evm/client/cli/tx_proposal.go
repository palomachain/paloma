package cli

import (
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/spf13/cobra"
)

func applyFlags(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")

	//nolint:errcheck
	cmd.MarkFlagRequired(cli.FlagTitle)
	//nolint:errcheck
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
	cmd.AddCommand(CmdEvmProposalChangeMinOnChainBalance())
	cmd.AddCommand(CmdEvmProposalChangeRelayWeights())

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
		Use:   "propose-new-chain [chain-reference-id] [chain-id] [min-on-chain-balance] [block-height] [block-hash-at-height]",
		Short: "Proposal to add a new EVM chain",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return whoops.Try(func() {
				clientCtx, err := client.GetClientTxContext(cmd)
				whoops.Assert(err)

				chainReferenceID, chainIDStr, minOnChainBalance, blockHeightStr, blockHashAtHeight := args[0], args[1], args[2], args[3], args[4]

				blockHeight, err := strconv.ParseInt(blockHeightStr, 10, 64)
				whoops.Assert(err)

				chainID, err := strconv.ParseInt(chainIDStr, 10, 64)
				whoops.Assert(err)

				if len(blockHashAtHeight) <= len("0x") || !isHex(blockHashAtHeight[2:]) {
					whoops.Assert(whoops.String("invalid block hash"))
				}

				_, ok := new(big.Int).SetString(minOnChainBalance, 10)
				if !ok {
					whoops.Assert(whoops.String("minimum on change balance is incorret"))
				}

				addChainProposal := &types.AddChainProposal{
					ChainReferenceID:  chainReferenceID,
					ChainID:           uint64(chainID),
					Title:             whoops.Must(cmd.Flags().GetString(cli.FlagTitle)),
					Description:       whoops.Must(cmd.Flags().GetString(cli.FlagDescription)),
					BlockHeight:       uint64(blockHeight),
					BlockHashAtHeight: blockHashAtHeight,
					MinOnChainBalance: minOnChainBalance,
				}

				from := clientCtx.GetFromAddress()

				depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
				whoops.Assert(err)

				deposit, err := sdk.ParseCoinsNormalized(depositStr)
				whoops.Assert(err)

				msg, err := govv1beta1types.NewMsgSubmitProposal(addChainProposal, deposit, from)
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
		Use:   "propose-new-smart-contract [abi-json] [bytecode-hex]",
		Short: "Proposal to add a new EVM chain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return whoops.Try(func() {
				clientCtx, err := client.GetClientTxContext(cmd)
				whoops.Assert(err)

				abiJSON, bytecodeHex := args[0], args[1]

				deployNewSmartContractProposal := &types.DeployNewSmartContractProposal{
					Title:       whoops.Must(cmd.Flags().GetString(cli.FlagTitle)),
					Description: whoops.Must(cmd.Flags().GetString(cli.FlagDescription)),
					AbiJSON:     abiJSON,
					BytecodeHex: bytecodeHex,
				}

				from := clientCtx.GetFromAddress()

				depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
				whoops.Assert(err)

				deposit, err := sdk.ParseCoinsNormalized(depositStr)
				whoops.Assert(err)

				msg, err := govv1beta1types.NewMsgSubmitProposal(deployNewSmartContractProposal, deposit, from)
				whoops.Assert(err)

				err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
				whoops.Assert(err)
			})
		},
	}
	applyFlags(cmd)

	return cmd
}

func CmdEvmProposalChangeMinOnChainBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "change-min-on-chain-balance [chain-reference-id] [balance]",
		Short:   "Changes the min on chain balance for a given EVM chain referenced by the chain-reference-id",
		Example: "change-min-on-chain-balance eth-main 50000000",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return whoops.Try(func() {
				clientCtx, err := client.GetClientTxContext(cmd)
				whoops.Assert(err)

				chainReferenceID, balance := args[0], args[1]

				_, ok := new(big.Int).SetString(balance, 10)
				if !ok {
					whoops.Assert(whoops.String("invalid provided balance"))
				}

				proposal := &types.ChangeMinOnChainBalanceProposal{
					Title:             whoops.Must(cmd.Flags().GetString(cli.FlagTitle)),
					Description:       whoops.Must(cmd.Flags().GetString(cli.FlagDescription)),
					ChainReferenceID:  chainReferenceID,
					MinOnChainBalance: balance,
				}

				from := clientCtx.GetFromAddress()

				depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
				whoops.Assert(err)

				deposit, err := sdk.ParseCoinsNormalized(depositStr)
				whoops.Assert(err)

				msg, err := govv1beta1types.NewMsgSubmitProposal(proposal, deposit, from)
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
		Use:   "propose-chain-removal [chain-reference-id]",
		Short: "Proposal to remove an existing EVM chain from Paloma",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return whoops.Try(func() {
				clientCtx, err := client.GetClientTxContext(cmd)
				whoops.Assert(err)

				chainReferenceID := args[0]

				addChainProposal := &types.RemoveChainProposal{
					Title:            whoops.Must(cmd.Flags().GetString(cli.FlagTitle)),
					Description:      whoops.Must(cmd.Flags().GetString(cli.FlagDescription)),
					ChainReferenceID: chainReferenceID,
				}

				from := clientCtx.GetFromAddress()

				depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
				whoops.Assert(err)

				deposit, err := sdk.ParseCoinsNormalized(depositStr)
				whoops.Assert(err)

				msg, err := govv1beta1types.NewMsgSubmitProposal(addChainProposal, deposit, from)
				whoops.Assert(err)

				err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
				whoops.Assert(err)
			})
		},
	}
	applyFlags(cmd)

	return cmd
}

func CmdEvmProposalChangeRelayWeights() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "propose-relay-weights [chain-reference-id] [weights]",
		Short:   "Changes the relay weights for a given EVM chain referenced by the chain-reference-id",
		Example: "change-relay-weights eth-main {\"fee\": 0.50, \"uptime\": 0.75, \"successRate\": 0.90, \"executionTime\": 0.20}",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return whoops.Try(func() {
				clientCtx, err := client.GetClientTxContext(cmd)
				whoops.Assert(err)

				chainReferenceID, weightsString := args[0], args[1]

				var weights types.RelayWeights
				err = json.Unmarshal([]byte(weightsString), &weights)
				whoops.Assert(err)

				proposal := &types.RelayWeightsProposal{
					Title:            whoops.Must(cmd.Flags().GetString(cli.FlagTitle)),
					Description:      whoops.Must(cmd.Flags().GetString(cli.FlagDescription)),
					ChainReferenceID: chainReferenceID,
					Fee:              weights.Fee,
					Uptime:           weights.Uptime,
					SuccessRate:      weights.SuccessRate,
					ExecutionTime:    weights.ExecutionTime,
				}

				from := clientCtx.GetFromAddress()

				depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
				whoops.Assert(err)

				deposit, err := sdk.ParseCoinsNormalized(depositStr)
				whoops.Assert(err)

				msg, err := govv1beta1types.NewMsgSubmitProposal(proposal, deposit, from)
				whoops.Assert(err)

				err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
				whoops.Assert(err)
			})
		},
	}
	applyFlags(cmd)

	return cmd
}
