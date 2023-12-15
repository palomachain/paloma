package consensus

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmdb "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/consensus/types"
	consensustypemocks "github.com/palomachain/paloma/x/consensus/types/mocks"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestConsensusQueueAllMethods(t *testing.T) {
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
	registry.RegisterImplementations((*types.ConsensusMsg)(nil), &types.SimpleMessage{})
	types.RegisterInterfaces(registry)

	sg := keeperutil.SimpleStoreGetter(stateStore.GetKVStore(storeKey))
	var msgType *types.SimpleMessage
	mck := consensustypemocks.NewAttestator(t)
	mck.On("ValidateEvidence", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	cq := Queue{
		qo: QueueOptions{
			QueueTypeName:         "simple-message",
			Sg:                    sg,
			Ider:                  keeperutil.NewIDGenerator(sg, nil),
			Cdc:                   types.ModuleCdc,
			TypeCheck:             types.StaticTypeChecker(msgType),
			Attestator:            mck,
			BytesToSignCalculator: msgType.ConsensusSignBytes(),
			VerifySignature: func([]byte, []byte, []byte) bool {
				return true
			},
			ChainType:        types.ChainTypeEVM,
			ChainReferenceID: "bla",
		},
	}
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, nil)

	msg := &types.SimpleMessage{
		Sender: "bob",
		Hello:  "HEY",
		World:  "WORLD",
	}

	var msgs []types.QueuedSignedMessageI

	t.Run("putting message", func(t *testing.T) {
		_, err := cq.Put(ctx, msg, nil)
		assert.NoError(t, err)
	})

	t.Run("getting all messages should return one", func(t *testing.T) {
		var err error
		msgs, err = cq.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)
		assert.Len(t, msgs[0].GetSignData(), 0)
		assert.Positive(t, msgs[0].GetId())

		// lets see if it's equal to what we actually put in the queue
		realMsg := msgs[0]
		consensusMsg, err := realMsg.ConsensusMsg(registry)
		assert.NoError(t, err)
		assert.Equal(t, msg, consensusMsg)
	})

	// lets add a signature to the message
	sig := &types.SignData{
		ValAddress: sdk.ValAddress("bob"),
		Signature:  []byte(`custom signature`),
		PublicKey:  []byte("it does not matter"),
	}

	cq.AddSignature(ctx, msgs[0].GetId(), sig)

	t.Run("getting all messages should still return one", func(t *testing.T) {
		var err error
		msgs, err = cq.GetAll(ctx)
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
		err := cq.Remove(ctx, msgs[0].GetId())
		assert.NoError(t, err)
		t.Run("getting all should return zero messages", func(t *testing.T) {
			msgs, err = cq.GetAll(ctx)
			assert.NoError(t, err)
			assert.Len(t, msgs, 0)
		})

		t.Run("adding two new messages should add them", func(t *testing.T) {
			cq.Put(
				ctx,
				&types.SimpleMessage{},
				nil,
			)
			cq.Put(
				ctx,
				&types.SimpleMessage{},
				nil,
			)
			msgs, err = cq.GetAll(ctx)
			assert.NoError(t, err)
			assert.Len(t, msgs, 2)
		})
	})

	t.Run("putting a message of a wrong type returns an error", func(t *testing.T) {
		msgOfWrongType := &types.EvenSimplerMessage{
			Boo: "boo",
		}
		_, err := cq.Put(ctx, msgOfWrongType, nil)
		assert.ErrorIs(t, err, ErrIncorrectMessageType)
	})

	t.Run("saving without an ID raises an error", func(t *testing.T) {
		err := cq.save(ctx, &types.QueuedSignedMessage{})
		assert.ErrorIs(t, err, ErrUnableToSaveMessageWithoutID)
	})

	t.Run("fetching a message that does not exist raises an error", func(t *testing.T) {
		_, err := cq.GetMsgByID(ctx, 999999)
		assert.ErrorIs(t, err, ErrMessageDoesNotExist)
	})
}
