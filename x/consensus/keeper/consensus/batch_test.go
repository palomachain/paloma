package consensus

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	testtypes "github.com/palomachain/paloma/x/consensus/testdata/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

func TestBatching(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, sdk.StoreTypeMemory, nil)
	assert.NoError(t, stateStore.LoadLatestVersion())

	registry := types.ModuleCdc.InterfaceRegistry()
	registry.RegisterInterface(
		"volumefi.tests.SimpleMessage",
		(*types.ConsensusMsg)(nil),
		&testtypes.SimpleMessage{},
	)
	registry.RegisterImplementations(
		(*types.ConsensusMsg)(nil),
		&testtypes.SimpleMessage{},
	)
	types.RegisterInterfaces(registry)

	sg := keeperutil.SimpleStoreGetter(stateStore.GetKVStore(storeKey))
	batchSigner := BatchSignerFunc(func(ctx sdk.Context, batch types.Batch) ([]byte, error) {
		return []byte("hello"), nil
	})
	cq := NewBatchQueue(
		QueueOptions{
			QueueTypeName: types.ConsensusQueueType("simple-message"),
			Sg:            sg,
			Ider:          keeperutil.NewIDGenerator(sg, nil),
			Cdc:           types.ModuleCdc,
			TypeCheck:     types.StaticTypeChecker(&testtypes.SimpleMessage{}),
		},
		batchSigner,
	)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, nil)

	var consensusMsgs []ConsensusMsg

	for i := 0; i < 666; i++ {
		consensusMsgs = append(consensusMsgs, &testtypes.SimpleMessage{
			Sender: fmt.Sprintf("sender_%d", i),
		})
	}

	t.Run("putting messages in", func(t *testing.T) {
		err := cq.Put(ctx, consensusMsgs...)
		assert.NoError(t, err)
	})

	t.Run("without calling ProcessBatch", func(t *testing.T) {
		t.Run("GetAll should return nothing", func(t *testing.T) {
			msgs, err := cq.GetAll(ctx)
			require.NoError(t, err)
			require.Empty(t, msgs)
		})
	})

	t.Run("after calling ProcessBatch", func(t *testing.T) {
		err := cq.ProcessBatches(ctx)
		require.NoError(t, err)
		t.Run("GetAll should return 666/consensusQueueMaxBatchSize=7 batches", func(t *testing.T) {
			msgs, err := cq.GetAll(ctx)
			require.NoError(t, err)
			require.Len(t, msgs, 7)
		})
		t.Run("calling ProcessBatch shouldn't do anything", func(t *testing.T) {
			err := cq.ProcessBatches(ctx)
			require.NoError(t, err)
			t.Run("still 7 results", func(t *testing.T) {
				msgs, err := cq.GetAll(ctx)
				require.NoError(t, err)
				require.Len(t, msgs, 7)
			})
		})
		t.Run("verifying the bytes to sign", func(t *testing.T) {
			msgs, err := cq.GetAll(ctx)
			require.NoError(t, err)
			for _, msg := range msgs {
				require.Equal(t, msg.GetBytesToSign(), []byte("hello"))
			}
		})
		t.Run("removing all removes items", func(t *testing.T) {
			msgs, err := cq.GetAll(ctx)
			require.NoError(t, err)
			for _, msg := range msgs {
				cq.Remove(ctx, msg.GetId())
			}
			msgs, err = cq.GetAll(ctx)
			require.NoError(t, err)
			require.Empty(t, msgs)
		})
	})

}
