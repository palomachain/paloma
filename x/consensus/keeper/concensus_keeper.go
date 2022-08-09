package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	"github.com/palomachain/paloma/x/consensus/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/vizualni/whoops"
)

const (
	encodingDelimiter = byte('|')
)

// getConsensusQueue gets the consensus queue for the given type.
func (k Keeper) getConsensusQueue(ctx sdk.Context, queueTypeName string) (consensus.Queuer, error) {
	for _, q := range k.registry.slice {
		supportedQueues, err := q.SupportedQueues(ctx)
		if err != nil {
			return nil, err
		}
		opts, ok := supportedQueues[queueTypeName]
		if !ok {
			continue
		}

		if opts.Ider.Zero() {
			opts.Ider = k.ider
		}

		if opts.Sg == nil {
			opts.Sg = k
		}

		if opts.Cdc == nil {
			opts.Cdc = k.cdc
		}

		if opts.Batched {
			return consensus.NewBatchQueue(opts.QueueOptions), nil
		}
		return consensus.NewQueue(opts.QueueOptions), nil
	}
	return nil, ErrConsensusQueueNotImplemented.Format(queueTypeName)
}

func (k Keeper) RemoveConsensusQueue(ctx sdk.Context, queueTypeName string) error {
	cq, err := k.getConsensusQueue(ctx, queueTypeName)
	if err != nil {
		return err
	}
	consensus.RemoveQueueCompletely(ctx, cq)
	return nil
}

func (k Keeper) PutMessageForSigning(ctx sdk.Context, queueTypeName string, msg consensus.ConsensusMsg) error {
	cq, err := k.getConsensusQueue(ctx, queueTypeName)
	if err != nil {
		k.Logger(ctx).Error("error while getting consensus queue", "error", err)
		return err
	}
	err = cq.Put(ctx, msg)
	if err != nil {
		k.Logger(ctx).Error("error while putting message into queue", "error", err)
		return err
	}
	k.Logger(ctx).Info(
		"put message into consensus queue",
		"queue-type-name", queueTypeName,
	)
	return nil
}

