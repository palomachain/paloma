package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CheckAndProcessAttestedMessages is supposed to be used within the
// EndBlocker. It will get messages for the attestators that have reached a
// consensus and process them.
func (k Keeper) CheckAndProcessAttestedMessages(ctx sdk.Context) error {
	for _, supportedConsensusQueue := range k.registry.slice {
		opts, err := supportedConsensusQueue.SupportedQueues(ctx)
		if err != nil {
			return err
		}
		for _, opt := range opts {
			msgs, err := k.GetMessagesFromQueue(ctx, opt.QueueTypeName, 9999)
			if err != nil {
				return err
			}

			for _, msg := range msgs {
				k.Logger(ctx).Debug(
					"check-and-process-attested-messages-queue",
					"id", msg.GetId(),
					"nonce", msg.Nonce(),
				)
				cq, err := k.getConsensusQueue(ctx, opt.QueueTypeName)
				if err != nil {
					return err
				}
				if err := opt.ProcessMessageForAttestation(ctx, cq, msg); err != nil {
					return err
				}
			}
		}

	}
	return nil
}
