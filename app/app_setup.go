package app

import (
	"encoding/json"
	"os"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/version"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type TestApp struct {
	App
}

type testing interface {
	TempDir() string
	Cleanup(func())
}

func NewTestApp(t testing, isCheckTx bool) TestApp {
	db := dbm.NewMemDB()
	encCfg := MakeEncodingConfig()

	oldVersion := version.Version
	version.Version = "v5.1.6"
	t.Cleanup(func() {
		version.Version = oldVersion
	})
	app := New(
		log.NewTMJSONLogger(os.Stdout),
		db,
		nil,
		true,
		map[int64]bool{},
		t.TempDir(),
		5,
		encCfg,
		simapp.EmptyAppOptions{},
	)

	if !isCheckTx {
		genesisState := NewDefaultGenesisState(encCfg.Codec)
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")

		if err != nil {
			panic(err)
		}

		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simapp.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return TestApp{*app.(*App)}
}
