package valset

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/palomachain/paloma/testutil/sample"
	valsetsimulation "github.com/palomachain/paloma/x/valset/simulation"
	"github.com/palomachain/paloma/x/valset/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = valsetsimulation.FindAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgAddExternalChainInfoForValidator = "op_weight_msg_add_external_chain_info_for_validator"
	// TODO: Determine the simulation weight value
	defaultWeightMsgAddExternalChainInfoForValidator int = 100
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	valsetGenesis := types.GenesisState{}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&valsetGenesis)
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

	var weightMsgAddExternalChainInfoForValidator int
	simState.AppParams.GetOrGenerate(opWeightMsgAddExternalChainInfoForValidator, &weightMsgAddExternalChainInfoForValidator, simState.Rand,
		func(_ *rand.Rand) {
			weightMsgAddExternalChainInfoForValidator = defaultWeightMsgAddExternalChainInfoForValidator
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgAddExternalChainInfoForValidator,
		valsetsimulation.SimulateMsgAddExternalChainInfoForValidator(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	return operations
}
