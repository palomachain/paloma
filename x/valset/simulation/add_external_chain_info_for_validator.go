package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/palomachain/paloma/x/valset/keeper"
	"github.com/palomachain/paloma/x/valset/types"
)

func SimulateMsgAddExternalChainInfoForValidator(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainReferenceID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgAddExternalChainInfoForValidator{
			Metadata: types.MsgMetadata{
				Creator: simAccount.Address.String(),
			},
		}

		// TODO: Handling the AddExternalChainInfoForValidator simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "AddExternalChainInfoForValidator simulation not implemented"), nil, nil
	}
}
