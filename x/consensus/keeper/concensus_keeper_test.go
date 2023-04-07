package keeper

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensusmock "github.com/palomachain/paloma/x/consensus/keeper/consensus/mocks"
	"github.com/palomachain/paloma/x/consensus/types"
	consensustypemocks "github.com/palomachain/paloma/x/consensus/types/mocks"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

const (
	chainType        = types.ChainTypeEVM
	chainReferenceID = "test"
	defaultQueueName = "simple-message"
)

func TestEndToEndTestingOfPuttingAndGettingMessagesOfTheConsensusQueue(t *testing.T) {
	queue := types.Queue(defaultQueueName, chainType, chainReferenceID)
	keeper, _, ctx := newConsensusKeeper(t)

	t.Run("it returns a message if type is not registered with the queue", func(t *testing.T) {
		err := keeper.PutMessageInQueue(ctx, "i don't exist", &types.SimpleMessage{
			Sender: "bob",
			Hello:  "hello",
			World:  "mars",
		}, nil)

		require.ErrorIs(t, err, ErrConsensusQueueNotImplemented)
	})
	msgType := &types.SimpleMessage{}

	keeper.registry.Add(
		queueSupporter{
			opt: consensus.ApplyOpts(nil,
				consensus.WithQueueTypeName(queue),
				consensus.WithStaticTypeCheck(msgType),
				consensus.WithBytesToSignCalc(msgType.ConsensusSignBytes()),
				consensus.WithChainInfo(chainType, chainReferenceID),
				consensus.WithVerifySignature(func([]byte, []byte, []byte) bool {
					return true
				}),
			),
		},
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

	t.Run("it successfully puts message into the queue", func(t *testing.T) {
		err := keeper.PutMessageInQueue(ctx, queue, &types.SimpleMessage{
			Sender: "bob",
			Hello:  "hello",
			World:  "mars",
		}, nil)

		require.NoError(t, err)
	})

	t.Run("it successfully gets message from the queue", func(t *testing.T) {
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

	type setupData struct {
		keeper *Keeper
		ms     mockedServices
		cq     *consensusmock.Queuer
		ctx    sdk.Context
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
			queueTypeName: "nope",
			expErr:        ErrConsensusQueueNotImplemented,
		},
		{
			name: "if there are no signatures, no message is returned",
			preRun: func(t *testing.T, sd setupData) {
				// sd.cq.On("GetAll", mock.Anything).Return(nil, nil).Once()
			},
		},
		{
			name: "if there is a message but snapshot does not exist",
			preRun: func(t *testing.T, sd setupData) {
				msg := &types.SimpleMessage{}
				// sd.cq.On("GetAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
				sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(&valsettypes.Snapshot{}, nil)
			},
		},
		{
			name: "with messages returned but no signature data it returns nothing",
			preRun: func(t *testing.T, sd setupData) {
				msg := &types.SimpleMessage{}
				sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
				// msg := consensustypesmock.NewQueuedSignedMessageI(t)
				// msg.On("GetSignData").Return(nil).Once()
				// sd.cq.On("GetAll", mock.Anything).Return([]types.QueuedSignedMessageI{msg}, nil).Once()
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
				msg := &types.SimpleMessage{}
				err := sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
				require.NoError(t, err)
				sd.ms.ValsetKeeper.On("GetSigningKey", mock.Anything, sdk.ValAddress("val1"), "evm", "test", "bob").Return(
					[]byte("signing-key"),
					nil,
				)
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(
					&valsettypes.Snapshot{
						TotalShares: total,
						Validators:  testValidators,
					},
					nil,
				)
				err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("val1"), []*types.ConsensusMessageSignature{
					{
						Id:              1,
						QueueTypeName:   defaultQueueName,
						Signature:       []byte("abc"),
						SignedByAddress: "bob",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name:       "with enough signatures for a consensus it returns messages",
			expMsgsLen: 1,
			preRun: func(t *testing.T, sd setupData) {
				msg := &types.SimpleMessage{}
				err := sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
				require.NoError(t, err)

				sd.ms.ValsetKeeper.On("GetSigningKey", mock.Anything, sdk.ValAddress("val3"), "evm", "test", "bob3").Return(
					[]byte("signing-key-1"),
					nil,
				)

				sd.ms.ValsetKeeper.On("GetSigningKey", mock.Anything, sdk.ValAddress("val4"), "evm", "test", "bob4").Return(
					[]byte("signing-key-2"),
					nil,
				)
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(
					&valsettypes.Snapshot{
						TotalShares: total,
						Validators:  testValidators,
					},
					nil,
				)
				err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("val3"), []*types.ConsensusMessageSignature{
					{
						Id:              1,
						QueueTypeName:   defaultQueueName,
						Signature:       []byte("abc"),
						SignedByAddress: "bob3",
					},
				})
				require.NoError(t, err)
				err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("val4"), []*types.ConsensusMessageSignature{
					{
						Id:              1,
						QueueTypeName:   defaultQueueName,
						Signature:       []byte("abc"),
						SignedByAddress: "bob4",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name:       "with multiple messages where only one has enough signatures",
			expMsgsLen: 1,
			preRun: func(t *testing.T, sd setupData) {
				msg := &types.SimpleMessage{}
				err := sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
				err = sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
				require.NoError(t, err)

				sd.ms.ValsetKeeper.On("GetSigningKey", mock.Anything, sdk.ValAddress("val3"), "evm", "test", "bob3").Return(
					[]byte("signing-key-1"),
					nil,
				)

				sd.ms.ValsetKeeper.On("GetSigningKey", mock.Anything, sdk.ValAddress("val4"), "evm", "test", "bob4").Return(
					[]byte("signing-key-2"),
					nil,
				)
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(
					&valsettypes.Snapshot{
						TotalShares: total,
						Validators:  testValidators,
					},
					nil,
				)
				t.Run("adding message signatures to first message which has enough signatures", func(t *testing.T) {
					err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("val3"), []*types.ConsensusMessageSignature{
						{
							Id:              1,
							QueueTypeName:   defaultQueueName,
							Signature:       []byte("abc"),
							SignedByAddress: "bob3",
						},
					})
					require.NoError(t, err)
					err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("val4"), []*types.ConsensusMessageSignature{
						{
							Id:              1,
							QueueTypeName:   defaultQueueName,
							Signature:       []byte("abc"),
							SignedByAddress: "bob4",
						},
					})
					require.NoError(t, err)
				})
				t.Run("adding message signatures to a second message doesn't have enough signatures", func(t *testing.T) {
					err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("val4"), []*types.ConsensusMessageSignature{
						{
							Id:              2,
							QueueTypeName:   defaultQueueName,
							Signature:       []byte("abc"),
							SignedByAddress: "bob4",
						},
					})
					require.NoError(t, err)
				})
			},
		},
		{
			name:       "with multiple messages where all have enough signatures",
			expMsgsLen: 2,
			preRun: func(t *testing.T, sd setupData) {
				msg := &types.SimpleMessage{}
				err := sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
				require.NoError(t, err)
				err = sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
				require.NoError(t, err)

				sd.ms.ValsetKeeper.On("GetSigningKey", mock.Anything, sdk.ValAddress("val2"), "evm", "test", "bob2").Return(
					[]byte("signing-key-2"),
					nil,
				)

				sd.ms.ValsetKeeper.On("GetSigningKey", mock.Anything, sdk.ValAddress("val3"), "evm", "test", "bob3").Return(
					[]byte("signing-key-3"),
					nil,
				)

				sd.ms.ValsetKeeper.On("GetSigningKey", mock.Anything, sdk.ValAddress("val4"), "evm", "test", "bob4").Return(
					[]byte("signing-key-4"),
					nil,
				)
				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(
					&valsettypes.Snapshot{
						TotalShares: total,
						Validators:  testValidators,
					},
					nil,
				)
				t.Run("adding message signatures to first message which has enough signatures", func(t *testing.T) {
					err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("val2"), []*types.ConsensusMessageSignature{
						{
							Id:              1,
							QueueTypeName:   defaultQueueName,
							Signature:       []byte("abc"),
							SignedByAddress: "bob2",
						},
					})
					require.NoError(t, err)
					err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("val3"), []*types.ConsensusMessageSignature{
						{
							Id:              1,
							QueueTypeName:   defaultQueueName,
							Signature:       []byte("abc"),
							SignedByAddress: "bob3",
						},
					})
					require.NoError(t, err)
					err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("val4"), []*types.ConsensusMessageSignature{
						{
							Id:              1,
							QueueTypeName:   defaultQueueName,
							Signature:       []byte("abc"),
							SignedByAddress: "bob4",
						},
					})
					require.NoError(t, err)
				})
				t.Run("adding message signatures to second message which has enough signatures", func(t *testing.T) {
					err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("val3"), []*types.ConsensusMessageSignature{
						{
							Id:              2,
							QueueTypeName:   defaultQueueName,
							Signature:       []byte("abc"),
							SignedByAddress: "bob3",
						},
					})
					require.NoError(t, err)
					err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("val4"), []*types.ConsensusMessageSignature{
						{
							Id:              2,
							QueueTypeName:   defaultQueueName,
							Signature:       []byte("abc"),
							SignedByAddress: "bob4",
						},
					})
					require.NoError(t, err)
				})
			},
		},
		{
			name: "if it's signed by a validator which is not in the snapshot it skips it",
			preRun: func(t *testing.T, sd setupData) {
				msg := &types.SimpleMessage{}
				err := sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
				require.NoError(t, err)

				sd.ms.ValsetKeeper.On("GetSigningKey", mock.Anything, sdk.ValAddress("404"), "evm", "test", "404").Return(
					[]byte("signing-key-1"),
					nil,
				)

				sd.ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(
					&valsettypes.Snapshot{
						TotalShares: total,
						Validators:  testValidators,
					},
					nil,
				)
				err = sd.keeper.AddMessageSignature(sd.ctx, sdk.ValAddress("404"), []*types.ConsensusMessageSignature{
					{
						Id:              1,
						QueueTypeName:   defaultQueueName,
						Signature:       []byte("abc"),
						SignedByAddress: "404",
					},
				})
				require.NoError(t, err)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ms, ctx := newConsensusKeeper(t)
			cq := consensusmock.NewQueuer(t)
			mck := consensustypemocks.NewAttestator(t)
			mck.On("ValidateEvidence", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			keeper.registry.Add(
				queueSupporter{
					opt: consensus.ApplyOpts(nil,
						consensus.WithAttestator(mck),
						consensus.WithStaticTypeCheck(&types.SimpleMessage{}),
						consensus.WithQueueTypeName(defaultQueueName),
						consensus.WithChainInfo(chainType, chainReferenceID),
						consensus.WithBytesToSignCalc(func(msg types.ConsensusMsg, salt types.Salt) []byte {
							return []byte("sign-me")
						}),
						consensus.WithVerifySignature(func([]byte, []byte, []byte) bool {
							return true
						}),
					),
				},
			)
			if tt.preRun != nil {
				tt.preRun(t, setupData{
					keeper, ms, cq, ctx,
				})
			}
			queueTypeName := defaultQueueName
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
	queue := types.Queue(defaultQueueName, chainType, chainReferenceID)
	keeper, ms, ctx := newConsensusKeeper(t)

	types.ModuleCdc.InterfaceRegistry().RegisterImplementations((*types.ConsensusMsg)(nil), &types.SimpleMessage{})

	var msgType *types.SimpleMessage

	mck := consensustypemocks.NewAttestator(t)
	mck.On("ValidateEvidence", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	keeper.registry.Add(
		queueSupporter{
			opt: consensus.ApplyOpts(nil,
				consensus.WithAttestator(mck),
				consensus.WithQueueTypeName(queue),
				consensus.WithStaticTypeCheck(msgType),
				consensus.WithBytesToSignCalc(msgType.ConsensusSignBytes()),
				consensus.WithChainInfo(chainType, chainReferenceID),
				consensus.WithVerifySignature(func(msg []byte, sig []byte, pk []byte) bool {
					p := secp256k1.PubKey(pk)
					return p.VerifySignature(msg, sig)
				}),
			),
		},
	)

	err := keeper.PutMessageInQueue(ctx, queue, &types.SimpleMessage{
		Sender: "bob",
		Hello:  "hello",
		World:  "mars",
	}, nil)
	require.NoError(t, err)
	val1 := sdk.ValAddress("val1")

	msgs, err := keeper.GetMessagesForSigning(ctx, queue, val1)
	require.NoError(t, err)
	require.Len(t, msgs, 1)

	msg := msgs[0]
	signedBytes, err := key1.Sign(msg.GetBytesToSign())

	require.NoError(t, err)

	t.Run("with incorrect key it returns an error", func(t *testing.T) {
		ms.ValsetKeeper.On("GetSigningKey", ctx, val1, chainType, chainReferenceID, "bob").Return(
			key2.PubKey().Bytes(),
			nil,
		).Once()
		err = keeper.AddMessageSignature(ctx, val1, []*types.ConsensusMessageSignature{
			{
				Id:              msg.GetId(),
				QueueTypeName:   queue,
				Signature:       signedBytes,
				SignedByAddress: "bob",
			},
		})
		require.ErrorIs(t, err, consensus.ErrInvalidSignature)
	})

	t.Run("with correct key it adds it to the store", func(t *testing.T) {
		ms.ValsetKeeper.On("GetSigningKey", ctx, val1, chainType, chainReferenceID, "bob").Return(
			key1.PubKey().Bytes(),
			nil,
		).Once()
		err = keeper.AddMessageSignature(ctx, val1, []*types.ConsensusMessageSignature{
			{
				Id:              msg.GetId(),
				QueueTypeName:   queue,
				Signature:       signedBytes,
				SignedByAddress: "bob",
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

type queueSupporter struct {
	opt *consensus.QueueOptions
}

func (q queueSupporter) SupportedQueues(ctx sdk.Context) ([]consensus.SupportsConsensusQueueAction, error) {
	return []consensus.SupportsConsensusQueueAction{
		{
			QueueOptions: *q.opt,
		},
	}, nil
}