// GetMessagesForSigning returns messages for a single validator that needs to be signed.
func (k Keeper) GetMessagesForSigning(ctx sdk.Context, queueTypeName string, val sdk.ValAddress) (msgs []types.QueuedSignedMessageI, err error) {
	all, err := k.GetMessagesFromQueue(ctx, queueTypeName, 1000)
	if err != nil {
		return nil, err
	}
	for _, msg := range all {
		// did this validator already signed this message
		alreadySigned := false
		for _, signData := range msg.GetSignData() {
			if signData.ValAddress.Equals(val) {
				alreadySigned = true
				break
			}
		}
		if alreadySigned {
			// they have already signed it
			continue
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// GetMessagesFromQueue gets N messages from the queue.
func (k Keeper) GetMessagesFromQueue(ctx sdk.Context, queueTypeName string, n int) (msgs []types.QueuedSignedMessageI, err error) {
	if n <= 0 {
		return nil, ErrInvalidLimitValue.Format(n)
	}
	cq, err := k.getConsensusQueue(ctx, queueTypeName)
	if err != nil {
		k.Logger(ctx).Error("error while getting consensus queue", "err", err)
		return nil, err
	}
	msgs, err = cq.GetAll(ctx)
	if err != nil {
		k.Logger(ctx).Error("error while getting all messages from queue", "err", err)
		return nil, err
	}

	if len(msgs) > n {
		msgs = msgs[:n]
	}

	return
}

func (k Keeper) DeleteJob(ctx sdk.Context, queueTypeName string, id uint64) (err error) {
	cq, err := k.getConsensusQueue(ctx, queueTypeName)
	if err != nil {
		k.Logger(ctx).Error("error while getting consensus queue", "err", err)
		return err
	}
	return cq.Remove(ctx, id)
}

// GetMessagesThatHaveReachedConsensus returns messages from a given
// queueTypeName that have reached consensus based on the latest snapshot
// available.
func (k Keeper) GetMessagesThatHaveReachedConsensus(ctx sdk.Context, queueTypeName string) ([]types.QueuedSignedMessageI, error) {
	var consensusReached []types.QueuedSignedMessageI

	err := whoops.Try(func() {
		cq, err := k.getConsensusQueue(ctx, queueTypeName)
		whoops.Assert(err)

		msgs := whoops.Must(cq.GetAll(ctx))
		if len(msgs) == 0 {
			return
		}
		snapshot := whoops.Must(k.valset.GetCurrentSnapshot(ctx))

		if len(snapshot.Validators) == 0 || snapshot.TotalShares.Equal(sdk.ZeroInt()) {
			return
		}

		validatorMap := make(map[string]valsettypes.Validator)
		for _, validator := range snapshot.GetValidators() {
			validatorMap[validator.Address.String()] = validator
		}

		for _, msg := range msgs {
			msgTotal := sdk.ZeroInt()
			// add shares of validators that have signed the message
			for _, signData := range msg.GetSignData() {
				signedValidator, ok := validatorMap[signData.ValAddress.String()]
				if !ok {
					k.Logger(ctx).Info("validator not found", "validator", signData.ValAddress)
					continue
				}
				msgTotal = msgTotal.Add(signedValidator.ShareCount)
			}

			// Now we need to check if the consensus was reached. We do this by
			// checking if there were at least 2/3 of total signatures in
			// staking power.
			// The formula goes like this:
			// msgTotal >= 2/3 * snapshotTotal
			// the "issue" now becomes: 2/3 which could be problematic as we
			// could lose precision using floating point arithmetic.
			// If we multiply both sides with 3, we don't need to do division.
			// 3 * msgTotal >= 2 * snapshotTotal
			if msgTotal.Mul(sdk.NewInt(3)).GTE(snapshot.TotalShares.Mul(sdk.NewInt(2))) {
				// consensus has been reached
				consensusReached = append(consensusReached, msg)
			}
		}
	})
	if err != nil {
		return nil, err
	}
	return consensusReached, nil
}

// AddMessageSignature adds signatures to the messages.
func (k Keeper) AddMessageSignature(
	ctx sdk.Context,
	valAddr sdk.ValAddress,
	msgs []*types.ConsensusMessageSignature,
) error {
	err := whoops.Try(func() {
		for _, msg := range msgs {
			cq := whoops.Must(
				k.getConsensusQueue(ctx, msg.GetQueueTypeName()),
			)
			chainType, chainReferenceID := cq.ChainInfo()

			publicKey := whoops.Must(k.valset.GetSigningKey(
				ctx,
				valAddr,
				chainType,
				chainReferenceID,
				msg.GetSignedByAddress(),
			))

			whoops.Assert(
				cq.AddSignature(
					ctx,
					msg.Id,
					&types.SignData{
						ValAddress:             valAddr,
						Signature:              msg.GetSignature(),
						ExternalAccountAddress: msg.GetSignedByAddress(),
						PublicKey:              publicKey,
					},
				),
			)

			k.Logger(ctx).Info("added message signature",
				"message-id", msg.GetId(),
				"queue-type-name", msg.GetQueueTypeName(),
				"signed-by-address", msg.GetSignedByAddress(),
				"chain-type", chainType,
				"chain-reference-id", chainReferenceID,
			)
		}
	})

	if err != nil {
		k.Logger(ctx).Error("error while adding messages signatures",
			"err", err,
		)
	}

	return err
}

func (k Keeper) AddMessageEvidence(
	ctx sdk.Context,
	valAddr sdk.ValAddress,
	msg *types.MsgAddEvidence,
) error {
	err := whoops.Try(func() {
		cq := whoops.Must(
			k.getConsensusQueue(ctx, msg.GetQueueTypeName()),
		)

		whoops.Assert(
			cq.AddEvidence(
				ctx,
				msg.GetMessageID(),
				&types.Evidence{
					ValAddress: valAddr,
					Proof:      msg.GetProof(),
				},
			),
		)
		chainType, chainReferenceID := cq.ChainInfo()
		k.Logger(ctx).Info("added message evidence",
			"message-id", msg.GetMessageID(),
			"queue-type-name", msg.GetQueueTypeName(),
			"chain-type", chainType,
			"chain-reference-id", chainReferenceID,
		)
	})
	if err != nil {
		k.Logger(ctx).Error("error while adding message evidence",
			"err", err,
		)
	}

	return err
}

func (k Keeper) SetMessagePublicAccessData(
	ctx sdk.Context,
	valAddr sdk.ValAddress,
	msg *types.MsgSetPublicAccessData,
) error {
	err := whoops.Try(func() {
		cq := whoops.Must(
			k.getConsensusQueue(ctx, msg.GetQueueTypeName()),
		)

		whoops.Assert(
			cq.SetPublicAccessData(
				ctx,
				msg.GetMessageID(),
				&types.PublicAccessData{
					ValAddress: valAddr,
					Data:       msg.GetData(),
				},
			),
		)
		chainType, chainReferenceID := cq.ChainInfo()
		k.Logger(ctx).Info("added message public access data",
			"message-id", msg.GetMessageID(),
			"queue-type-name", msg.GetQueueTypeName(),
			"chain-type", chainType,
			"chain-reference-id", chainReferenceID,
		)
	})
	if err != nil {
		k.Logger(ctx).Error("error while adding message public access data",
			"err", err,
		)
	}

	return err
}

func nonceFromID(id uint64) []byte {
	return sdk.Uint64ToBigEndian(id)
}

func (k Keeper) queuedMessageToMessageToSign(msg types.QueuedSignedMessageI) *types.MessageToSign {
	consensusMsg, err := msg.ConsensusMsg(k.cdc)
	if err != nil {
		panic(err)
	}
	anyMsg, err := codectypes.NewAnyWithValue(consensusMsg)
	if err != nil {
		panic(err)
	}
	return &types.MessageToSign{
		Nonce:       nonceFromID(msg.GetId()),
		Id:          msg.GetId(),
		BytesToSign: msg.GetBytesToSign(),
		Msg:         anyMsg,
	}
}
