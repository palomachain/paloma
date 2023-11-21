package client

import (
	govclient "cosmossdk.io/x/gov/client"
	"github.com/palomachain/paloma/x/gravity/client/cli"
)

var ProposalHandler = govclient.NewProposalHandler(cli.CmdGravityProposalHandler)
