package keeper

import (
	"errors"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/vizualni/whoops"
)

type Distance int

const (
	RightNear Distance = iota + 1
	LeftNear  Distance
	RightFar  Distance
	LeftFar   Distance
)


type RangedEvent[T any] interface {
	Merge(T) (T, bool)
	Less(T) bool
}

func SimplifyRangeEvents[T any](events []RangedEvent[T]) []any {
	slice := events[:]
	sort.Slice(slice, func(i, j int) bool { return events[i].Less[events[j]] })
	events = slice

	newEvents := make([]any)
	newEvents = append(newEvents, events[i])
	for i:=1; i<len(events); i++ {
		evt := newEvents[len(newEvents)-1]	

		newEvt, ok := evt.Merge(events[i])
		if !ok {
			newEvents = append(newEvents, events[i])
			continue
		}

		newEvents[len(newEvents) - 1]=newEvt
	}

	return newEvents
}

const collectFundsMinutesRange = 10 * time.Minute

func (k Keeper) CollectJobFundEvents(ctx sdk.Context) error {
	return whoops.Try(func() {
		var g whoops.Group
		for _, ci := range whoops.Must(k.GetAllChainInfos(ctx)) {
			tt := k.calculatNextTimeForCollectingFunds(ctx, xchain.ReferenceID(ci.GetChainReferenceID()))

			// check if the same

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
	if len(msg.GetEvidence()) == 0 {
		return nil
	}

	ctx, writeCache := ctx.CacheContext()
	defer func() {
		if retErr == nil {
			writeCache()
		}
	}()

	consensusMsg, err := msg.ConsensusMsg(k.cdc)
	if err != nil {
		return err
	}

	request := consensusMsg.(*types.CollectFunds)

	evidence, err := k.findEvidenceThatWon(ctx, msg.GetEvidence())
	if err != nil {
		if errors.Is(err, ErrConsensusNotAchieved) {
			return nil
		}
		return err
	}

	defer func() {
		// given that there was enough evidence for a proof, regardless of the outcome,
		// we should remove this from the queue as there isn't much that we can do about it.
		q.Remove(ctx, msg.GetId())
	}()

	_, chainReferenceID := q.ChainInfo()

	return k.processFundCollectedEvidence(ctx, request, chainReferenceID, evidence)
}

func (k Keeper) processFundCollectedEvidence(
	ctx sdk.Context,
	request *types.CollectFunds,
	refID xchain.ReferenceID,
	anyEvidence any,
) error {
	switch evidence := anyEvidence.(type) {
	case *types.FundCollectedEvidence:
		var g whoops.Group

		for _, e := range evidence.GetEvidence() {
			g.Add(k.Treasury.AddFunds(ctx, chainType, refID, e.GetJobID()))
		}

		return g.Return()
	}

	return nil

}
