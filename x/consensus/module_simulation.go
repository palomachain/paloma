package consensus

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/palomachain/paloma/testutil/sample"
	consensussimulation "github.com/palomachain/paloma/x/consensus/simulation"
	"github.com/palomachain/paloma/x/consensus/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = consensussimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgAddMessagesSignatures = "op_weight_msg_create_chain"
	// TODO: Determine the simulation weight value
	defaultWeightMsgAddMessagesSignatures int = 100

	opWeightMsgDeleteJob = "op_weight_msg_delete_job"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteJob int = 100

	opWeightMsgAddEvidence = "op_weight_msg_add_evidence"
	// TODO: Determine the simulation weight value
	defaultWeightMsgAddEvidence int = 100

	opWeightMsgSetPublicAccessData = "op_weight_msg_set_public_access_data"
	// TODO: Determine the simulation weight value
	defaultWeightMsgSetPublicAccessData int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	consensusGenesis := types.GenesisState{
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&consensusGenesis)
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

	var weightMsgAddMessagesSignatures int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgAddMessagesSignatures, &weightMsgAddMessagesSignatures, nil,
		func(_ *rand.Rand) {
			weightMsgAddMessagesSignatures = defaultWeightMsgAddMessagesSignatures
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgAddMessagesSignatures,
		consensussimulation.SimulateMsgAddMessagesSignatures(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeleteJob int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDeleteJob, &weightMsgDeleteJob, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteJob = defaultWeightMsgDeleteJob
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteJob,
		consensussimulation.SimulateMsgDeleteJob(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgAddEvidence int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgAddEvidence, &weightMsgAddEvidence, nil,
		func(_ *rand.Rand) {
			weightMsgAddEvidence = defaultWeightMsgAddEvidence
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgAddEvidence,
		consensussimulation.SimulateMsgAddEvidence(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgSetPublicAccessData int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgSetPublicAccessData, &weightMsgSetPublicAccessData, nil,
		func(_ *rand.Rand) {
			weightMsgSetPublicAccessData = defaultWeightMsgSetPublicAccessData
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSetPublicAccessData,
		consensussimulation.SimulateMsgSetPublicAccessData(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
