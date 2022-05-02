package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	signingutils "github.com/palomachain/utils/signing"
	"github.com/vizualni/whoops"
)

// TODO: add private type for queueTypeName

func AddConcencusQueueType[T ConsensusMsg](k *Keeper, queueTypeName string) {
	if k.queueRegistry == nil {
		k.queueRegistry = make(map[string]consensusQueuer)
	}
	_, ok := k.queueRegistry[queueTypeName]
	if ok {
		panic("concencus queue already registered")
	}

	k.queueRegistry[queueTypeName] = consensusQueue[T]{
		queueTypeName: queueTypeName,
		sg:            *k,
		ider:          k.ider,
		cdc:           k.cdc,
	}
}

// getConsensusQueue gets the consensus queue for the given type.
func (k Keeper) getConsensusQueue(queueTypeName string) (consensusQueuer, error) {
	cq, ok := k.queueRegistry[queueTypeName]
	if !ok {
		return nil, ErrConsensusQueueNotImplemented.Format(queueTypeName)
	}
	return cq, nil
}

func (k Keeper) PutMessageForSigning(ctx sdk.Context, queueTypeName string, msg ConsensusMsg) error {
	cq, err := k.getConsensusQueue(queueTypeName)
	if err != nil {
		k.Logger(ctx).Error("error while getting consensus queue: %s", err)
		return err
	}
	err = cq.put(ctx, msg)
	if err != nil {
		k.Logger(ctx).Error("error while putting message into queue: %s", err)
		return err
	}
	return nil
}

// GetMessagesForSigning returns messages for a single validator that needs to be signed.
func (k Keeper) GetMessagesForSigning(ctx sdk.Context, queueTypeName string, val sdk.ValAddress) (msgs []types.QueuedSignedMessageI, err error) {
	cq, err := k.getConsensusQueue(queueTypeName)
	if err != nil {
		k.Logger(ctx).Error("error while getting consensus queue: %s", err)
		return nil, err
	}
	all, err := cq.getAll(ctx)
	if err != nil {
		k.Logger(ctx).Error("error while getting all messages from queue: %s", err)
		return nil, err
	}

	// TODO: return a max of 100 items (100 chosen at random)
	for _, msg := range all {
		// did this validator already signed this message
		alreadySigned := false
		for _, signData := range msg.GetSignData() {
			if sdk.ValAddress(signData.ValAddress).Equals(val) {
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
func (k Keeper) GetMessagesThatHaveReachedConsensus(ctx sdk.Context, queueTypeName string) ([]types.QueuedSignedMessageI, error) {
	var consensusReached []types.QueuedSignedMessageI

	err := whoops.Try(func() {
		snapshot := whoops.Must(k.valset.GetCurrentSnapshot(ctx))
		cq := whoops.Must(k.getConsensusQueue(queueTypeName))
		msgs := whoops.Must(cq.getAll(ctx))

		validatorMap := make(map[string]valsettypes.Validator)
		for _, validator := range snapshot.GetValidators() {
			validatorMap[validator.Address] = validator
		}

		for _, msg := range msgs {
			msgTotal := sdk.ZeroInt()
			// add shares of validators that have signed the message
			for _, signData := range msg.GetSignData() {
				signedValidator, ok := validatorMap[signData.ValAddress]
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
				k.getConsensusQueue(msg.GetQueueTypeName()),
			)
			wrappedMsg := whoops.Must(
				cq.getMsgByID(ctx, msg.Id),
			)
			rawMsg := whoops.Must(
				wrappedMsg.SdkMsg(),
			)
			bytes := whoops.Must(
				signingutils.JsonDeterministicEncoding(rawMsg),
			)

			bytes = append(bytes, wrappedMsg.Nonce()...)

			if !pk.VerifySignature(bytes, msg.Signature) {
				whoops.Assert(
					ErrSignatureVerificationFailed.Format(
						msg.Id, valAddr, pk.Address(),
					),
				)
			}

			whoops.Assert(
				cq.addSignature(ctx, msg.Id, &types.SignData{
					ValAddress: string(valAddr.Bytes()),
					PubKey:     pk.Bytes(),
					Signature:  msg.Signature,
				}),
			)
		}
	})
}
