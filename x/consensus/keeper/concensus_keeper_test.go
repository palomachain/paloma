package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensusmock "github.com/palomachain/paloma/x/consensus/keeper/consensus/mocks"
	"github.com/palomachain/paloma/x/consensus/types"
	consensustypesmock "github.com/palomachain/paloma/x/consensus/types/mocks"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/vizualni/whoops"
)

const (
	simpleMessageQueue = types.ConsensusQueueType("simple-message")
	chainType          = types.ChainTypeEVM
	chainID            = "test"
)

func TestEndToEndTestingOfPuttingAndGettingMessagesOfTheConsensusQueue(t *testing.T) {
	queue := types.Queue(simpleMessageQueue, chainType, chainID)
	keeper, _, ctx := newConsensusKeeper(t)

	t.Run("it returns a message if type is not registered with the queue", func(t *testing.T) {
		err := keeper.PutMessageForSigning(ctx, "i don't exist", &types.SimpleMessage{
			Sender: "bob",
			Hello:  "hello",
			World:  "mars",
		})

		require.ErrorIs(t, err, ErrConsensusQueueNotImplemented)
	})
	msgType := &types.SimpleMessage{}

	keeper.AddConcencusQueueType(
		false,
		consensus.WithQueueTypeName(simpleMessageQueue),
		consensus.WithStaticTypeCheck(msgType),
		consensus.WithBytesToSignCalc(msgType.ConsensusSignBytes()),
		consensus.WithChainInfo(chainType, chainID),
		consensus.WithVerifySignature(func([]byte, []byte, []byte) bool {
			return true
		}),
	)

	t.Run("it returns no messages for signing", func(t *testing.T) {
		msgs, err := keeper.GetMessagesForSigning(
			ctx,
			queue,
			sdk.ValAddress(`bob`),
		)
		require.NoError(t, err)
		require.Empty(t, msgs)
	})
	t.Run("it sucessfully puts message into the queue", func(t *testing.T) {
		err := keeper.PutMessageForSigning(ctx, queue, &types.SimpleMessage{
			Sender: "bob",
			Hello:  "hello",
			World:  "mars",
		})

		require.NoError(t, err)
	})

	t.Run("it sucessfully gets message from the queue", func(t *testing.T) {
		msgs, err := keeper.GetMessagesForSigning(
			ctx,
			queue,
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
		cq     *consensusmock.Queuer
	}

	for _, tt := range []struct {
		name          string
		queueTypeName string
		preRun        func(*testing.T, setupData)

		expMsgsLen int
		expErr     error
	}{
		{
			name:          "if consensusQueuer returns an error it returns it back",
			queueTypeName: "i don't exist",
			expErr:        ErrConsensusQueueNotImplemented,
		},
		{
			name: "if there are no signatures, no message is returned",
			preRun: func(t *testing.T, sd setupData) {
				sd.cq.On("GetAll", mock.Anything).Return(nil, nil).Once()
			},
		},
		{
			name:   "if consensus queue returns an error it returns it back",
			expErr: fakeErr,
			preRun: func(t *testing.T, sd setupData) {
				sd.cq.On("GetAll", mock.Anything).Return(nil, fakeErr).Once()
			},
		},
		{
			name: "if there is a message but snapshot does not exist",
			preRun: func(t *testing.T, sd setupData) {
				msg := consensustypesmock.NewQueuedSignedMessageI(t)
				sd.cq.On("GetAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(&valsettypes.Snapshot{}, nil)
			},
		},
		{
			name: "with messages returned but no signature data it returns nothing",
			preRun: func(t *testing.T, sd setupData) {
				msg := consensustypesmock.NewQueuedSignedMessageI(t)
				msg.On("GetSignData").Return(nil).Once()
				sd.cq.On("GetAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
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
				msg := consensustypesmock.NewQueuedSignedMessageI(t)
				msg.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("val1"), // val1 has only 2 shares
					},
				}).Once()
				sd.cq.On("GetAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
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
				msg := consensustypesmock.NewQueuedSignedMessageI(t)
				msg.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("val3"),
					},
					{
						ValAddress: sdk.ValAddress("val4"),
					},
				}).Once()
				sd.cq.On("GetAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
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
				msg1 := consensustypesmock.NewQueuedSignedMessageI(t)
				msg1.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("val4"),
					},
				}).Once()
				msg2 := consensustypesmock.NewQueuedSignedMessageI(t)
				msg2.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("val3"),
					},
					{
						ValAddress: sdk.ValAddress("val4"),
					},
				}).Once()
				sd.cq.On("GetAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg1, msg2}, nil).Once()
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
				msg1 := consensustypesmock.NewQueuedSignedMessageI(t)
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
				msg2 := consensustypesmock.NewQueuedSignedMessageI(t)
				msg2.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("val3"),
					},
					{
						ValAddress: sdk.ValAddress("val4"),
					},
				}).Once()
				sd.cq.On("GetAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg1, msg2}, nil).Once()
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
				msg := consensustypesmock.NewQueuedSignedMessageI(t)
				msg.On("GetSignData").Return([]*types.SignData{
					{
						ValAddress: sdk.ValAddress("i don't exist"),
					},
				}).Once()
				sd.cq.On("GetAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
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
			cq := consensusmock.NewQueuer(t)
			keeper.queueRegistry["simple-message"] = cq
			if tt.preRun != nil {
				tt.preRun(t, setupData{
					keeper, ms, cq,
				})
			}
			queueTypeName := "simple-message"
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
	queue := types.Queue(simpleMessageQueue, chainType, chainID)
	keeper, ms, ctx := newConsensusKeeper(t)

	types.ModuleCdc.InterfaceRegistry().RegisterImplementations((*types.ConsensusMsg)(nil), &types.SimpleMessage{})

	var msgType *types.SimpleMessage

	keeper.AddConcencusQueueType(
		false,
		consensus.WithQueueTypeName("simple-message"),
		consensus.WithStaticTypeCheck(msgType),
		consensus.WithBytesToSignCalc(msgType.ConsensusSignBytes()),
		consensus.WithChainInfo(chainType, chainID),
		consensus.WithVerifySignature(func(msg []byte, sig []byte, pk []byte) bool {
			p := secp256k1.PubKey(pk)
			return p.VerifySignature(msg, sig)
		}),
	)

	err := keeper.PutMessageForSigning(ctx, queue, &types.SimpleMessage{
		Sender: "bob",
		Hello:  "hello",
		World:  "mars",
	})
	require.NoError(t, err)
	val1 := sdk.ValAddress("val1")

	msgs, err := keeper.GetMessagesForSigning(ctx, queue, val1)
	require.NoError(t, err)
	require.Len(t, msgs, 1)

	msg := msgs[0]
	signedBytes, err := key1.Sign(msg.GetBytesToSign())

	require.NoError(t, err)

	t.Run("with incorrect key it returns an error", func(t *testing.T) {
		ms.ValsetKeeper.On("GetSigningKey", ctx, val1, chainType, chainID).Return(
			key2.PubKey().Bytes(),
			nil,
		).Once()
		err = keeper.AddMessageSignature(ctx, val1, []*types.MsgAddMessagesSignatures_MsgSignedMessage{
			{
				Id:            msg.GetId(),
				QueueTypeName: queue,
				Signature:     signedBytes,
			},
		})
		require.ErrorIs(t, err, consensus.ErrInvalidSignature)
	})

	t.Run("with correct key it adds it to the store", func(t *testing.T) {
		ms.ValsetKeeper.On("GetSigningKey", ctx, val1, chainType, chainID).Return(
			key1.PubKey().Bytes(),
			nil,
		).Once()
		err = keeper.AddMessageSignature(ctx, val1, []*types.MsgAddMessagesSignatures_MsgSignedMessage{
			{
				Id:            msg.GetId(),
				QueueTypeName: queue,
				Signature:     signedBytes,
			},
		})
		require.NoError(t, err)

		t.Run("it is no longer available for signing for this validator", func(t *testing.T) {
			msgs, err := keeper.GetMessagesForSigning(ctx, queue, val1)
			require.NoError(t, err)
			require.Empty(t, msgs)
		})
		t.Run("it is available for other validator", func(t *testing.T) {
			msgs, err := keeper.GetMessagesForSigning(ctx, queue, sdk.ValAddress("404"))
			require.NoError(t, err)
			require.Len(t, msgs, 1)
		})

	})
}
