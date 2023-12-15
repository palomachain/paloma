package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) DeleteOldMessages(ctx context.Context, blocksAgo int64) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, supportedConsensusQueue := range k.registry.slice {
		opts, err := supportedConsensusQueue.SupportedQueues(sdkCtx)
		if err != nil {
			return err
		}
		for _, opt := range opts {
			msgs, err := k.GetMessagesFromQueue(sdkCtx, opt.QueueTypeName, 9999)
			if err != nil {
				return err
			}

			for _, msg := range msgs {
				if sdkCtx.BlockHeight()-msg.GetAddedAtBlockHeight() > blocksAgo {
					err = k.DeleteJob(sdkCtx, opt.QueueTypeName, msg.GetId())
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
