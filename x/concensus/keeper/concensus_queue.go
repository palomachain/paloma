package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	keeperutil "github.com/volumefi/cronchain/util/keeper"
	"github.com/volumefi/cronchain/x/concensus/types"
)

const (
	concensusQueueIDCounterKey = `concensus-queue-counter-`
	concensusQueueSigningKey   = `concensus-queue-signing-type-`
)

type (
	ConcensusMsg interface {
		sdk.Msg
		QueueName() string
	}

	codecMarshaler interface {
		MarshalInterface(i proto.Message) ([]byte, error)
		UnmarshalInterface(bz []byte, ptr interface{}) error
	}

	// concensusQueue is a database storing messages that need to be signed.
	// It holds only messages of type `M`. That means that every message type
	// needs to have their own corresponding concensusQueue.
	// The reason for that is so that we can control which message types need
	// to pass the concensus so that we don't accidently put any number of
	// arbitrary messages into the queue and fill up the chain.
	concensusQueue[M ConcensusMsg] struct {
		sg   keeperutil.StoreGetter
		ider keeperutil.IDGenerator
		cdc  codecMarshaler
	}
)

type Message[T ConcensusMsg] struct {
	types.QueuedSignedMessageI
}

// RealMessage returns a an actual real message that was put into the concensus queue.
// e.g. it can be a message to send money from acc A to acc B, or a msg to execute
// a smart contract.
func (m Message[T]) RealMessage() (ret T, err error) {
	sdkMsg, err := m.SdkMsg()
	if err != nil {
		return
	}
	ret, ok := sdkMsg.(T)
	if !ok {
		err = ErrUnableToUnpackSDKMessage
		return
	}
	return ret, nil
}

// put puts raw message into a signing queue.
func (c concensusQueue[M]) put(ctx sdk.Context, msgs ...M) error {
	for _, msg := range msgs {
		newID := c.ider.IncrementNextID(ctx, concensusQueueIDCounterKey)
		anyMsg, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return err
		}
		queuedMsg := &types.QueuedSignedMessage{
			Id:      newID,
			Msg:     anyMsg,
			Signers: []*types.Signer{},
		}
		if err := c.save(ctx, queuedMsg); err != nil {
			return err
		}
	}
	return nil
}

// getAll returns all messages from a signing queu
func (c concensusQueue[M]) getAll(ctx sdk.Context) ([]Message[M], error) {
	var msgs []Message[M]
	queue := c.queue(ctx)
	iterator := queue.Iterator(nil, nil)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		iterData := iterator.Value()

		var sm types.QueuedSignedMessageI
		if err := c.cdc.UnmarshalInterface(iterData, &sm); err != nil {
			return nil, err
		}

		msgs = append(msgs, Message[M]{sm})

	}

	return msgs, nil
}

// addSignature adds a signature to the message. It does not verifies the signature. It
// just adds it to the message.
func (c concensusQueue[M]) addSignature(ctx sdk.Context, msgID uint64, sig *types.Signer) error {
	msg, err := c.getMsgByID(ctx, msgID)
	if err != nil {
		return err
	}

	msg.AddSignature(sig)

	return c.save(ctx, msg)
}

// remove removes the message from the queue.
func (c concensusQueue[M]) remove(ctx sdk.Context, msgID uint64) error {
	_, err := c.getMsgByID(ctx, msgID)
	if err != nil {
		return err
	}
	queue := c.queue(ctx)
	queue.Delete(sdk.Uint64ToBigEndian(msgID))
	return nil
}

// getMsgByID given a message ID, it returns the message
func (c concensusQueue[M]) getMsgByID(ctx sdk.Context, id uint64) (types.QueuedSignedMessageI, error) {
	queue := c.queue(ctx)
	data := queue.Get(sdk.Uint64ToBigEndian(id))

	var sm types.QueuedSignedMessageI
	if err := c.cdc.UnmarshalInterface(data, &sm); err != nil {
		return nil, err
	}

	return sm, nil
}

// save saves the message into the queue
func (c concensusQueue[M]) save(ctx sdk.Context, msg types.QueuedSignedMessageI) error {
	if msg.GetId() == 0 {
		return ErrUnableToSaveMessageWithoutID
	}
	data, err := c.cdc.MarshalInterface(msg)
	if err != nil {
		return err
	}
	c.queue(ctx).Set(sdk.Uint64ToBigEndian(msg.GetId()), data)
	return nil
}

// queue is a simple helper function to return the queue store
func (c concensusQueue[M]) queue(ctx sdk.Context) sdk.KVStore {
	store := c.sg.Store(ctx)
	return prefix.NewStore(store, []byte(c.signingQueueKey()))
}

// signingQueueKey builds a key for the store where are we going to store those.
func (c concensusQueue[M]) signingQueueKey() string {
	var m M
	return concensusQueueSigningKey + m.QueueName()
}
