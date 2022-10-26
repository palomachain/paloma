package evm

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/palomachain/paloma/testutil/sample"
	evmsimulation "github.com/palomachain/paloma/x/evm/simulation"
	"github.com/palomachain/paloma/x/evm/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = evmsimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgSubmitNewJob = "op_weight_msg_submit_new_job"
	// TODO: Determine the simulation weight value
	defaultWeightMsgSubmitNewJob int = 100
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	evmGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&evmGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {

	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgSubmitNewJob int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgSubmitNewJob, &weightMsgSubmitNewJob, nil,
		func(_ *rand.Rand) {
			weightMsgSubmitNewJob = defaultWeightMsgSubmitNewJob
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSubmitNewJob,
		evmsimulation.SimulateMsgSubmitNewJob(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	return operations
}
