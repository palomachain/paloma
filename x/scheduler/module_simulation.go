package scheduler

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/palomachain/paloma/testutil/sample"
	schedulersimulation "github.com/palomachain/paloma/x/scheduler/simulation"
	"github.com/palomachain/paloma/x/scheduler/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = schedulersimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	//nolint:gosec
	opWeightMsgSubmitRecurringJob = "op_weight_msg_create_chain"
	// TODO: Determine the simulation weight value
	defaultWeightMsgSubmitRecurringJob int = 100

	//nolint:gosec
	opWeightMsgHello = "op_weight_msg_create_chain"
	// TODO: Determine the simulation weight value
	defaultWeightMsgHello int = 100

	opWeightMsgPauseRecurringJob = "op_weight_msg_create_chain"
	// TODO: Determine the simulation weight value
	defaultWeightMsgPauseRecurringJob int = 100

	opWeightMsgResumeRecurringJob = "op_weight_msg_create_chain"
	// TODO: Determine the simulation weight value
	defaultWeightMsgResumeRecurringJob int = 100

	opWeightMsgSigningQueueMessage = "op_weight_msg_create_chain"
	// TODO: Determine the simulation weight value
	defaultWeightMsgSigningQueueMessage int = 100

	opWeightMsgCreateJob = "op_weight_msg_create_job"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateJob int = 100

	opWeightMsgExecuteJob = "op_weight_msg_execute_job"
	// TODO: Determine the simulation weight value
	defaultWeightMsgExecuteJob int = 100
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	schedulerGenesis := types.GenesisState{}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&schedulerGenesis)
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

	var weightMsgSubmitRecurringJob int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgSubmitRecurringJob, &weightMsgSubmitRecurringJob, nil,
		func(_ *rand.Rand) {
			weightMsgSubmitRecurringJob = defaultWeightMsgSubmitRecurringJob
		},
	)

	var weightMsgHello int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgHello, &weightMsgHello, nil,
		func(_ *rand.Rand) {
			weightMsgHello = defaultWeightMsgHello
		},
	)

	var weightMsgPauseRecurringJob int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgPauseRecurringJob, &weightMsgPauseRecurringJob, nil,
		func(_ *rand.Rand) {
			weightMsgPauseRecurringJob = defaultWeightMsgPauseRecurringJob
		},
	)

	var weightMsgResumeRecurringJob int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgResumeRecurringJob, &weightMsgResumeRecurringJob, nil,
		func(_ *rand.Rand) {
			weightMsgResumeRecurringJob = defaultWeightMsgResumeRecurringJob
		},
	)

	var weightMsgSigningQueueMessage int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgSigningQueueMessage, &weightMsgSigningQueueMessage, nil,
		func(_ *rand.Rand) {
			weightMsgSigningQueueMessage = defaultWeightMsgSigningQueueMessage
		},
	)

	var weightMsgCreateJob int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCreateJob, &weightMsgCreateJob, nil,
		func(_ *rand.Rand) {
			weightMsgCreateJob = defaultWeightMsgCreateJob
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateJob,
		schedulersimulation.SimulateMsgCreateJob(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgExecuteJob int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgExecuteJob, &weightMsgExecuteJob, nil,
		func(_ *rand.Rand) {
			weightMsgExecuteJob = defaultWeightMsgExecuteJob
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgExecuteJob,
		schedulersimulation.SimulateMsgExecuteJob(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	return operations
}
