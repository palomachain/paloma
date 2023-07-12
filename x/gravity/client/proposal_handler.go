package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/palomachain/paloma/x/gravity/client/cli"
)

// ProposalHandler is the community Ethereum spend proposal handler.
var ProposalHandler = govclient.NewProposalHandler(cli.CmdSubmitCommunityPoolEthereumSpendProposal)
