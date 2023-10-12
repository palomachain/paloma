package keeper

import (
	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k Keeper) CollectJobFundEvents(ctx sdk.Context) error {
	return whoops.Try(func() {
		var g whoops.Group
		for _, ci := range whoops.Must(k.GetAllChainInfos(ctx)) {
			_, err := k.ConsensusKeeper.PutMessageInQueue(
				ctx,
				consensustypes.Queue(
					ConsensusCollectFundEvents,
					xchainType,
					ci.GetChainReferenceID(),
				),
				&types.CollectFunds{
					// ChainReferenceID: ci.GetChainReferenceID(),
					// FromBlockTime:    ci.GetBlockHeight(),
					// ToBlockTime:      ci.GetBlockHeight(),
				},
				nil,
			)
			g.Add(err)
		}

		whoops.Assert(g.Return())
	})
}

func (k Keeper) attestCollectedFunds(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) (retErr error) {
	return nil
}
