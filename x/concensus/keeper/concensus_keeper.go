package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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

type concensusQueueKeeper interface{}

// GetConcensusQueue gets the concensus queue for the given type.
func GetConcensusQueue[T ConcensusMsg](k Keeper) concensusQueue[T] {
	var msg T
	cqAny, ok := k.queueRegistry[msg.QueueName()]
	if !ok {
		panic("concencus queue not implemented for type")
	}

	cq := cqAny.(concensusQueue[T])
	return cq
}

func MessagesForSigning[T ConcensusMsg](ctx sdk.Context, k Keeper, val sdk.ValAddress) (msgs []Message[T], err error) {
	cq := GetConcensusQueue[T](k)
	all, err := cq.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, msg := range all {
		alreadySigned := false
		// TODO: this is using wrong method for GetSigners
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

