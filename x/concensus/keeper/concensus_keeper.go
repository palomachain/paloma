package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/volumefi/cronchain/x/concensus/types"
)

func addConcencusQueueType[T ConcensusMsg](k *Keeper, queueTypeName string) {
	if k.queueRegistry == nil {
		k.queueRegistry = make(map[string]concensusQueuer)
	}
	_, ok := k.queueRegistry[queueTypeName]
	if ok {
		panic("concencus queue already registered")
	}

	k.queueRegistry[queueTypeName] = concensusQueue[T]{
		queueTypeName: queueTypeName,
		sg:            *k,
		ider:          k.ider,
		cdc:           k.cdc,
	}
}

// getConcensusQueue gets the concensus queue for the given type.
func (k Keeper) getConcensusQueue(queueTypeName string) (concensusQueuer, error) {
	cq, ok := k.queueRegistry[queueTypeName]
	if !ok {
		return nil, fmt.Errorf("type %s: %w", queueTypeName, ErrConcensusQueueNotImplemented)
	}
	return cq, nil
}

func (k Keeper) PutMessageForSigning(ctx sdk.Context, queueTypeName string, msg ConcensusMsg) error {
	cq, err := k.getConcensusQueue(queueTypeName)
	if err != nil {
		k.Logger(ctx).Error("error while getting concensus queue: %s", err)
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
	cq, err := k.getConcensusQueue(queueTypeName)
	if err != nil {
		k.Logger(ctx).Error("error while getting concensus queue: %s", err)
		return nil, err
	}
	all, err := cq.getAll(ctx)
	if err != nil {
		k.Logger(ctx).Error("error while getting all messages from queue: %s", err)
		return nil, err
	}

	// TODO: return a max of 100 items (100 chosen at random
	for _, msg := range all {
		// did this validator already signed this message
		alreadySigned := false
		for _, signer := range msg.GetSigners() {
			if sdk.ValAddress(signer.ValAddress).Equals(val) {
				alreadySigned = true
				break
			}
		}
		if alreadySigned {
			// they did signed it
			continue
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func AddMessageSignature[T ConcensusMsg](ctx sdk.Context, k Keeper, msgID uint64, sig *types.Signer) (err error) {
	// TODO: verify signature here? (or maybe just store the signature and let runners verify them)
	panic("NOT IMPLEMENTED")
	return nil
}
