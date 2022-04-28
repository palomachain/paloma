package valset

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgRegisterConductor = "op_weight_msg_register_conductor"
	// TODO: Determine the simulation weight value
	defaultWeightMsgRegisterConductor int = 100

	opWeightMsgAddExternalChainInfoForValidator = "op_weight_msg_add_external_chain_info_for_validator"
	// TODO: Determine the simulation weight value
	defaultWeightMsgAddExternalChainInfoForValidator int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	valsetGenesis := types.GenesisState{
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&valsetGenesis)
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

	var weightMsgRegisterConductor int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgRegisterConductor, &weightMsgRegisterConductor, nil,
		func(_ *rand.Rand) {
			weightMsgRegisterConductor = defaultWeightMsgRegisterConductor
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRegisterConductor,
		valsetsimulation.SimulateMsgRegisterConductor(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgAddExternalChainInfoForValidator int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgAddExternalChainInfoForValidator, &weightMsgAddExternalChainInfoForValidator, nil,
		func(_ *rand.Rand) {
			weightMsgAddExternalChainInfoForValidator = defaultWeightMsgAddExternalChainInfoForValidator
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgAddExternalChainInfoForValidator,
		valsetsimulation.SimulateMsgAddExternalChainInfoForValidator(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
