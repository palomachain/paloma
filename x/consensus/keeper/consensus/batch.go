package consensus

import (
	"context"

	"cosmossdk.io/store/prefix"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
)

var _ QueueBatcher = BatchQueue{}

type batchOfConsensusMessages = types.BatchOfConsensusMessages

type BatchQueue struct {
	base               Queue
	batchedTypeChecker types.TypeChecker
}

func NewBatchQueue(qo QueueOptions) BatchQueue {
	staticTypeCheck := qo.TypeCheck
	batchedTypeCheck := types.BatchedTypeChecker(staticTypeCheck)

	qo.TypeCheck = batchedTypeCheck
	return BatchQueue{
		base:               NewQueue(qo),
		batchedTypeChecker: staticTypeCheck,
	}
}

func (c BatchQueue) Put(ctx context.Context, msg ConsensusMsg, opts *PutOptions) (uint64, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if !c.batchedTypeChecker(msg) {
		return 0, ErrIncorrectMessageType.Format(msg)
	}

	newID := c.base.qo.Ider.IncrementNextID(sdkCtx, consensusBatchQueueIDCounterKey)

	anyMsg, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return 0, err
	}

	var batchedMsg types.MessageQueuedForBatchingI = &batchOfConsensusMessages{
		Msg: anyMsg,
	}

	data, err := c.base.qo.Cdc.MarshalInterface(batchedMsg)
	if err != nil {
		return 0, err
	}
	c.batchQueue(sdkCtx).Set(sdk.Uint64ToBigEndian(newID), data)
	return newID, nil
}

func (c BatchQueue) ProcessBatches(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	queue := c.batchQueue(sdkCtx)
	deleteKeys := [][]byte{}

	iterator := queue.Iterator(nil, nil)
	defer iterator.Close()

	var msgs []types.MessageQueuedForBatchingI
	for ; iterator.Valid(); iterator.Next() {
		iterData := iterator.Value()

		var batchedMsg types.MessageQueuedForBatchingI
		if err := c.base.qo.Cdc.UnmarshalInterface(iterData, &batchedMsg); err != nil {
			return err
		}

		msgs = append(msgs, batchedMsg)
		deleteKeys = append(deleteKeys, iterator.Key())
	}

	var batches []*types.Batch
	var batch *types.Batch

	for _, msg := range msgs {
		if batch == nil || len(batch.Msgs) >= consensusQueueMaxBatchSize {
			batch = &types.Batch{}
			batches = append(batches, batch)
		}

		batch.Msgs = append(batch.Msgs, msg.GetMsg())
	}

	// now that we have batches ready, we need to delete those elements from the db
	// and also create consensus messages of those batches.
	for _, deleteKey := range deleteKeys {
		queue.Delete(deleteKey)
	}

	for _, batch := range batches {
		_, err := c.base.Put(sdkCtx, batch, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// batchQueue returns queue of messages that have been batched
func (c BatchQueue) batchQueue(ctx context.Context) prefix.Store {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := c.base.qo.Sg.Store(sdkCtx)
	return prefix.NewStore(store, []byte("batching:"+c.base.signingQueueKey()))
}

func (c BatchQueue) AddSignature(ctx context.Context, id uint64, signData *types.SignData) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return c.base.AddSignature(sdkCtx, id, signData)
}

func (c BatchQueue) Remove(ctx context.Context, msgID uint64) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return c.base.Remove(sdkCtx, msgID)
}

func (c BatchQueue) GetMsgByID(ctx context.Context, id uint64) (types.QueuedSignedMessageI, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return c.base.GetMsgByID(sdkCtx, id)
}

func (c BatchQueue) GetAll(ctx context.Context) ([]types.QueuedSignedMessageI, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return c.base.GetAll(sdkCtx)
}

func (c BatchQueue) AddEvidence(ctx context.Context, id uint64, evidence *types.Evidence) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return c.base.AddEvidence(sdkCtx, id, evidence)
}

func (c BatchQueue) ReassignValidator(ctx sdk.Context, id uint64, val string) error {
	return c.base.ReassignValidator(ctx, id, val)
}

func (c BatchQueue) SetPublicAccessData(ctx context.Context, id uint64, data *types.PublicAccessData) error {
	return c.base.SetPublicAccessData(ctx, id, data)
}

func (c BatchQueue) GetPublicAccessData(ctx context.Context, id uint64) (*types.PublicAccessData, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return c.base.GetPublicAccessData(sdkCtx, id)
}

func (c BatchQueue) SetErrorData(ctx context.Context, id uint64, data *types.ErrorData) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return c.base.SetErrorData(sdkCtx, id, data)
}

func (c BatchQueue) GetErrorData(ctx context.Context, id uint64) (*types.ErrorData, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return c.base.GetErrorData(sdkCtx, id)
}

func (c BatchQueue) ChainInfo() (types.ChainType, string) {
	return c.base.ChainInfo()
}
