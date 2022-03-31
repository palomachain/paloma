package client

import (
	"io"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	lens "github.com/strangelove-ventures/lens/client"
	prov "github.com/tendermint/tendermint/light/provider/http"
)

type LensClient struct {
	lens.ChainClient
}

func (cc *LensClient) Init() error {
	// TODO: test key directory and return error if not created
	keybase, err := keyring.New(cc.Config.ChainID, cc.Config.KeyringBackend, cc.Config.KeyDirectory, cc.Input, cc.KeyringOptions...)
	if err != nil {
		return err
	}
	// TODO: figure out how to deal with input or maybe just make all keyring backends test?

	timeout, _ := time.ParseDuration(cc.Config.Timeout)
	rpcClient, err := lens.NewRPCClient(cc.Config.RPCAddr, timeout)
	if err != nil {
		return err
	}

	lightprovider, err := prov.New(cc.Config.ChainID, cc.Config.RPCAddr)
	if err != nil {
		return err
	}

	cc.RPCClient = rpcClient
	cc.LightProvider = lightprovider
	cc.Keybase = keybase

	return nil
}

func NewChainClient(ccc *lens.ChainClientConfig, homepath string, input io.Reader, output io.Writer, kro ...keyring.Option) (*LensClient, error) {
	cc := LensClient{lens.ChainClient{
		KeyringOptions: kro,
		Config:         ccc,
		Input:          input,
		Output:         output,
		Codec:          lens.MakeCodec(ccc.Modules),
	}}
	if err := cc.Init(); err != nil {
		return nil, err
	}
	return &cc, nil
}
