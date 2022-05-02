package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testdata "github.com/palomachain/paloma/x/consensus/testdata/types"
	"github.com/palomachain/paloma/x/consensus/types"
	signingutils "github.com/palomachain/utils/signing"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestEndToEndTestingOfPuttingAndGettingMessagesOfTheConsensusQueue(t *testing.T) {
	keeper, _, ctx := newConsensusKeeper(t)

	t.Run("it returns a message if type is not registered with the queue", func(t *testing.T) {
		err := keeper.PutMessageForSigning(ctx, "i don't exist", &testdata.SimpleMessage{
			Sender: "bob",
			Hello:  "hello",
			World:  "mars",
		})

		require.ErrorIs(t, err, ErrConsensusQueueNotImplemented)
	})

	AddConcencusQueueType[*testdata.SimpleMessage](keeper, "simple-message")

	t.Run("it returns no messages for signing", func(t *testing.T) {
		msgs, err := keeper.GetMessagesForSigning(
			ctx,
			"simple-message",
			sdk.ValAddress(`bob`),
		)
		require.NoError(t, err)
		require.Empty(t, msgs)
	})
	t.Run("it sucessfully puts message into the queue", func(t *testing.T) {
		err := keeper.PutMessageForSigning(ctx, "simple-message", &testdata.SimpleMessage{
			Sender: "bob",
			Hello:  "hello",
			World:  "mars",
		})

		require.NoError(t, err)
	})

	t.Run("it sucessfully gets message from the queue", func(t *testing.T) {
		msgs, err := keeper.GetMessagesForSigning(
			ctx,
			"simple-message",
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

func TestAddingSignatures(t *testing.T) {
	keeper, ms, ctx := newConsensusKeeper(t)

	types.ModuleCdc.InterfaceRegistry().RegisterImplementations((*sdk.Msg)(nil), &testdata.SimpleMessage{})

	AddConcencusQueueType[*testdata.SimpleMessage](keeper, "simple-message")

	err := keeper.PutMessageForSigning(ctx, "simple-message", &testdata.SimpleMessage{
		Sender: "bob",
		Hello:  "hello",
		World:  "mars",
	})
	require.NoError(t, err)

	msgs, err := keeper.GetMessagesForSigning(ctx, "simple-message", sdk.ValAddress("val1"))
	require.NoError(t, err)
	require.Len(t, msgs, 1)

	msg := msgs[0]
	msgToSign, err := msg.SdkMsg()
	require.NoError(t, err)

	signedBytes, _, err := signingutils.SignBytes(
		key1,
		signingutils.SerializeFnc(signingutils.JsonDeterministicEncoding),
		msgToSign,
		msg.Nonce(),
	)
	require.NoError(t, err)

	t.Run("with incorrect key it returns an error", func(t *testing.T) {
		ms.ValsetKeeper.On("GetSigningKey", ctx, sdk.ValAddress("val1")).Return(key2.PubKey()).Once()
		err = keeper.AddMessageSignature(ctx, sdk.ValAddress("val1"), []*types.MsgAddMessagesSignatures_MsgSignedMessage{
			{
				Id:            msg.GetId(),
				QueueTypeName: "simple-message",
				Signature:     signedBytes,
			},
		})
		require.ErrorIs(t, err, ErrSignatureVerificationFailed)
	})

	t.Run("with correct key it adds it to the store", func(t *testing.T) {
		ms.ValsetKeeper.On("GetSigningKey", ctx, sdk.ValAddress("val1")).Return(key1.PubKey()).Once()
		err = keeper.AddMessageSignature(ctx, sdk.ValAddress("val1"), []*types.MsgAddMessagesSignatures_MsgSignedMessage{
			{
				Id:            msg.GetId(),
				QueueTypeName: "simple-message",
				Signature:     signedBytes,
			},
		})
		require.NoError(t, err)

		t.Run("it is no longer available for signing for this validator", func(t *testing.T) {
			msgs, err := keeper.GetMessagesForSigning(ctx, "simple-message", sdk.ValAddress("val1"))
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
