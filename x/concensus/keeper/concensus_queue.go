package keeper

import (
	"fmt"

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

type ConcensusMsg interface {
	sdk.Msg
	QueueName() string
}

type codecMarshaler interface {
	MarshalInterface(i proto.Message) ([]byte, error)
	UnmarshalInterface(bz []byte, ptr interface{}) error
}

type concensusQueue[M ConcensusMsg] struct {
	sg   keeperutil.StoreGetter
	ider keeperutil.IDGenerator
	cdc  codecMarshaler
}

// TODO: ovo je samo db
type concensusQueuer[T ConcensusMsg] interface {
	Put()
	Sign()
	All()
	Delete()
}

type concensusKeeper interface {
	Put()
	Sign()
	QueryAll()
	Delete()
}

type Message[T ConcensusMsg] struct {
	types.QueuedSignedMessageI
}

func (m Message[T]) RealMessage() (ret T, err error) {
	sdkMsg, err := m.SdkMsg()
	if err != nil {
		return
	}
	ret, ok := sdkMsg.(T)
	if !ok {
		err = fmt.Errorf("WHAAAAAAAAT")
		return
	}
	return ret, nil
}

// Put puts raw message into a signing queue.
func (c concensusQueue[M]) Put(ctx sdk.Context, msgs ...M) error {
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

func (c concensusQueue[M]) GetAll(ctx sdk.Context) ([]Message[M], error) {
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

func (c concensusQueue[M]) AddSignature(ctx sdk.Context, msgID uint64, sig *types.Signer) error {
	msg, err := c.getMsgByID(ctx, msgID)
	if err != nil {
		return err
	}

	msg.AddSignature(sig)

	return c.save(ctx, msg)
}

func (c concensusQueue[M]) Remove(ctx sdk.Context, msgID uint64) error {
	_, err := c.getMsgByID(ctx, msgID)
	if err != nil {
		return err
	}
	queue := c.queue(ctx)
	queue.Delete(sdk.Uint64ToBigEndian(msgID))
	return nil
}

func (c concensusQueue[M]) getMsgByID(ctx sdk.Context, id uint64) (types.QueuedSignedMessageI, error) {
	queue := c.queue(ctx)
	data := queue.Get(sdk.Uint64ToBigEndian(id))

	var sm types.QueuedSignedMessageI
	if err := c.cdc.UnmarshalInterface(data, &sm); err != nil {
		return nil, err
	}

	return sm, nil
}

func (c concensusQueue[M]) save(ctx sdk.Context, msg types.QueuedSignedMessageI) error {
	if msg.GetId() == 0 {
		return fmt.Errorf("TODO ERROR ID IS NOT SET")
	}
	data, err := c.cdc.MarshalInterface(msg)
	if err != nil {
		return err
	}
	c.queue(ctx).Set(sdk.Uint64ToBigEndian(msg.GetId()), data)
	return nil
}

func (c concensusQueue[M]) queue(ctx sdk.Context) sdk.KVStore {
	store := c.sg.Store(ctx)
	return prefix.NewStore(store, []byte(c.signingQueueKey()))
}

func (c concensusQueue[M]) signingQueueKey() string {
	var m M
	return concensusQueueSigningKey + m.QueueName()
}
