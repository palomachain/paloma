package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	"github.com/palomachain/paloma/x/consensus/types"

	"github.com/vizualni/whoops"
)

// CheckAndProcessAttestedMessages is supposed to be used within the
// EndBlocker. It will get messages for the attestators that have reached a
// consensus and process them.
func (k Keeper) CheckAndProcessAttestedMessages(ctx sdk.Context) error {
	var gerr whoops.Group
	// TODO
	if ctx.BlockHeight()%5 != 0 {
		return nil
	}

mainLoop:
	for _, att := range k.attestator.registry {
		chainType, chainID := att.ChainInfo()
		queue := types.Queue(att.ConsensusQueue(), chainType, chainID)
		msgs, err := k.GetMessagesThatHaveReachedConsensus(ctx, queue)
		if err != nil {
			gerr.Add(err)
			continue
		}

		for _, msg := range msgs {
			origMsg, err := msg.ConsensusMsg(k.cdc)
			if err != nil {
				gerr.Add(err)
				continue mainLoop
			}

			task, ok := origMsg.(types.AttestTask)
			if !ok {
				// TODO:
				panic("what now")
			}

			evidence := []types.Evidence{}
			for _, sd := range msg.GetSignData() {
				evidence = append(evidence, types.Evidence{
					From: sd.ValAddress,
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

			cq, err := k.getConsensusQueue(queue)
			if err != nil {
				gerr.Add(err)
				continue mainLoop
			}
			err = cq.Remove(ctx, msg.GetId())
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
		registry: make(map[string]types.Attestator),
	}
}

type Attestator struct {
	registry        map[string]types.Attestator
	ConsensusKeeper *Keeper
}

func (a *Attestator) RegisterAttestator(att types.Attestator) {
	chainType, chainID := att.ChainInfo()
	a.ConsensusKeeper.AddConcencusQueueType(
		false,
		consensus.WithQueueTypeName(att.ConsensusQueue()),
		consensus.WithStaticTypeCheck(att.Type()),
		consensus.WithBytesToSignCalc(att.BytesToSign()),
		consensus.WithVerifySignature(att.VerifySignature()),
		consensus.WithChainInfo(chainType, chainID),
	)
	a.registry[types.Queue(att.ConsensusQueue(), chainType, chainID)] = att
}

func (a *Attestator) validateIncoming(ctx context.Context, queueTypeName string, task types.AttestTask, evidence types.Evidence) error {
	if att, ok := a.registry[queueTypeName]; ok {
		return att.ValidateEvidence(ctx, task, evidence)
	}

	return nil
}
