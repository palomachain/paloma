package consensus

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
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
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgAddMessagesSignatures = "op_weight_msg_create_chain"
	// TODO: Determine the simulation weight value
	defaultWeightMsgAddMessagesSignatures int = 100

	opWeightMsgAddEvidence = "op_weight_msg_add_evidence"
	// TODO: Determine the simulation weight value
	defaultWeightMsgAddEvidence int = 100

	opWeightMsgSetPublicAccessData = "op_weight_msg_set_public_access_data"
	// TODO: Determine the simulation weight value
	defaultWeightMsgSetPublicAccessData int = 100
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	consensusGenesis := types.GenesisState{}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&consensusGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgAddMessagesSignatures int
	simState.AppParams.GetOrGenerate(opWeightMsgAddMessagesSignatures, &weightMsgAddMessagesSignatures, nil,
		func(_ *rand.Rand) {
			weightMsgAddMessagesSignatures = defaultWeightMsgAddMessagesSignatures
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgAddMessagesSignatures,
		consensussimulation.SimulateMsgAddMessagesSignatures(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgAddEvidence int
	simState.AppParams.GetOrGenerate(opWeightMsgAddEvidence, &weightMsgAddEvidence, nil,
		func(_ *rand.Rand) {
			weightMsgAddEvidence = defaultWeightMsgAddEvidence
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgAddEvidence,
		consensussimulation.SimulateMsgAddEvidence(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgSetPublicAccessData int
	simState.AppParams.GetOrGenerate(opWeightMsgSetPublicAccessData, &weightMsgSetPublicAccessData, nil,
		func(_ *rand.Rand) {
			weightMsgSetPublicAccessData = defaultWeightMsgSetPublicAccessData
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSetPublicAccessData,
		consensussimulation.SimulateMsgSetPublicAccessData(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	return operations
}
