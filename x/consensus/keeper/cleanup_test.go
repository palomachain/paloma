package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteOldMessages(t *testing.T) {
	testcases := []struct {
		name          string
		blocksAgo     int64
		setup         func() (Keeper, sdk.Context, string)
		expectedError error
	}{
		{
			name:      "deletes messages over a number of blocks ago, but not messages under",
			blocksAgo: 50,
			setup: func() (Keeper, sdk.Context, string) {
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

				// Add a message at block 1
				ctx = ctx.WithBlockHeight(1)
				_, err := k.PutMessageInQueue(ctx, queue, &types.SimpleMessage{}, nil)
				require.NoError(t, err)

				// Add a message at block 10
				ctx = ctx.WithBlockHeight(10)
				_, err = k.PutMessageInQueue(ctx, queue, &types.SimpleMessage{}, nil)
				require.NoError(t, err)

				// Add a message at block 15
				ctx = ctx.WithBlockHeight(15)
				_, err = k.PutMessageInQueue(ctx, queue, &types.SimpleMessage{}, nil)
				require.NoError(t, err)

				// Add a message at block 20
				ctx = ctx.WithBlockHeight(20)
				_, err = k.PutMessageInQueue(ctx, queue, &types.SimpleMessage{}, nil)
				require.NoError(t, err)

				// Advance to block 61
				ctx = ctx.WithBlockHeight(61)

				return *k, ctx, queue
			},
			expectedError: nil,
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx, queue := tt.setup()

			msgsInQueue, err := k.GetMessagesFromQueue(ctx, queue, 50)
			require.NoError(t, err)

			asserter.Equal(4, len(msgsInQueue))

			actualErr := k.PruneOldMessages(ctx, tt.blocksAgo)

			asserter.Equal(tt.expectedError, actualErr)

			msgsInQueue, err = k.GetMessagesFromQueue(ctx, queue, 50)
			require.NoError(t, err)
			asserter.Equal(2, len(msgsInQueue))
		})
	}
}

func TestGetMessagesOlderThan(t *testing.T) {
	testcases := []struct {
		name          string
		blocksAgo     int64
		setup         func() (Keeper, sdk.Context, string)
		expectedError error
	}{
		{
			name:      "get messages over a number of blocks ago, but not messages under",
			blocksAgo: 50,
			setup: func() (Keeper, sdk.Context, string) {
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

				// Add a message at block 1
				ctx = ctx.WithBlockHeight(1)
				_, err := k.PutMessageInQueue(ctx, queue, &types.SimpleMessage{}, nil)
				require.NoError(t, err)

				// Add a message at block 10
				ctx = ctx.WithBlockHeight(10)
				_, err = k.PutMessageInQueue(ctx, queue, &types.SimpleMessage{}, nil)
				require.NoError(t, err)

				// Add a message at block 15
				ctx = ctx.WithBlockHeight(15)
				_, err = k.PutMessageInQueue(ctx, queue, &types.SimpleMessage{}, nil)
				require.NoError(t, err)

				// Add a message at block 20
				ctx = ctx.WithBlockHeight(20)
				_, err = k.PutMessageInQueue(ctx, queue, &types.SimpleMessage{}, nil)
				require.NoError(t, err)

				// Advance to block 61
				ctx = ctx.WithBlockHeight(61)

				return *k, ctx, queue
			},
			expectedError: nil,
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx, queue := tt.setup()

			msgsInQueue, err := k.GetMessagesFromQueue(ctx, queue, 50)
			require.NoError(t, err)

			asserter.Equal(4, len(msgsInQueue))

			msgs, actualErr := k.getMessagesOlderThan(ctx, queue, tt.blocksAgo)

			asserter.Equal(tt.expectedError, actualErr)
			asserter.Equal(2, len(msgs))
		})
	}
}
