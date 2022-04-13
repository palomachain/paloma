package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/volumefi/cronchain/x/consensus/types"
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

func (k Keeper) AddMessageSignature(ctx sdk.Context, msgID uint64, queueTypeName string, signature []byte) (err error) {
	// TODO: once we have validatorsets get them here, take their keys and validate the signatures
	// cq, err := k.getConsensusQueue(queueTypeName)
	// if err != nil {
	// 	return err
	// }

	return nil
}
