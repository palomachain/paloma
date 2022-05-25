package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testdata "github.com/palomachain/paloma/x/consensus/testdata/types"
	"github.com/palomachain/paloma/x/consensus/types"
	consensusmocks "github.com/palomachain/paloma/x/consensus/types/mocks"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	signingutils "github.com/palomachain/utils/signing"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/vizualni/whoops"
)

func TestEndToEndTestingOfPuttingAndGettingMessagesOfTheConsensusQueue(t *testing.T) {
	const simpleMessageQueue = types.ConsensusQueueType("simple-message")
	keeper, _, ctx := newConsensusKeeper(t)

	t.Run("it returns a message if type is not registered with the queue", func(t *testing.T) {
		err := keeper.PutMessageForSigning(ctx, "i don't exist", &testdata.SimpleMessage{
			Sender: "bob",
			Hello:  "hello",
			World:  "mars",
		})

		require.ErrorIs(t, err, ErrConsensusQueueNotImplemented)
	})

	keeper.AddConcencusQueueType(simpleMessageQueue, &testdata.SimpleMessage{})

	t.Run("it returns no messages for signing", func(t *testing.T) {
		msgs, err := keeper.GetMessagesForSigning(
			ctx,
			simpleMessageQueue,
			sdk.ValAddress(`bob`),
		)
		require.NoError(t, err)
		require.Empty(t, msgs)
	})
	t.Run("it sucessfully puts message into the queue", func(t *testing.T) {
		err := keeper.PutMessageForSigning(ctx, simpleMessageQueue, &testdata.SimpleMessage{
			Sender: "bob",
			Hello:  "hello",
			World:  "mars",
		})

		require.NoError(t, err)
	})

	t.Run("it sucessfully gets message from the queue", func(t *testing.T) {
		msgs, err := keeper.GetMessagesForSigning(
			ctx,
			simpleMessageQueue,
			sdk.ValAddress(`bob`),
		)
		require.NoError(t, err)
		require.Len(t, msgs, 1)
		t.Run("it then sends a signature for the message", func(t *testing.T) {
			// TODO: test this once we've implemented message signature adding
			// keeper.AddMessageSignature()
		})
	})
}

var (
	key1 = secp256k1.GenPrivKey()
	key2 = secp256k1.GenPrivKey()
)

func TestGettingMessagesThatHaveReachedConsensus(t *testing.T) {
	testValidators := []valsettypes.Validator{
		{
			ShareCount: sdk.NewInt(2),
			Address:    sdk.ValAddress("val1"),
		},
		{
			ShareCount: sdk.NewInt(3),
			Address:    sdk.ValAddress("val2"),
		},
		{
			ShareCount: sdk.NewInt(6),
			Address:    sdk.ValAddress("val3"),
		},
		{
			ShareCount: sdk.NewInt(7),
			Address:    sdk.ValAddress("val4"),
		},
	}
	total := sdk.NewInt(18)

	_ = testValidators
	_ = total

	fakeErr := whoops.String("oh no")

	type setupData struct {
		keeper *Keeper
		ms     mockedServices
		cq     *mockConsensusQueuer
	}

	for _, tt := range []struct {
		name          string
		queueTypeName types.ConsensusQueueType
		preRun        func(*testing.T, setupData)

		expMsgsLen int
		expErr     error
	}{
		{
			name:          "if consensusQueuer returns an error it returns it back",
			queueTypeName: types.ConsensusQueueType("i don't exist"),
			expErr:        ErrConsensusQueueNotImplemented,
		},
		{
			name: "if there are no signatures, no message is returned",
			preRun: func(t *testing.T, sd setupData) {
				sd.cq.On("getAll", mock.Anything).Return(nil, nil).Once()
			},
		},
		{
			name:   "if consensus queue returns an error it returns it back",
			expErr: fakeErr,
			preRun: func(t *testing.T, sd setupData) {
				sd.cq.On("getAll", mock.Anything).Return(nil, fakeErr).Once()
			},
		},
		{
			name: "if there is a message but snapshot does not exist",
			preRun: func(t *testing.T, sd setupData) {
				msg := consensusmocks.NewQueuedSignedMessageI(t)
				sd.cq.On("getAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(&valsettypes.Snapshot{}, nil)
			},
		},
		{
			name: "with messages returned but no signature data it returns nothing",
			preRun: func(t *testing.T, sd setupData) {
				msg := consensusmocks.NewQueuedSignedMessageI(t)
				msg.On("GetSignData").Return(nil).Once()
				sd.cq.On("getAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(
					&valsettypes.Snapshot{
						TotalShares: total,
						Validators:  testValidators,
					},
					nil,
				)
			},
		},
		{
			name: "with a single signature only which is not enough it returns nothing",
			preRun: func(t *testing.T, sd setupData) {
				msg := consensusmocks.NewQueuedSignedMessageI(t)
				msg.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("val1"), // val1 has only 2 shares
					},
				}).Once()
				sd.cq.On("getAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(
					&valsettypes.Snapshot{
						TotalShares: total,
						Validators:  testValidators,
					},
					nil,
				)
			},
		},
		{
			name:       "with enough signatures for a consensus it returns messages",
			expMsgsLen: 1,
			preRun: func(t *testing.T, sd setupData) {
				msg := consensusmocks.NewQueuedSignedMessageI(t)
				msg.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("val3"),
					},
					{
						ValAddress: sdk.ValAddress("val4"),
					},
				}).Once()
				sd.cq.On("getAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(
					&valsettypes.Snapshot{
						TotalShares: total,
						Validators:  testValidators,
					},
					nil,
				)
			},
		},
		{
			name:       "with multiple messages where only one has enough signatures",
			expMsgsLen: 1,
			preRun: func(t *testing.T, sd setupData) {
				msg1 := consensusmocks.NewQueuedSignedMessageI(t)
				msg1.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("val4"),
					},
				}).Once()
				msg2 := consensusmocks.NewQueuedSignedMessageI(t)
				msg2.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("val3"),
					},
					{
						ValAddress: sdk.ValAddress("val4"),
					},
				}).Once()
				sd.cq.On("getAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg1, msg2}, nil).Once()
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(
					&valsettypes.Snapshot{
						TotalShares: total,
						Validators:  testValidators,
					},
					nil,
				)
			},
		},
		{
			name:       "with multiple messages where all have enough signatures",
			expMsgsLen: 2,
			preRun: func(t *testing.T, sd setupData) {
				msg1 := consensusmocks.NewQueuedSignedMessageI(t)
				msg1.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("val2"),
					},
					{
						ValAddress: sdk.ValAddress("val3"),
					},
					{
						ValAddress: sdk.ValAddress("val4"),
					},
				}).Once()
				msg2 := consensusmocks.NewQueuedSignedMessageI(t)
				msg2.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("val3"),
					},
					{
						ValAddress: sdk.ValAddress("val4"),
					},
				}).Once()
				sd.cq.On("getAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg1, msg2}, nil).Once()
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(
					&valsettypes.Snapshot{
						TotalShares: total,
						Validators:  testValidators,
					},
					nil,
				)
			},
		},
		{
			name: "if it's signed by a validator which is not in the snapshot it skips it",
			preRun: func(t *testing.T, sd setupData) {
				msg := consensusmocks.NewQueuedSignedMessageI(t)
				msg.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("i don't exist"),
					},
				}).Once()
				sd.cq.On("getAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(
					&valsettypes.Snapshot{
						TotalShares: total,
						Validators:  testValidators,
					},
					nil,
				)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ms, ctx := newConsensusKeeper(t)
			cq := newMockConsensusQueuer(t)
			keeper.queueRegistry["simple-message"] = cq
			if tt.preRun != nil {
				tt.preRun(t, setupData{
					keeper, ms, cq,
				})
			}
			queueTypeName := types.ConsensusQueueType("simple-message")
			if len(tt.queueTypeName) > 0 {
				queueTypeName = tt.queueTypeName
			}
			msgs, err := keeper.GetMessagesThatHaveReachedConsensus(ctx, queueTypeName)
			require.ErrorIs(t, err, tt.expErr)
			require.Len(t, msgs, tt.expMsgsLen)
		})
	}

}

