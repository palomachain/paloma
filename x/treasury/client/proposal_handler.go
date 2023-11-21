package client

import (
	govclient "cosmossdk.io/x/gov/client"
	"github.com/palomachain/paloma/x/treasury/client/cli"
)

var ProposalHandler = govclient.NewProposalHandler(cli.CmdTreasuryProposalHandler)
