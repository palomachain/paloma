package consensus

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/consensus/types"
)

var _ QueueBatcher = BatchQueue{}

type batchOfConsensusMessages = types.BatchOfConsensusMessages

type BatchToBytesToSigner interface {
	BytesToSign(ctx sdk.Context, batch types.Batch) ([]byte, error)
}

type BatchQueue struct {
	base         Queue
	batchToBytes BatchToBytesToSigner
}

func NewBatchQueue(qo QueueOptions, batchToBytes BatchToBytesToSigner) BatchQueue {
	return BatchQueue{
		base:         NewQueue(qo),
		batchToBytes: batchToBytes,
	}
}

func (c BatchQueue) Put(ctx sdk.Context, msgs ...ConsensusMsg) error {
	for _, msg := range msgs {

		if !c.base.qo.TypeCheck(msg) {
			return ErrIncorrectMessageType.Format(msg)
		}

		newID := c.base.qo.Ider.IncrementNextID(ctx, consensusBatchQueueIDCounterKey)

		anyMsg, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return err
		}

		batchedMsg := &batchOfConsensusMessages{
			Msg: anyMsg,
		}

		data, err := c.base.qo.Cdc.MarshalInterface(batchedMsg)
		if err != nil {
			return err
		}

		c.batchQueue(ctx).Set(sdk.Uint64ToBigEndian(newID), data)
	}
	return nil
}

func (c BatchQueue) ProcessBatches(ctx sdk.Context) error {
	queue := c.batchQueue(ctx)
	deleteKeys := [][]byte{}

	var batches []*types.Batch

	var batch *types.Batch
	err := keeperutil.IterAllFnc(queue, c.base.qo.Cdc, func(key []byte, val *batchOfConsensusMessages) bool {
		if batch == nil || len(batch.Msgs) >= consensusQueueMaxBatchSize {
			batch = &types.Batch{}
			batches = append(batches, batch)
		}

		batch.Msgs = append(batch.Msgs, val.GetMsg())
		deleteKeys = append(deleteKeys, key)

		return true
	})

	if len(batch.Msgs) > 0 {
		batches = append(batches, batch)
	}

	if err != nil {
		return err
	}

	// now that we have batches ready, we need to delete those elements from the db
	// and also create consensus messages of those batches.
	for _, deleteKey := range deleteKeys {
		queue.Delete(deleteKey)
	}

	for _, batch := range batches {
		bytesToSign, err := c.batchToBytes.BytesToSign(ctx, *batch)
		if err != nil {
			return err
		}
		batch.BytesToSign = bytesToSign
		err = c.base.Put(ctx, batch)
		if err != nil {
			return err
		}
	}

	return nil
}

// batchQueue returns queue of messages that have been batched
func (c BatchQueue) batchQueue(ctx sdk.Context) prefix.Store {
	store := c.base.qo.Sg.Store(ctx)
	return prefix.NewStore(store, []byte("batching:"+c.base.signingQueueKey()))
}

func (c BatchQueue) AddSignature(ctx sdk.Context, id uint64, signData *types.SignData) error {
	return c.base.AddSignature(ctx, id, signData)
}

func (c BatchQueue) Remove(ctx sdk.Context, msgID uint64) error {
	return c.base.Remove(ctx, msgID)
}

func (c BatchQueue) GetMsgByID(ctx sdk.Context, id uint64) (types.QueuedSignedMessageI, error) {
	return c.base.GetMsgByID(ctx, id)
}

func (c BatchQueue) GetAll(ctx sdk.Context) ([]types.QueuedSignedMessageI, error) {
	return c.base.GetAll(ctx)
}
