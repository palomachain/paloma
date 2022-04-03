package keeper

import (
	"context"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/volumefi/cronchain/x/scheduler/types"
)

const (
	concensusQueueIDCounterKey = `concensus-queue-counter-`
	concensusQueueSigningKey   = `concensus-queue-signing-type-`
)

type ConcensusMsg interface {
	sdk.Msg
	QueueName() string
}

type codecMarshaler interface {
	MarshalInterface(i proto.Message) ([]byte, error)
	UnmarshalInterface(bz []byte, ptr interface{}) error
}

type concensusQueue[M ConcensusMsg] struct {
	storeGetter storeGetter
	ider        ider
	cdc         codecMarshaler
}

// Put puts raw message into a signing queue.
func (c concensusQueue[M]) Put(ctx sdk.Context, msg M) error {
	store := c.storeGetter.Store(ctx)
	newID := c.ider.incrementNextID(ctx, concensusQueueIDCounterKey)
	prefixStore := prefix.NewStore(store, []byte(c.signingQueueKey()))
	anyMsg, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	queuedMsg := &types.QueuedSignedMessage{
		Msg:     anyMsg,
		Signers: []*types.Signer{},
	}
	data, err := c.cdc.MarshalInterface(queuedMsg)
	if err != nil {
		return err
	}
	prefixStore.Set(uint64ToByte(newID), data)
	return nil
}

func (c concensusQueue[M]) GetMsgsForSigning(ctx sdk.Context, valaddr sdk.ValAddress) ([]M, error) {
	var msgs []M
	store := c.storeGetter.Store(ctx)
	prefixStore := prefix.NewStore(store, []byte(c.signingQueueKey()))
	iterator := prefixStore.Iterator(nil, nil)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		iterData := iterator.Value()
		fmt.Println("BLAAAA", iterData, string(iterData))
		var sm types.QueuedSignedMessageI
		if err := c.cdc.UnmarshalInterface(iterData, &sm); err != nil {
			fmt.Println("SRANJE", err)
			return nil, err
		}
		fmt.Println(sm.SdkMsg())
		anyMsg, err := sm.SdkMsg()
		if err != nil {
			return nil, err
		}
		msg, ok := anyMsg.(M)
		if !ok {
			return nil, fmt.Errorf("TODO")
		}

		foundSignature := false
		// for _, sg := range sm.GetSigners() {
		// 	// if sg.SignedBy.Equals(valaddr) {
		// 	// 	foundSignature = true
		// 	// 	break
		// 	// }
		// }
		if foundSignature {
			continue
		}

		msgs = append(msgs, msg)
	}
	fmt.Println(msgs)
	return msgs, nil
}

func (c concensusQueue[M]) PutSignedMessages(ctx context.Context, valaddr sdk.ValAddress, tx sdk.Tx) ([]M, error) {
	return nil, nil
}

func (c concensusQueue[M]) GetConcensusReachedMsgs(ctx context.Context, valaddr sdk.ValAddress, txs []sdk.Tx) ([]M, error) {
	return nil, nil
}

func (c concensusQueue[M]) signingQueueKey() string {
	var m M
	return concensusQueueSigningKey + m.QueueName()
}
