package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/palomachain/paloma/x/bech32ibc/client/cli"
)

var ProposalHandler = govclient.NewProposalHandler(cli.CmdBech32IBCProposalHandler)
