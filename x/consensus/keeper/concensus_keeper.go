package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	"github.com/palomachain/paloma/x/consensus/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	signingutils "github.com/palomachain/utils/signing"
	"github.com/vizualni/whoops"
)

const (
	encodingDelimiter = byte('|')
)

func (k Keeper) AddConcencusQueueType(queueTypeName types.ConsensusQueueType, typ any, bytesToSign types.BytesToSignFunc) {
	cq := consensus.NewQueue(consensus.QueueOptions{
		QueueTypeName:         queueTypeName,
		Sg:                    k,
		Ider:                  k.ider,
		Cdc:                   k.cdc,
		TypeCheck:             types.StaticTypeChecker(typ),
		BytesToSignCalculator: bytesToSign,
	})
	k.addConcencusQueueType(queueTypeName, cq)
}

func (k Keeper) AddBatchedConcencusQueueType(
	queueTypeName types.ConsensusQueueType,
	typ any,
	bytesToSign types.BytesToSignFunc,
) {
	cq := consensus.NewBatchQueue(consensus.QueueOptions{
		QueueTypeName:         queueTypeName,
		Sg:                    k,
		Ider:                  k.ider,
		Cdc:                   k.cdc,
		TypeCheck:             types.StaticTypeChecker(typ),
		BytesToSignCalculator: bytesToSign,
	})
	k.addConcencusQueueType(queueTypeName, cq)
}

func (k Keeper) addConcencusQueueType(queueTypeName types.ConsensusQueueType, cq consensus.Queuer) {
	if k.queueRegistry == nil {
		k.queueRegistry = make(map[types.ConsensusQueueType]consensus.Queuer)
	}
	_, ok := k.queueRegistry[queueTypeName]
	if ok {
		panic("concencus queue already registered")
	}

	k.queueRegistry[queueTypeName] = cq
}

// getConsensusQueue gets the consensus queue for the given type.
func (k Keeper) getConsensusQueue(queueTypeName types.ConsensusQueueType) (consensus.Queuer, error) {
	cq, ok := k.queueRegistry[queueTypeName]
	if !ok {
		return nil, ErrConsensusQueueNotImplemented.Format(queueTypeName)
	}
	return cq, nil
}

func (k Keeper) PutMessageForSigning(ctx sdk.Context, queueTypeName types.ConsensusQueueType, msg consensus.ConsensusMsg) error {
	cq, err := k.getConsensusQueue(queueTypeName)
	if err != nil {
		k.Logger(ctx).Error("error while getting consensus queue: %s", err)
		return err
	}
	err = cq.Put(ctx, msg)
	if err != nil {
		k.Logger(ctx).Error("error while putting message into queue: %s", err)
		return err
	}
	return nil
}

// GetMessagesForSigning returns messages for a single validator that needs to be signed.
func (k Keeper) GetMessagesForSigning(ctx sdk.Context, queueTypeName types.ConsensusQueueType, val sdk.ValAddress) (msgs []types.QueuedSignedMessageI, err error) {
	cq, err := k.getConsensusQueue(queueTypeName)
	if err != nil {
		k.Logger(ctx).Error("error while getting consensus queue: %s", err)
		return nil, err
	}
	all, err := cq.GetAll(ctx)
	if err != nil {
		k.Logger(ctx).Error("error while getting all messages from queue: %s", err)
		return nil, err
	}

	// TODO: return a max of 100 items (100 chosen at random)
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

// GetMessagesThatHaveReachedConsensus returns messages from a given
// queueTypeName that have reached consensus based on the latest snapshot
// available.
func (k Keeper) GetMessagesThatHaveReachedConsensus(ctx sdk.Context, queueTypeName types.ConsensusQueueType) ([]types.QueuedSignedMessageI, error) {
	var consensusReached []types.QueuedSignedMessageI

	err := whoops.Try(func() {
		cq := whoops.Must(k.getConsensusQueue(queueTypeName))
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
	msgs []*types.MsgAddMessagesSignatures_MsgSignedMessage,
) error {
	pk := k.valset.GetSigningKey(ctx, valAddr)
	if pk == nil {
		return ErrUnableToFindPubKeyForValidator.Format(valAddr)
	}

	return whoops.Try(func() {
		for _, msg := range msgs {
			cq := whoops.Must(
				k.getConsensusQueue(types.ConsensusQueueType(msg.GetQueueTypeName())),
			)
			consensusMsg := whoops.Must(
				cq.GetMsgByID(ctx, msg.Id),
			)
			rawMsg := whoops.Must(
				consensusMsg.ConsensusMsg(),
			)

			if ok := signingutils.VerifySignature(
				pk,
				signingutils.SerializeFnc(signingutils.JsonDeterministicEncoding),
				msg.GetSignature(),
				rawMsg,
				consensusMsg.Nonce(),
				msg.GetExtraData(),
			); !ok {
				whoops.Assert(
					ErrSignatureVerificationFailed.Format(
						msg.Id, valAddr, pk.Address(),
					),
				)
			}

			if task, ok := rawMsg.(types.AttestTask); ok {
				whoops.Assert(
					k.attestator.validateIncoming(
						ctx.Context(),
						types.ConsensusQueueType(msg.GetQueueTypeName()),
						task,
						types.Evidence{
							From: valAddr,
							Data: msg.GetExtraData(),
						},
					),
				)
			}

			whoops.Assert(
				cq.AddSignature(ctx, msg.Id, &types.SignData{
					ValAddress: valAddr,
					Signature:  msg.GetSignature(),
					ExtraData:  msg.GetExtraData(),
				}),
			)
		}
	})
}

func nonceFromID(id uint64) []byte {
	return sdk.Uint64ToBigEndian(id)
}

func queuedMessageToMessageToSign(msg types.QueuedSignedMessageI) *types.MessageToSign {
	consensusMsg, err := msg.ConsensusMsg()
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
