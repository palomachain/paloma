package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/consensus/types/mocks"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	signingutils "github.com/palomachain/utils/signing"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

const (
	simpleQueue = types.ConsensusQueueType("simple-message")
)

func TestAttesting(t *testing.T) {
	key1 := secp256k1.GenPrivKey()
	testMsg := types.SimpleMessage{
		Sender: "bob",
		Hello:  "hello",
		World:  "mars",
	}
	val1 := sdk.ValAddress("val1")
	type state struct {
		t      *testing.T
		att    *Attestator
		keeper Keeper
		valset *mocks.ValsetKeeper
		ctx    sdk.Context
	}

	for _, tt := range []struct {
		name  string
		setup func(s *state)
	}{
		{
			name: "happy path",
			setup: func(s *state) {
				t := s.t
				chainType, chainID := types.ChainTypeEVM, "test"
				queue := types.Queue(simpleQueue, chainType, chainID)

				att := mocks.NewAttestator(t)
				var msgType *types.SimpleMessage
				att.On("Type").Return(msgType)
				att.On("ChainInfo").Return(chainType, chainID)
				att.On("ConsensusQueue").Return(simpleQueue)
				att.On("BytesToSign").Return(msgType.ConsensusSignBytes())
				att.On("VerifySignature").Return(types.VerifySignatureFunc(func([]byte, []byte, []byte) bool {
					return true
				}))

				s.att.RegisterAttestator(att)

				err := s.keeper.PutMessageForSigning(s.ctx, queue, &testMsg)
				require.NoError(t, err)

				msgs, err := s.keeper.GetMessagesForSigning(s.ctx, queue, val1)
				require.NoError(t, err)
				require.Len(t, msgs, 1)

				msg := msgs[0]
				msgToSign, err := msg.ConsensusMsg(s.keeper.cdc)
				require.NoError(t, err)

				extraData := []byte("extra data")
				signedBytes, _, err := signingutils.SignBytes(
					key1,
					signingutils.SerializeFnc(signingutils.JsonDeterministicEncoding),
					msgToSign,
					msg.Nonce(),
					extraData,
				)
				require.NoError(t, err)

				att.On("ValidateEvidence", mock.Anything, &testMsg, types.Evidence{
					From: val1,
					Data: extraData,
				}).Return(nil)

				att.On("ProcessAllEvidence", mock.Anything, &testMsg, []types.Evidence{
					{
						From: val1,
						Data: extraData,
					},
				}).Return(types.AttestResult{}, nil)

				s.valset.On("GetSigningKey", s.ctx, val1, chainType, chainID).Return(
					key1.PubKey().Bytes(),
					nil,
				)

				s.valset.On("GetCurrentSnapshot", s.ctx).Return(
					&valsettypes.Snapshot{
						Validators: []valsettypes.Validator{
							{
								ShareCount: sdk.NewInt(5),
								Address:    val1,
							},
						},
						TotalShares: sdk.NewInt(5),
					},
					nil,
				)
				err = s.keeper.AddMessageSignature(s.ctx, val1, []*types.MsgAddMessagesSignatures_MsgSignedMessage{
					{
						Id:            msg.GetId(),
						QueueTypeName: queue,
						Signature:     signedBytes,
						ExtraData:     extraData,
					},
				})
				require.NoError(t, err)

				err = s.keeper.CheckAndProcessAttestedMessages(s.ctx)
				require.NoError(t, err)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ms, ctx := newConsensusKeeper(t)
			s := &state{
				t:      t,
				att:    ms.Attestator,
				keeper: *keeper,
				ctx:    ctx,
				valset: ms.ValsetKeeper,
			}

			tt.setup(s)
		})
	}
}