func TestAddingSignatures(t *testing.T) {
	keeper, ms, ctx := newConsensusKeeper(t)

	types.ModuleCdc.InterfaceRegistry().RegisterImplementations((*types.ConsensusMsg)(nil), &testdata.SimpleMessage{})

	keeper.AddConcencusQueueType("simple-message", &testdata.SimpleMessage{})

	err := keeper.PutMessageForSigning(ctx, "simple-message", &testdata.SimpleMessage{
		Sender: "bob",
		Hello:  "hello",
		World:  "mars",
	})
	require.NoError(t, err)
	val1 := sdk.ValAddress("val1")

	msgs, err := keeper.GetMessagesForSigning(ctx, "simple-message", val1)
	require.NoError(t, err)
	require.Len(t, msgs, 1)

	msg := msgs[0]
	msgToSign, err := msg.ConsensusMsg()
	require.NoError(t, err)

	signedBytes, _, err := signingutils.SignBytes(
		key1,
		signingutils.SerializeFnc(signingutils.JsonDeterministicEncoding),
		msgToSign,
		msg.Nonce(),
	)
	require.NoError(t, err)

	t.Run("with incorrect key it returns an error", func(t *testing.T) {
		ms.ValsetKeeper.On("GetSigningKey", ctx, val1).Return(key2.PubKey()).Once()
		err = keeper.AddMessageSignature(ctx, val1, []*types.MsgAddMessagesSignatures_MsgSignedMessage{
			{
				Id:            msg.GetId(),
				QueueTypeName: "simple-message",
				Signature:     signedBytes,
			},
		})
		require.ErrorIs(t, err, ErrSignatureVerificationFailed)
	})

	t.Run("with correct key it adds it to the store", func(t *testing.T) {
		ms.ValsetKeeper.On("GetSigningKey", ctx, val1).Return(key1.PubKey()).Once()
		err = keeper.AddMessageSignature(ctx, val1, []*types.MsgAddMessagesSignatures_MsgSignedMessage{
			{
				Id:            msg.GetId(),
				QueueTypeName: "simple-message",
				Signature:     signedBytes,
			},
		})
		require.NoError(t, err)

		t.Run("it is no longer available for signing for this validator", func(t *testing.T) {
			msgs, err := keeper.GetMessagesForSigning(ctx, "simple-message", val1)
			require.NoError(t, err)
			require.Empty(t, msgs)
		})
		t.Run("it is available for other validator", func(t *testing.T) {
			msgs, err := keeper.GetMessagesForSigning(ctx, "simple-message", sdk.ValAddress("404"))
			require.NoError(t, err)
			require.Len(t, msgs, 1)
		})

	})
}
