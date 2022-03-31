package client

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/strangelove-ventures/lens/client"
)

func newClient() *LensClient {
	return &LensClient{
		ChainClient: client.ChainClient{
			Config: &client.ChainClientConfig{
				ChainID: "random-chain",
			},
			Codec: client.MakeCodec(testModules()),
		},
	}
}

func testModules() []module.AppModuleBasic {
	return append(
		client.ModuleBasics[:],
		// byop.NewModule(
		// "testModule",
		// (&types.SimpleMessage)(nil),
		// ),
	)
}
