package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libcons"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/util/libmsg"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	"github.com/palomachain/paloma/x/consensus/types"
	evmtypes "github.com/palomachain/paloma/x/evm/types"
)

// CheckAndProcessEstimatedMessages is supposed to be used within the
// EndBlocker. It will get messages which have received gas estimates
// and try to process them.
func (k Keeper) CheckAndProcessEstimatedMessages(ctx context.Context) (err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := liblog.FromKeeper(ctx, k).WithComponent("check-and-process-estimated-messages")
	for _, supportedConsensusQueue := range k.registry.slice {
		opts, err := supportedConsensusQueue.SupportedQueues(sdkCtx)
		if err != nil {
			logger.WithError(err).Warn("Failed to get supported queues")
			continue
		}
		for _, opt := range opts {
			msgs, err := k.GetMessagesFromQueue(sdkCtx, opt.QueueTypeName, 0)
			if err != nil {
				logger.WithError(err).Warn("Failed to get messages from queue")
				continue
			}

			cq, err := k.getConsensusQueue(sdkCtx, opt.QueueTypeName)
			if err != nil {
				logger.WithChain(opt.ChainReferenceID).
					WithError(err).
					Warn("Failed to get consensus queue")
				continue
			}
			for _, msg := range msgs {
				cachedCtx, commit := sdk.UnwrapSDKContext(ctx).CacheContext()
				if err := k.checkAndProcessEstimatedMessage(cachedCtx, msg, cq); err != nil {
					logger.WithFields("msg_id", msg.GetId()).
						WithError(err).
						Warn("Failed to process estimated message")
					continue
				}
				commit()
			}
		}
	}
	return nil
}

func (k Keeper) checkAndProcessEstimatedMessage(ctx context.Context,
	msg types.QueuedSignedMessageI,
	q consensus.Queuer,
) error {
	_, rcid := q.ChainInfo()
	// Skip messages that don't require gas estimation
	if !msg.GetRequireGasEstimation() {
		return nil
	}

	// Skip messages that don't have gas estimates
	if len(msg.GetGasEstimates()) < 1 {
		return nil
	}

	// Skip messages that have gas estimate
	if msg.GetGasEstimate() > 0 {
		return nil
	}
	logger := liblog.FromKeeper(ctx, k).
		WithComponent("check-and-process-estimated-messages").
		WithChain(rcid).
		WithFields(
			"id", msg.GetId(),
			"nonce", msg.Nonce())
	logger.Debug("Processing gas estimates for message.")

	estimate, err := k.consensusChecker.VerifyGasEstimates(ctx, k,
		slice.Map(msg.GetGasEstimates(), func(ge *types.GasEstimate) libcons.GasEstimate {
			return ge
		}))
	if err != nil {
		if errors.Is(err, libcons.ErrConsensusNotAchieved) {
			logger.WithError(err).Debug("Consensus not achieved for gas estimates")
			return nil
		}
		return fmt.Errorf("failed to verify gas estimates: %w", err)
	}

	if err := q.SetElectedGasEstimate(ctx, msg.GetId(), estimate); err != nil {
		return fmt.Errorf("failed to set elected gas estimate: %w", err)
	}

	if err := k.checkAndProcessEstimatedFeePayer(ctx, msg, q, estimate); err != nil {
		return fmt.Errorf("failed to process estimated submit logic call: %w", err)
	}

	return nil
}

func (k Keeper) checkAndProcessEstimatedFeePayer(
	ctx context.Context,
	msg types.QueuedSignedMessageI,
	q consensus.Queuer,
	estimate uint64,
) error {
	m, err := libmsg.ToEvmMessage(msg, k.cdc)
	if err != nil {
		return fmt.Errorf("failed to convert message to evm message: %w", err)
	}
	action, ok := m.Action.(evmtypes.FeePayer)
	if !ok {
		// Skip messages that do not contain fees
		return nil
	}

	valAddr, err := sdk.ValAddressFromBech32(m.GetAssignee())
	if err != nil {
		return fmt.Errorf("failed to parse validator address: %w", err)
	}

	fees, err := k.calculateFeesForEstimate(ctx, valAddr, m.GetChainReferenceID(), estimate)
	if err != nil {
		return fmt.Errorf("failed to calculate fees for estimate: %w", err)
	}

	action.SetFees(fees)

	_, err = q.Put(ctx, m, &consensus.PutOptions{
		MsgIDToReplace: msg.GetId(),
	})

	return err
}

func (k Keeper) calculateFeesForEstimate(
	ctx context.Context,
	relayer sdk.ValAddress,
	chainReferenceID string,
	estimate uint64,
) (*evmtypes.Fees, error) {
	fees := &evmtypes.Fees{}
	multiplicators, err := k.feeProvider.GetCombinedFeesForRelay(ctx, relayer, chainReferenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fee settings: %w", err)
	}

	// Relayer fees are stored as multiplicator of total gas estimate value
	fees.RelayerFee = multiplicators.RelayerFee.
		MulInt(math.NewIntFromUint64(estimate)).
		Ceil().
		TruncateInt().
		Uint64()

		// Community fees are stored as multiplicator of the relayer fee
	fees.CommunityFee = multiplicators.CommunityFee.
		MulInt(math.NewIntFromUint64(fees.RelayerFee)).
		Ceil().
		TruncateInt().
		Uint64()

		// Security fees are stored as multiplicator of the relayer fee
	fees.SecurityFee = multiplicators.SecurityFee.
		MulInt(math.NewIntFromUint64(fees.RelayerFee)).
		Ceil().
		TruncateInt().
		Uint64()

	return fees, nil
}
