package app

import (
	"encoding/json"
	"os"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/version"
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
var DefaultConsensusParams = &tmproto.ConsensusParams{
	Block: &tmproto.BlockParams{
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

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = t.TempDir()
	appOptions[server.FlagInvCheckPeriod] = 5

	app := New(
		log.NewTMJSONLogger(os.Stdout),
		db,
		nil,
		true,
		encCfg,
		appOptions,
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
