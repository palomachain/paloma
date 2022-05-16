package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"

	"github.com/vizualni/whoops"
)

type AttestSeeing struct {
	ID           uint64
	PipelineName string
	Data         []byte
}

func (k Keeper) CheckAndProcessAttestedMessages(ctx sdk.Context) error {

	var gerr whoops.Group
	// TODO
	if ctx.BlockHeight()%5 != 0 {
		return nil
	}

mainLoop:
	for _, att := range k.attestator.registry {
		msgs, err := k.GetMessagesThatHaveReachedConsensus(ctx, att.ConsensusQueue())
		if err != nil {
			gerr.Add(err)
			continue
		}

		for _, msg := range msgs {
			origMsg, err := msg.SdkMsg()
			if err != nil {
				gerr.Add(err)
				continue mainLoop
			}

			task, ok := origMsg.(types.AttestTask)
			if !ok {
				panic("what now")
			}

			evidence := []types.Evidence{}
			for _, sd := range msg.GetSignData() {
				evidence = append(evidence, types.Evidence{
					From: whoops.Must(sdk.ValAddressFromBech32(sd.ValAddress)),
					Data: sd.ExtraData,
				})
			}

			res, err := att.ProcessAllEvidence(ctx.Context(), task, evidence)
			if err != nil {
				gerr.Add(err)
				continue mainLoop
			}

			// TODO: process result of processing evidence
			_ = res

			cq, err := k.getConsensusQueue(att.ConsensusQueue())
			if err != nil {
				gerr.Add(err)
				continue mainLoop
			}
			err = cq.remove(ctx, msg.GetId())
			if err != nil {
				gerr.Add(err)
				continue mainLoop
			}
		}
	}
	if gerr.Err() {
		return gerr
	}
	return nil
}

func NewAttestator() *Attestator {
	return &Attestator{
		registry: make(map[types.ConsensusQueueType]types.Attestator),
	}
}

type Attestator struct {
	registry        map[types.ConsensusQueueType]types.Attestator
	ConsensusKeeper Keeper
}

func (a *Attestator) RegisterProcessor(att types.Attestator) {
	AddConcencusQueueType(&a.ConsensusKeeper, att.ConsensusQueue())
	a.registry[att.ConsensusQueue()] = att
}

func (a *Attestator) validateIncoming(ctx context.Context, queueTypeName types.ConsensusQueueType, task types.AttestTask, evidence types.Evidence) error {
	if att, ok := a.registry[queueTypeName]; ok {
		return att.ValidateEvidence(ctx, task, evidence)
	}

	return nil
}
