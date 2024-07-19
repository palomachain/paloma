package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/palomachain/paloma/x/paloma/client/cli"
)

var ProposalHandler = govclient.NewProposalHandler(cli.CmdPalomaChainProposalHandler)
