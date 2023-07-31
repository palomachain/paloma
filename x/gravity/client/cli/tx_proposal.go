package cli

import (
	"encoding/json"
	"fmt"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/palomachain/paloma/x/gravity/types"
)

func CmdGravityProposalHandler() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gravity",
		Short: "Gravity proposals",
	}
	cmd.AddCommand([]*cobra.Command{
		CmdGovIbcMetadataProposal(),
	}...)

	return cmd
}

// CmdGovIbcMetadataProposal enables users to easily submit json file proposals for IBC Metadata registration, needed to
// send Cosmos tokens over to Ethereum
func CmdGovIbcMetadataProposal() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "gov-ibc-metadata [path-to-proposal-json] [initial-deposit]",
		Short: "Creates a governance proposal to set the Metadata of the given IBC token. Once the metadata is set this token can be moved to Ethereum using Gravity Bridge",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cosmosAddr := cliCtx.GetFromAddress()

			initialDeposit, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "bad initial deposit amount")
			}

			if len(initialDeposit) != 1 {
				return fmt.Errorf("unexpected coin amounts, expecting just 1 coin amount for initialDeposit")
			}

			proposalFile := args[0]

			contents, err := os.ReadFile(proposalFile)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to read proposal json file")
			}

			proposal := &types.IBCMetadataProposal{}
			err = json.Unmarshal(contents, proposal)
			if err != nil {
				return sdkerrors.Wrap(err, "proposal json file is not valid json")
			}
			if proposal.IbcDenom == "" ||
				proposal.Title == "" ||
				proposal.Description == "" ||
				proposal.Metadata.Base == "" ||
				proposal.Metadata.Name == "" ||
				proposal.Metadata.Display == "" ||
				proposal.Metadata.Symbol == "" {
				return fmt.Errorf("proposal json file is not valid, please check example json in docs")
			}

			// checks if the provided token denom is a proper IBC token, not a native token.
			if !strings.HasPrefix(proposal.IbcDenom, "ibc/") && !strings.HasPrefix(proposal.IbcDenom, "IBC/") {
				return sdkerrors.Wrap(types.ErrInvalid, "Target denom is not an IBC token")
			}

			// check that our base unit is the IBC token name on this chain. This makes setting/loading denom
			// metadata work out, as SetDenomMetadata uses the base denom as an index
			if proposal.Metadata.Base != proposal.IbcDenom {
				return sdkerrors.Wrap(types.ErrInvalid, "Metadata base must be the same as the IBC denom!")
			}

			metadataErr := proposal.Metadata.Validate()
			if metadataErr != nil {
				return sdkerrors.Wrap(metadataErr, "invalid metadata or proposal details!")
			}

			queryClientBank := banktypes.NewQueryClient(cliCtx)
			_, err = queryClientBank.DenomMetadata(cmd.Context(), &banktypes.QueryDenomMetadataRequest{Denom: proposal.IbcDenom})
			if err == nil {
				return sdkerrors.Wrap(metadataErr, "Attempting to set the metadata for a token that already has metadata!")
			}

			supply, err := queryClientBank.SupplyOf(cmd.Context(), &banktypes.QuerySupplyOfRequest{Denom: proposal.IbcDenom})
			if err != nil {
				return sdkerrors.Wrap(types.ErrInternal, "Failed to get supply data?")
			}
			if supply.GetAmount().Amount.Equal(sdk.ZeroInt()) {
				return sdkerrors.Wrap(types.ErrInvalid, "This ibc hash does not seem to exist on Gravity, are you sure you have the right one?")
			}

			proposalAny, err := codectypes.NewAnyWithValue(proposal)
			if err != nil {
				return sdkerrors.Wrap(err, "invalid metadata or proposal details!")
			}

			// Make the message
			msg := govv1beta1types.MsgSubmitProposal{
				Proposer:       cosmosAddr.String(),
				InitialDeposit: initialDeposit,
				Content:        proposalAny,
			}
			if err := msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(err, "Your proposal.json is not valid, please correct it")
			}
			// Send it
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
