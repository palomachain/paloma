package app

import (
	"encoding/json"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/version"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

type TestApp struct {
	App
}

type testing interface {
	TempDir() string
	Cleanup(func())
}

// DefaultConsensusParams defines default Tendermint consensus parameters used
// for testing purposes.
var DefaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
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
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return TestApp{*app}
}
