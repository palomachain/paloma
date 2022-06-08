package consensus

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/consensus/types"
)

var _ Queuer = Queue{}

type ConsensusMsg = types.ConsensusMsg

type codecMarshaler interface {
	MarshalInterface(i proto.Message) ([]byte, error)
	UnmarshalInterface(bz []byte, ptr interface{}) error
	Unmarshal(bz []byte, ptr codec.ProtoMarshaler) error
}

// Queue is a database storing messages that need to be signed.
type Queue struct {
	qo QueueOptions
}

type QueueOptions struct {
	QueueTypeName         types.ConsensusQueueType
	Sg                    keeperutil.StoreGetter
	Ider                  keeperutil.IDGenerator
	Cdc                   codecMarshaler
	TypeCheck             types.TypeChecker
	BytesToSignCalculator types.BytesToSignFunc
	VerifySignature       types.VerifySignatureFunc
	ChainType             types.ChainType
	ChainID               string
}

type OptFnc func(*QueueOptions)

func WithQueueTypeName(val types.ConsensusQueueType) OptFnc {
	return func(opt *QueueOptions) {
		opt.QueueTypeName = val
	}
}

func WithStaticTypeCheck(val any) OptFnc {
	return func(opt *QueueOptions) {
		opt.TypeCheck = types.StaticTypeChecker(val)
	}
}

func WithBytesToSignCalc(val types.BytesToSignFunc) OptFnc {
	return func(opt *QueueOptions) {
		opt.BytesToSignCalculator = val
	}
}

func WithVerifySignature(val types.VerifySignatureFunc) OptFnc {
	return func(opt *QueueOptions) {
		opt.VerifySignature = val
	}
}

func WithChainInfo(chainType, chainID string) OptFnc {
	return func(opt *QueueOptions) {
		opt.ChainID = chainID
		opt.ChainType = types.ChainType(chainType)
	}
}

func NewQueue(qo QueueOptions) Queue {
	if qo.BytesToSignCalculator == nil {
		panic("BytesToSignCalculator can't be nil")
	}
	if qo.VerifySignature == nil {
		panic("VerifySignature can't be nil")
	}

	if len(qo.ChainType) == 0 {
		panic("chain type can't be empty")
	}

	if len(qo.ChainID) == 0 {
		panic("chain id can't be empty")
	}

	if len(qo.QueueTypeName) == 0 {
		panic("queue type name can't be empty")
	}
	return Queue{
		qo: qo,
	}
}

// Put puts raw message into a signing queue.
func (c Queue) Put(ctx sdk.Context, msgs ...ConsensusMsg) error {
	for _, msg := range msgs {
		if !c.qo.TypeCheck(msg) {
			return ErrIncorrectMessageType.Format(msg)
		}
		newID := c.qo.Ider.IncrementNextID(ctx, consensusQueueIDCounterKey)
		// just so it's clear that nonce is an actual ID
		nonce := newID
		anyMsg, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return err
		}
		queuedMsg := &types.QueuedSignedMessage{
			Id:          newID,
			Msg:         anyMsg,
			SignData:    []*types.SignData{},
			BytesToSign: c.qo.BytesToSignCalculator(msg, nonce),
		}
		if err := c.save(ctx, queuedMsg); err != nil {
			return err
		}
	}
	return nil
}

// getAll returns all messages from a signing queu
func (c Queue) GetAll(ctx sdk.Context) ([]types.QueuedSignedMessageI, error) {
	var msgs []types.QueuedSignedMessageI
	queue := c.queue(ctx)
	iterator := queue.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		iterData := iterator.Value()

		var sm types.QueuedSignedMessageI
		if err := c.qo.Cdc.UnmarshalInterface(iterData, &sm); err != nil {
			return nil, err
		}

		msgs = append(msgs, sm)
	}

	return msgs, nil
}

// AddSignature adds a signature to the message and checks if the signature is valid.
func (c Queue) AddSignature(ctx sdk.Context, msgID uint64, pk []byte, signData *types.SignData) error {
	msg, err := c.GetMsgByID(ctx, msgID)
	if err != nil {
		return err
	}

	if !c.qo.VerifySignature(append(msg.GetBytesToSign()[:], signData.GetExtraData()...), signData.Signature, pk) {
		return ErrInvalidSignature
	}

	msg.AddSignData(signData)

	return c.save(ctx, msg)
}

// remove removes the message from the queue.
func (c Queue) Remove(ctx sdk.Context, msgID uint64) error {
	_, err := c.GetMsgByID(ctx, msgID)
	if err != nil {
		return err
	}
	queue := c.queue(ctx)
	queue.Delete(sdk.Uint64ToBigEndian(msgID))
	return nil
}

// getMsgByID given a message ID, it returns the message
func (c Queue) GetMsgByID(ctx sdk.Context, id uint64) (types.QueuedSignedMessageI, error) {
	queue := c.queue(ctx)
	data := queue.Get(sdk.Uint64ToBigEndian(id))

	if data == nil {
		return nil, ErrMessageDoesNotExist.Format(id)
	}

	var sm types.QueuedSignedMessageI
	if err := c.qo.Cdc.UnmarshalInterface(data, &sm); err != nil {
		return nil, err
	}

	return sm, nil
}

func (c Queue) ChainInfo() (types.ChainType, string) {
	return c.qo.ChainType, c.qo.ChainID
}

func (c Queue) ConsensusQueue() string {
	return types.Queue(c.qo.QueueTypeName, c.qo.ChainType, c.qo.ChainID)
}

// save saves the message into the queue
func (c Queue) save(ctx sdk.Context, msg types.QueuedSignedMessageI) error {
	if msg.GetId() == 0 {
		return ErrUnableToSaveMessageWithoutID
	}
	data, err := c.qo.Cdc.MarshalInterface(msg)
	if err != nil {
		return err
	}
	c.queue(ctx).Set(sdk.Uint64ToBigEndian(msg.GetId()), data)
	return nil
}

// queue is a simple helper function to return the queue store
func (c Queue) queue(ctx sdk.Context) prefix.Store {
	store := c.qo.Sg.Store(ctx)
	return prefix.NewStore(store, []byte(c.signingQueueKey()))
}

// signingQueueKey builds a key for the store where are we going to store those.
func (c Queue) signingQueueKey() string {
	if c.qo.QueueTypeName == "" {
		panic("queueTypeName can't be empty")
	}
	return fmt.Sprintf("%s-%s", consensusQueueSigningKey, c.ConsensusQueue())
}
