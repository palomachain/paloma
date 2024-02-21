package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/palomachain/paloma/x/scheduler/keeper"
	"github.com/palomachain/paloma/x/scheduler/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

func SimulateMsgExecuteJob(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgExecuteJob{
			Metadata: valsettypes.MsgMetadata{
				Creator: simAccount.Address.String(),
			},
		}

		// TODO: Handling the ExecuteJob simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "ExecuteJob simulation not implemented"), nil, nil
	}
}
