package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/vizualni/whoops"
	"github.com/volumefi/cronchain/x/consensus/types"
	signingutils "github.com/volumefi/utils/signing"
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
						msg.Id, valAddr, pk.String(),
					),
				)
			}

			whoops.Assert(
				cq.addSignature(ctx, msg.Id, &types.SignData{
					ValAddress: string(valAddr.Bytes()),
					PubKey:     pk.String(),
					Signature:  msg.Signature,
				}),
			)
		}
	})
}
