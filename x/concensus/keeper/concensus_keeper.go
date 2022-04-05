package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/volumefi/cronchain/x/concensus/types"
)

func addConcencusQueueType[T ConcensusMsg](k *Keeper) {
	var msg T
	_, ok := k.queueRegistry[msg.QueueName()]
	if ok {
		panic("concencus queue already registered")
	}

	k.queueRegistry[msg.QueueName()] = concensusQueue[T]{
		sg:   *k,
		ider: k.ider,
		cdc:  k.cdc,
	}
}

// getConcensusQueue gets the concensus queue for the given type.
func getConcensusQueue[T ConcensusMsg](k Keeper) concensusQueue[T] {
	var msg T
	cqAny, ok := k.queueRegistry[msg.QueueName()]
	if !ok {
		panic("concencus queue not implemented for type")
	}

	cq := cqAny.(concensusQueue[T])
	return cq
}

// GetMessagesForSigning returns messages for a single validator that needs to be signed.
func GetMessagesForSigning[T ConcensusMsg](ctx sdk.Context, k Keeper, val sdk.ValAddress) (msgs []Message[T], err error) {
	cq := getConcensusQueue[T](k)
	all, err := cq.getAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, msg := range all {
		alreadySigned := false
		for _, signer := range msg.GetSigners() {
			if sdk.ValAddress(signer.ValAddress).Equals(val) {
				alreadySigned = true
				break
			}
		}
		if alreadySigned {
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

