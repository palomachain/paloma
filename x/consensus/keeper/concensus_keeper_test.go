package keeper

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensusmock "github.com/palomachain/paloma/x/consensus/keeper/consensus/mocks"
	"github.com/palomachain/paloma/x/consensus/types"
	consensustypemocks "github.com/palomachain/paloma/x/consensus/types/mocks"
	evmtypes "github.com/palomachain/paloma/x/evm/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
		_, err := keeper.PutMessageInQueue(ctx, "i don't exist", &types.SimpleMessage{
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
		_, err := keeper.PutMessageInQueue(ctx, queue, &types.SimpleMessage{
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

func TestGetMessagesFromQueue(t *testing.T) {
	testcases := []struct {
		name              string
		setupMessageCount int
		inputCount        int
		expectedCount     int
	}{
		{
			name:              "returns all messages when requesting 0",
			inputCount:        0,
			setupMessageCount: 5000,
			expectedCount:     5000,
		},
		{
			name:              "returns all messages when requesting -1",
			inputCount:        -1,
			setupMessageCount: 5000,
			expectedCount:     5000,
		},
		{
			name:              "returns 5 messages when requesting 5",
			inputCount:        5,
			setupMessageCount: 50,
			expectedCount:     5,
		},
		{
			name:              "returns all messages messages when requesting more than total",
			inputCount:        500,
			setupMessageCount: 50,
			expectedCount:     50,
		},
	}
	addMessages := func(ctx sdk.Context, k Keeper, queue string, numMessages int) {
		for i := 0; i < numMessages; i++ {
			_, err := k.PutMessageInQueue(ctx, queue, &types.SimpleMessage{}, nil)
			require.NoError(t, err)
		}
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			k, _, ctx := newConsensusKeeper(t)
			queue := types.Queue(defaultQueueName, chainType, chainReferenceID)

			msgType := &types.SimpleMessage{}

			k.registry.Add(
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
			addMessages(ctx, *k, queue, tt.setupMessageCount)

			msgsInQueue, err := k.GetMessagesFromQueue(ctx, queue, tt.inputCount)
			require.NoError(t, err)
			asserter.Equal(tt.expectedCount, len(msgsInQueue))
		})
	}
}

func TestGetMessagesForRelaying(t *testing.T) {
	k, _, ctx := newConsensusKeeper(t)
	queue := types.Queue(defaultQueueName, chainType, chainReferenceID)
	queueWithValsetUpdatesPending := types.Queue(defaultQueueName, chainType, "pending-chain")

	uvType := &evmtypes.Message{}

	types.RegisterInterfaces(types.ModuleCdc.InterfaceRegistry())
	evmtypes.RegisterInterfaces(types.ModuleCdc.InterfaceRegistry())

	types.ModuleCdc.InterfaceRegistry().RegisterImplementations((*types.ConsensusMsg)(nil), &evmtypes.Message{})
	types.ModuleCdc.InterfaceRegistry().RegisterImplementations((*evmtypes.TurnstoneMsg)(nil), &evmtypes.Message{})

	k.registry.Add(
		queueSupporter{
			opt: consensus.ApplyOpts(nil,
				consensus.WithQueueTypeName(queue),
				consensus.WithStaticTypeCheck(uvType),
				consensus.WithBytesToSignCalc(func(msg types.ConsensusMsg, salt types.Salt) []byte { return []byte{} }),
				consensus.WithChainInfo(chainType, chainReferenceID),
				consensus.WithVerifySignature(func([]byte, []byte, []byte) bool {
					return true
				}),
			),
		},
		queueSupporter{
			opt: consensus.ApplyOpts(nil,
				consensus.WithQueueTypeName(queueWithValsetUpdatesPending),
				consensus.WithStaticTypeCheck(uvType),
				consensus.WithBytesToSignCalc(func(msg types.ConsensusMsg, salt types.Salt) []byte { return []byte{} }),
				consensus.WithChainInfo(chainType, "pending-chain"),
				consensus.WithVerifySignature(func([]byte, []byte, []byte) bool {
					return true
				}),
			),
		},
	)

	val := sdk.ValAddress("validator-1")

	// message for other validator
	_, err := k.PutMessageInQueue(ctx, queue, &evmtypes.Message{
		TurnstoneID:      "abc",
		ChainReferenceID: chainReferenceID,
		Assignee:         sdk.ValAddress("other-validator").String(),
	}, nil)
	require.NoError(t, err)

	// message for test validator
	_, err = k.PutMessageInQueue(ctx, queue, &evmtypes.Message{
		TurnstoneID:      "abc",
		ChainReferenceID: chainReferenceID,
		Assignee:         val.String(),
	}, nil)
	require.NoError(t, err)

	// message for test validator on other chain
	_, err = k.PutMessageInQueue(ctx, queueWithValsetUpdatesPending, &evmtypes.Message{
		TurnstoneID:      "abc",
		ChainReferenceID: "pending-chain",
		Assignee:         val.String(),
	}, nil)
	require.NoError(t, err)

	msgs, err := k.GetMessagesForRelaying(ctx, queue, val)
	require.NoError(t, err)
	require.Len(t, msgs, 1, "validator should get exactly 1 message")

	msgs, err = k.GetMessagesForRelaying(ctx, queueWithValsetUpdatesPending, val)
	require.NoError(t, err)
	require.Len(t, msgs, 1, "validator should get exactly 1 message")

	// update valset message on other chain
	_, err = k.PutMessageInQueue(ctx, queueWithValsetUpdatesPending, &evmtypes.Message{
		TurnstoneID:      "abc",
		ChainReferenceID: "pending-chain",
		Assignee:         val.String(),
		Action: &evmtypes.Message_UpdateValset{
			UpdateValset: &evmtypes.UpdateValset{
				Valset: &evmtypes.Valset{
					ValsetID: 777,
				},
			},
		},
	}, nil)
	require.NoError(t, err)

	msgs, err = k.GetMessagesForRelaying(ctx, queueWithValsetUpdatesPending, val)
	require.NoError(t, err)
	require.Empty(t, msgs, "validator should not get any messages due to blocking update valset message")
}

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
				_, err := sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
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
				_, err := sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
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
				_, err := sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
				_, err = sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
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
				_, err := sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
				require.NoError(t, err)
				_, err = sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
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
				_, err := sd.keeper.PutMessageInQueue(sd.ctx, defaultQueueName, msg, nil)
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

	_, err := keeper.PutMessageInQueue(ctx, queue, &types.SimpleMessage{
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
