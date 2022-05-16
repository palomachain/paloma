package keeper

import (
	"fmt"
	"reflect"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/consensus/types"
)

const (
	consensusQueueIDCounterKey = `consensus-queue-counter-`
	consensusQueueSigningKey   = `consensus-queue-signing-type-`
)

type ConsensusMsg = types.ConsensusMsg

//go:generate mockery --name=consensusQueuer --inpackage --testonly
type (
	codecMarshaler interface {
		MarshalInterface(i proto.Message) ([]byte, error)
		UnmarshalInterface(bz []byte, ptr interface{}) error
	}

	// consensusQueue is a database storing messages that need to be signed.
	consensusQueue struct {
		queueTypeName types.ConsensusQueueType
		sg            keeperutil.StoreGetter
		ider          keeperutil.IDGenerator
		cdc           codecMarshaler
		typeCheck     TypeChecker
	}

	consensusQueuer interface {
		put(sdk.Context, ...ConsensusMsg) error
		getAll(sdk.Context) ([]types.QueuedSignedMessageI, error)
		getMsgByID(ctx sdk.Context, id uint64) (types.QueuedSignedMessageI, error)
		addSignature(sdk.Context, uint64, *types.SignData) error
		remove(sdk.Context, uint64) error
	}
)

type TypeChecker func(any) bool

func StaticTypeChecker(typ any) TypeChecker {
	return func(i any) bool {
		expectedType := reflect.TypeOf(typ)
		gotType := reflect.TypeOf(i)
		return expectedType == gotType
	}
}

// put puts raw message into a signing queue.
func (c consensusQueue) put(ctx sdk.Context, msgs ...ConsensusMsg) error {
	for _, msg := range msgs {
		if !c.typeCheck(msg) {
			return ErrIncorrectMessageType.Format(msg)
		}
		newID := c.ider.IncrementNextID(ctx, consensusQueueIDCounterKey)
		anyMsg, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return err
		}
		queuedMsg := &types.QueuedSignedMessage{
			Id:       newID,
			Msg:      anyMsg,
			SignData: []*types.SignData{},
		}
		if err := c.save(ctx, queuedMsg); err != nil {
			return err
		}
	}
	return nil
}

// getAll returns all messages from a signing queu
func (c consensusQueue) getAll(ctx sdk.Context) ([]types.QueuedSignedMessageI, error) {
	var msgs []types.QueuedSignedMessageI
	queue := c.queue(ctx)
	iterator := queue.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		iterData := iterator.Value()

		var sm types.QueuedSignedMessageI
		if err := c.cdc.UnmarshalInterface(iterData, &sm); err != nil {
			return nil, err
		}

		msgs = append(msgs, sm)
	}

	return msgs, nil
}

// addSignature adds a signature to the message. It does not verifies the signature. It
// just adds it to the message.
func (c consensusQueue) addSignature(ctx sdk.Context, msgID uint64, signData *types.SignData) error {
	msg, err := c.getMsgByID(ctx, msgID)
	if err != nil {
		return err
	}

	msg.AddSignData(signData)

	return c.save(ctx, msg)
}

// remove removes the message from the queue.
func (c consensusQueue) remove(ctx sdk.Context, msgID uint64) error {
	_, err := c.getMsgByID(ctx, msgID)
	if err != nil {
		return err
	}
	queue := c.queue(ctx)
	queue.Delete(sdk.Uint64ToBigEndian(msgID))
	return nil
}

// getMsgByID given a message ID, it returns the message
func (c consensusQueue) getMsgByID(ctx sdk.Context, id uint64) (types.QueuedSignedMessageI, error) {
	queue := c.queue(ctx)
	data := queue.Get(sdk.Uint64ToBigEndian(id))

	if data == nil {
		return nil, ErrMessageDoesNotExist.Format(id)
	}

	var sm types.QueuedSignedMessageI
	if err := c.cdc.UnmarshalInterface(data, &sm); err != nil {
		return nil, err
	}

	return sm, nil
}

// save saves the message into the queue
func (c consensusQueue) save(ctx sdk.Context, msg types.QueuedSignedMessageI) error {
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
func (c consensusQueue) queue(ctx sdk.Context) prefix.Store {
	store := c.sg.Store(ctx)
	return prefix.NewStore(store, []byte(c.signingQueueKey()))
}

// signingQueueKey builds a key for the store where are we going to store those.
func (c consensusQueue) signingQueueKey() string {
	if c.queueTypeName == "" {
		panic("queueTypeName can't be empty")
	}
	return fmt.Sprintf("%s-%s", consensusQueueSigningKey, string(c.queueTypeName))
}
