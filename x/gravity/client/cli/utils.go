package cli

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/palomachain/paloma/x/gravity/types"
)

// ParseCommunityPoolEthereumSpendProposal reads and parses a CommunityPoolEthereumSpendProposalForCLI from a file.
func ParseCommunityPoolEthereumSpendProposal(cdc codec.JSONCodec, proposalFile string) (types.CommunityPoolEthereumSpendProposalForCLI, error) {
	proposal := types.CommunityPoolEthereumSpendProposalForCLI{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
