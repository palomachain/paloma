package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	testtypes "github.com/palomachain/paloma/x/consensus/testdata/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/stretchr/testify/assert"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

func TestConsensusQueueAllMethods(t *testing.T) {
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
		(*sdk.Msg)(nil),
		&testtypes.SimpleMessage{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil), &testtypes.SimpleMessage{})
	types.RegisterInterfaces(registry)

	sg := keeperutil.SimpleStoreGetter(stateStore.GetKVStore(storeKey))
	cq := consensusQueue[*testtypes.SimpleMessage]{
		queueTypeName: "simple-message",
		sg:            sg,
		ider:          keeperutil.NewIDGenerator(sg, nil),
		cdc:           types.ModuleCdc,
	}
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, nil)

	msg := &testtypes.SimpleMessage{
		Sender: "bob",
		Hello:  "HEY",
		World:  "WORLD",
	}

	var msgs []types.QueuedSignedMessageI

	t.Run("putting message", func(t *testing.T) {
		err := cq.put(ctx, msg)
		assert.NoError(t, err)
	})

	t.Run("getting all messages should return one", func(t *testing.T) {
		var err error
		msgs, err = cq.getAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)
		assert.Len(t, msgs[0].GetSignData(), 0)
		assert.Positive(t, msgs[0].GetId())

		// lets see if it's equal to what we actually put in the queue
		realMsg := msgs[0]
		sdkMsg, err := realMsg.SdkMsg()
		assert.NoError(t, err)
		assert.Equal(t, msg, sdkMsg)
	})

	// lets add a signature to the message
	sig := &types.SignData{
		ValAddress: "bob",
		Signature:  []byte(`custom signature`),
	}

	cq.addSignature(ctx, msgs[0].GetId(), sig)

	t.Run("getting all messages should still return one", func(t *testing.T) {
		var err error
		msgs, err = cq.getAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)

		t.Run("there should be one signature only", func(t *testing.T) {
			// lets compare signatures
			signers := msgs[0].GetSignData()
			assert.Len(t, signers, 1)
			assert.Equal(t, sig, signers[0])
		})
	})

	t.Run("removing a message", func(t *testing.T) {
		err := cq.remove(ctx, msgs[0].GetId())
		assert.NoError(t, err)
		t.Run("getting all should return zero messages", func(t *testing.T) {
			msgs, err = cq.getAll(ctx)
			assert.NoError(t, err)
			assert.Len(t, msgs, 0)
		})

		t.Run("adding two new messages should add them", func(t *testing.T) {
			cq.put(
				ctx,
				&testtypes.SimpleMessage{},
				&testtypes.SimpleMessage{},
			)
			msgs, err = cq.getAll(ctx)
			assert.NoError(t, err)
			assert.Len(t, msgs, 2)
		})
	})

	t.Run("putting a message of a wrong type returns an error", func(t *testing.T) {
		msgOfWrongType := &testtypes.EvenSimplerMessage{
			Boo: "boo",
		}
		err := cq.put(ctx, msgOfWrongType)
		assert.ErrorIs(t, err, ErrIncorrectMessageType)
	})

	t.Run("saving without an ID raises an error", func(t *testing.T) {
		err := cq.save(ctx, &types.QueuedSignedMessage{})
		assert.ErrorIs(t, err, ErrUnableToSaveMessageWithoutID)
	})

	t.Run("fetching a message that does not exist raises an error", func(t *testing.T) {
		_, err := cq.getMsgByID(ctx, 999999)
		assert.ErrorIs(t, err, ErrMessageDoesNotExist)
	})
}
