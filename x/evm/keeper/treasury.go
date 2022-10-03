package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/vizualni/whoops"
)

const collectFundsBlockRange = 5000

func (k Keeper) CollectJobFundEvents(ctx sdk.Context) error {
	return whoops.Try(func() {
		var g whoops.Group
		for _, ci := range whoops.Must(k.GetAllChainInfos(ctx)) {
			height := k.calculatNextHeightForCollectingFunds(ctx, xchain.ReferenceID(ci.GetChainReferenceID()))
			g.Add(
				k.ConsensusKeeper.PutMessageInQueue(
					ctx,
					consensustypes.Queue(
						ConsensusCollectFundEvents,
						xchainType,
						ci.GetChainReferenceID(),
					),
					&types.CollectFunds{
						FromBlockHeight: height,
						ToBlockHeight:   height + collectFundsBlockRange,
					},
					nil,
				),
			)
		}
		whoops.Assert(g.Return())
	})
}

func (k Keeper) calculatNextHeightForCollectingFunds(ctx sdk.Context, refID xchain.ReferenceID) uint64 {
	s := k.collectFundEventsBlockHeightStore(ctx)
	key := []byte(refID)
	var height uint64
	if !s.Has(key) {
		ci, err := k.GetChainInfo(ctx, refID)
		if err != nil {
			panic(err)
		}
		height = ci.GetReferenceBlockHeight()
		s.Set(key, keeperutil.Uint64ToByte(height))
	} else {
		val := s.Get(key)
		height = keeperutil.BytesToUint64(val) + 1
	}

	return height
}

func (k Keeper) collectFundEventsBlockHeightStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("fund-events"))
}

func (k Keeper) attestCollectedFunds(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) (retErr error) {
	return nil
}
