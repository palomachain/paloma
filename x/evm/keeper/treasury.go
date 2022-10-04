package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/vizualni/whoops"
)

const collectFundsMinutesRange = 10 * time.Minute

func (k Keeper) CollectJobFundEvents(ctx sdk.Context) error {
	return whoops.Try(func() {
		var g whoops.Group
		for _, ci := range whoops.Must(k.GetAllChainInfos(ctx)) {
			tt := k.calculatNextTimeForCollectingFunds(ctx, xchain.ReferenceID(ci.GetChainReferenceID()))
			g.Add(
				k.ConsensusKeeper.PutMessageInQueue(
					ctx,
					consensustypes.Queue(
						ConsensusCollectFundEvents,
						xchainType,
						ci.GetChainReferenceID(),
					),
					&types.CollectFunds{
						FromBlockTime: tt,
						ToBlockTime:   ctx.BlockTime(),
					},
					nil,
				),
			)
		}
		whoops.Assert(g.Return())
	})
}

func (k Keeper) calculatNextTimeForCollectingFunds(ctx sdk.Context, refID xchain.ReferenceID) time.Time {
	s := k.collectFundEventsBlockHeightStore(ctx)
	key := []byte(refID)
	if !s.Has(key) {
		return time.Date(2022, 9, 30, 0, 0, 0, 0, time.UTC)
	}
	val := s.Get(key)
	var t time.Time
	err := (&t).UnmarshalBinary(val)
	if err != nil {
		panic(err)
	}

	return t
}

func (k Keeper) collectFundEventsBlockHeightStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("fund-events"))
}

func (k Keeper) attestCollectedFunds(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) (retErr error) {
	return nil
}
