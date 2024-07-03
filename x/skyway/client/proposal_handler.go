package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/palomachain/paloma/x/skyway/client/cli"
)

var ProposalHandler = govclient.NewProposalHandler(cli.CmdSkywayProposalHandler)
