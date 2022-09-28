package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/vizualni/whoops"
)

const collectFundsBlockRange = 100

func (k Keeper) CollectJobFundEvents(ctx sdk.Context) error {
	return whoops.Try(func() {
		var g whoops.Group
		for _, ci := range whoops.Must(k.GetAllChainInfos(ctx)) {
			g.Add(
				k.ConsensusKeeper.PutMessageInQueue(
					ctx,
					consensustypes.Queue(
						ConsensusCollectFundEvents,
						xchainType,
						ci.GetChainReferenceID(),
					),
					&types.CollectFunds{
						ChainReferenceID: ci.GetChainReferenceID(),
						FromBlockTime:    ci.GetBlockHeight(),
						ToBlockTime:      ci.GetBlockHeight(),
					},
					nil,
				),
			)
		}
		whoops.Assert(g.Return())
	})
}

func (k Keeper) attestCollectedFunds(ctx sdk.Context) error {

}
