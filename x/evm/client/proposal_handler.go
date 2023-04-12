package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/palomachain/paloma/x/evm/client/cli"
)

var ProposalHandler = govclient.NewProposalHandler(cli.CmdEvmChainProposalHandler)
