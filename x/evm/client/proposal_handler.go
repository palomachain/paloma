package client

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	"github.com/palomachain/paloma/x/evm/client/cli"
)

var ProposalHandler = govclient.NewProposalHandler(cli.CmdEvmChainProposalHandler, func(client.Context) rest.ProposalRESTHandler {
	return rest.ProposalRESTHandler{
		SubRoute: "evm-not-implemented",
		Handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprintf(w, "please use cli to issue a new EVM proposal or make a new github ticket if you wish to have this functionality via rest")
		},
	}
})
