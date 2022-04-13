package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	testdata "github.com/volumefi/cronchain/x/consensus/testdata/types"
)

func TestEndToEndTestingOfPuttingAndGettingMessagesOfTheConsensusQueue(t *testing.T) {
	keeper, ctx := newConsensusKeeper(t)

	t.Run("it returns a message if type is not registered with the queue", func(t *testing.T) {
		err := keeper.PutMessageForSigning(ctx, "i don't exist", &testdata.SimpleMessage{
			Sender: "bob",
			Hello:  "hello",
			World:  "mars",
		})

		assert.ErrorIs(t, err, ErrConsensusQueueNotImplemented)
	})

	AddConcencusQueueType[*testdata.SimpleMessage](keeper, "simple-message")

	t.Run("it returns no messages for signing", func(t *testing.T) {
		msgs, err := keeper.GetMessagesForSigning(
			ctx,
			"simple-message",
			sdk.ValAddress(`bob`),
		)
		assert.NoError(t, err)
		assert.Empty(t, msgs)
	})
	t.Run("it sucessfully puts message into the queue", func(t *testing.T) {
		err := keeper.PutMessageForSigning(ctx, "simple-message", &testdata.SimpleMessage{
			Sender: "bob",
			Hello:  "hello",
			World:  "mars",
		})

		assert.NoError(t, err)
	})

	t.Run("it sucessfully gets message from the queue", func(t *testing.T) {
		msgs, err := keeper.GetMessagesForSigning(
			ctx,
			"simple-message",
			sdk.ValAddress(`bob`),
		)
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)
		t.Run("it then sends a signature for the message", func(t *testing.T) {
			// TODO: test this once we've implemented message signature adding
			// keeper.AddMessageSignature()
		})
	})
}
