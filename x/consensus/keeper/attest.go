package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/liblog"
)

// CheckAndProcessAttestedMessages is supposed to be used within the
// EndBlocker. It will get messages for the attestators that have reached a
// consensus and process them.
func (k Keeper) CheckAndProcessAttestedMessages(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, supportedConsensusQueue := range k.registry.slice {
		opts, err := supportedConsensusQueue.SupportedQueues(sdkCtx)
		if err != nil {
			return err
		}
		for _, opt := range opts {
			msgs, err := k.GetMessagesFromQueue(sdkCtx, opt.QueueTypeName, 0)
			if err != nil {
				return err
			}

			for _, msg := range msgs {
				liblog.FromSDKLogger(k.Logger(sdkCtx)).WithFields("id", msg.GetId(), "nonce", msg.Nonce()).Debug(
					"check-and-process-attested-messages-queue.")
				cq, err := k.getConsensusQueue(sdkCtx, opt.QueueTypeName)
				if err != nil {
					return err
				}
				if err := opt.ProcessMessageForAttestation(sdkCtx, cq, msg); err != nil {
					return err
				}
			}
		}

	}
	return nil
}
