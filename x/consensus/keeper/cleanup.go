package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/util/libmsg"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/consensus/types"
	evm "github.com/palomachain/paloma/x/evm/types"
)

func (k Keeper) PruneOldMessages(ctx sdk.Context, blocksAgo int64) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, supportedConsensusQueue := range k.registry.slice {
		opts, err := supportedConsensusQueue.SupportedQueues(sdkCtx)
		if err != nil {
			return err
		}
		for _, opt := range opts {
			msgs, err := k.getMessagesOlderThan(ctx, opt.QueueTypeName, blocksAgo)
			if err != nil {
				return err
			}

			for _, msg := range msgs {
				err = k.PruneJob(ctx, opt.QueueTypeName, msg.GetId())
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (k Keeper) ReassignOrphanedMessages(ctx sdk.Context, blockAge int64) (err error) {
	logger := liblog.FromSDKLogger(k.Logger(ctx)).WithFields("component", "message-reassigner")
	logger.Debug("Message reassigning loop triggered.")

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered panic: %v, original error if any: %w", r, err)
			logger.WithError(err).Error("Recovered panic.")
		}
	}()

	for _, supportedConsensusQueue := range k.registry.slice {
		opts, err := supportedConsensusQueue.SupportedQueues(ctx)
		if err != nil {
			return err
		}

		for _, opt := range opts {
			logger = logger.WithFields("chain-reference-id", opt.ChainReferenceID, "queueType", opt.QueueTypeName)
			msgs, err := k.getMessagesOlderThan(ctx, opt.QueueTypeName, blockAge)
			if err != nil {
				return err
			}

			logger = logger.WithFields("num-stale-messages", len(msgs))
			logger.Debug("Retrieved messages.")

			for _, msg := range msgs {
				logger = logger.WithFields("msg-id", msg.GetId())
				if msg.GetPublicAccessData() != nil || msg.GetErrorData() != nil {
					logger.Debug("Skipping message with set data fields")
					continue
				}

				currentVal, err := libmsg.GetAssignee(msg, k.cdc)
				if err != nil || len(currentVal) < 1 {
					// Only reassign if message has an active assignee
					// Some messages like balance validations do not.
					logger.Debug("Skipping message without assignee.")
					continue
				}

				evmmsg, err := libmsg.ToEvmMessage(msg, k.cdc)
				if err != nil {
					logger.WithError(err).Error("failed to parse to evm message.")
					return fmt.Errorf("failed to parse to evm message: %w", err)
				}
				req := deriveMessageRequirements(evmmsg, k.cdc)
				newVal, err := k.evmKeeper.PickValidatorForMessage(ctx, evmmsg.ChainReferenceID, req)
				if err != nil {
					logger.WithError(err).Error("failed to pick new validator for message.")
					return fmt.Errorf("failed to pick validator for message: %w", err)
				}

				logger = logger.WithFields("current-validator", currentVal, "new-validator", newVal)
				logger.Debug("Attepting to reassign message...")
				if err := k.reassignMessageValidator(ctx, newVal, msg.GetId(), opt.QueueTypeName); err != nil {
					logger.WithError(err).Error("failed to reassign new validator to message.")
					return err
				}

				logger.Debug("Message reassigned.")
				keeperutil.EmitEvent(k, ctx, "OrphanedMessagesReassigner",
					sdk.NewAttribute("msg-id", strconv.FormatUint(msg.GetId(), 10)),
					sdk.NewAttribute("current-relayer", currentVal),
					sdk.NewAttribute("new-relayer", newVal),
				)
			}
		}
	}
	return nil
}

func (k Keeper) getMessagesOlderThan(ctx sdk.Context, queueType string, blockAge int64) ([]types.QueuedSignedMessageI, error) {
	msgs, err := k.GetMessagesFromQueue(ctx, queueType, 9999)
	if err != nil {
		return nil, err
	}

	bh := ctx.BlockHeight()
	return slice.Filter(msgs, func(val types.QueuedSignedMessageI) bool {
		return bh-val.GetAddedAtBlockHeight() > blockAge
	}), nil
}

func deriveMessageRequirements(msg *evm.Message, cdc types.AnyUnpacker) *xchain.JobRequirements {
	actionMsg := msg.GetAction()
	// TODO: Move type check
	switch origMsg := actionMsg.(type) {
	case *evm.Message_SubmitLogicCall:
		if origMsg.SubmitLogicCall.ExecutionRequirements.EnforceMEVRelay {
			return &xchain.JobRequirements{
				EnforceMEVRelay: true,
			}
		}
	}

	return nil
}
