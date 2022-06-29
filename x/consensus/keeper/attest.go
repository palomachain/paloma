package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	for _, supportedConsensusQueue := range *k.registry.slice {
		atts, err := supportedConsensusQueue.SupportedQueues(ctx)
		if err != nil {
			return err
		}
	loopAttestators:
		for _, att := range atts {
			if att.Attestator == nil {
				continue
			}
			msgs, err := k.GetMessagesThatHaveReachedConsensus(ctx, att.QueueTypeName)
			if err != nil {
				gerr.Add(err)
				continue
			}

			for _, msg := range msgs {
				origMsg, err := msg.ConsensusMsg(k.cdc)
				if err != nil {
					gerr.Add(err)
					continue loopAttestators
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

				res, err := att.Attestator.ProcessAllEvidence(ctx, task, evidence)
				if err != nil {
					gerr.Add(err)
					continue loopAttestators
				}

				// TODO: process result of proc
				_ = res

				cq, err := k.getConsensusQueue(ctx, att.QueueTypeName)
				if err != nil {
					gerr.Add(err)
					continue loopAttestators
				}
				err = cq.Remove(ctx, msg.GetId())
				if err != nil {
					gerr.Add(err)
					continue loopAttestators
				}
			}
		}

	}
	if gerr.Err() {
		return gerr
	}
	return nil
}
