package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) DeleteOldMessages(ctx sdk.Context, blocksAgo int64) error {
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
				if ctx.BlockHeight()-msg.GetAddedAtBlockHeight() > blocksAgo {
					err = k.DeleteJob(ctx, opt.QueueTypeName, msg.GetId())
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
