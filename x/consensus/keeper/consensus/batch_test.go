package consensus

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmdb "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/v2/util/keeper"
	"github.com/palomachain/paloma/v2/x/consensus/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatching(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	assert.NoError(t, stateStore.LoadLatestVersion())

	registry := types.ModuleCdc.InterfaceRegistry()
	registry.RegisterInterface(
		"palomachain.tests.SimpleMessage",
		(*types.ConsensusMsg)(nil),
		&types.SimpleMessage{},
	)
	registry.RegisterImplementations(
		(*types.ConsensusMsg)(nil),
		&types.SimpleMessage{},
	)
	types.RegisterInterfaces(registry)

	sg := keeperutil.SimpleStoreGetter(stateStore.GetKVStore(storeKey))
	msgType := &types.SimpleMessage{}
	cq, _ := NewBatchQueue(
		QueueOptions{
			QueueTypeName: "simple-message",
			Sg:            sg,
			Ider:          keeperutil.NewIDGenerator(sg, nil),
			Cdc:           types.ModuleCdc,
			TypeCheck:     types.StaticTypeChecker(msgType),
			VerifySignature: func([]byte, []byte, []byte) bool {
				return true
			},
			ChainType:        types.ChainTypeCosmos,
			ChainReferenceID: "test",
		})
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	t.Run("putting messages in", func(t *testing.T) {
		for i := 0; i < 666; i++ {
			consensusMsg := &types.SimpleMessage{
				Sender: fmt.Sprintf("sender_%d", i),
			}

			_, err := cq.Put(ctx, consensusMsg, nil)
			assert.NoError(t, err)
		}
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
		// TODO BytesToSign are never assigned to Batch messages
		// t.Run("verifying the bytes to sign", func(t *testing.T) {
		// 	msgs, err := cq.GetAll(ctx)
		// 	require.NoError(t, err)
		// 	for _, msg := range msgs {
		// 		bytesToSign, err := msg.GetBytesToSign(cq.base.qo.Cdc)
		// 		require.NoError(t, err)
		// 		require.Equal(t, bytesToSign, []byte("hello"))
		// 	}
		// })
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
