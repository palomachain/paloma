package keeper_test

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/testutil/rand"
	"github.com/palomachain/paloma/testutil/sample"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/keeper"
	"github.com/palomachain/paloma/x/evm/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/vizualni/whoops"
)

func genValidators(t *testing.T, numValidators, totalConsPower int) []stakingtypes.Validator {
	validators := make([]stakingtypes.Validator, numValidators)

	quotient, remainder := totalConsPower/numValidators, totalConsPower%numValidators

	for i := 0; i < numValidators; i++ {
		power := quotient
		if i == 0 {
			power += remainder
		}

		protoPK, err := cryptocodec.FromTmPubKeyInterface(ed25519.GenPrivKey().PubKey())
		if err != nil {
			panic(err)
		}

		pk, err := codectypes.NewAnyWithValue(protoPK)
		if err != nil {
			panic(err)
		}

		validators[i] = stakingtypes.Validator{
			OperatorAddress: rand.ValAddress().String(),
			Tokens:          sdk.TokensFromConsensusPower(int64(power), sdk.DefaultPowerReduction),
			Status:          stakingtypes.Bonded,
			ConsensusPubkey: pk,
		}
	}

	return validators
}

func TestEndToEndForEvmArbitraryCall(t *testing.T) {
	chainType, chainReferenceID := consensustypes.ChainTypeEVM, "eth-main"
	a := app.NewTestApp(t, false)
	ctx := a.NewContext(false, tmproto.Header{
		Height: 5,
	})

	newChain := &types.AddChainProposal{
		ChainReferenceID:           "eth-main",
		Title:             "bla",
		Description:       "bla",
		BlockHeight:       uint64(123),
		BlockHashAtHeight: "0x1234",
	}

	err := a.EvmKeeper.AddSupportForNewChain(ctx, newChain)
	require.NoError(t, err)

	err = a.EvmKeeper.ActivateChainReferenceID(ctx, newChain.ChainReferenceID, "addr", "id")
	require.NoError(t, err)

	validators := genValidators(t, 25, 25000)
	for _, val := range validators {
		a.StakingKeeper.SetValidator(ctx, val)
	}

	smartContractAddr := common.BytesToAddress(rand.Bytes(5))
	err = a.EvmKeeper.AddSmartContractExecutionToConsensus(
		ctx,
		chainReferenceID,
		"",
		&types.SubmitLogicCall{
			Payload: func() []byte {
				evm := whoops.Must(abi.JSON(strings.NewReader(sample.SimpleABI)))
				return whoops.Must(evm.Pack("store", big.NewInt(1337)))
			}(),
			HexContractAddress: smartContractAddr.Hex(),
			Abi:                []byte(sample.SimpleABI),
			Deadline:           1337,
		},
	)

	require.NoError(t, err)

	private, err := crypto.GenerateKey()
	require.NoError(t, err)

	accAddr := crypto.PubkeyToAddress(private.PublicKey)
	err = a.ValsetKeeper.AddExternalChainInfo(ctx, validators[0].GetOperator(), []*valsettypes.ExternalChainInfo{
		{
			ChainType: chainType,
			ChainReferenceID:   chainReferenceID,
			Address:   accAddr.Hex(),
			Pubkey:    accAddr[:],
		},
	})

	require.NoError(t, err)
	queue := consensustypes.Queue(keeper.ConsensusTurnstoneMessage, chainType, chainReferenceID)
	msgs, err := a.ConsensusKeeper.GetMessagesForSigning(ctx, queue, validators[0].GetOperator())

	for _, msg := range msgs {
		sigbz, err := crypto.Sign(
			crypto.Keccak256(
				[]byte(keeper.SignaturePrefix),
				msg.GetBytesToSign(),
			),
			private,
		)
		require.NoError(t, err)
		err = a.ConsensusKeeper.AddMessageSignature(
			ctx,
			validators[0].GetOperator(),
			[]*consensustypes.MsgAddMessagesSignatures_MsgSignedMessage{
				{
					Id:              msg.GetId(),
					QueueTypeName:   queue,
					Signature:       sigbz,
					SignedByAddress: accAddr.Hex(),
				},
			},
		)
		require.NoError(t, err)
	}

}

func TestOnSnapshotBuilt(t *testing.T) {
	a := app.NewTestApp(t, false)
	ctx := a.NewContext(false, tmproto.Header{
		Height: 5,
	})

	newChain := &types.AddChainProposal{
		ChainReferenceID:           "bob",
		Title:             "bla",
		Description:       "bla",
		BlockHeight:       uint64(123),
		BlockHashAtHeight: "0x1234",
	}
	err := a.EvmKeeper.AddSupportForNewChain(ctx, newChain)
	require.NoError(t, err)
	err = a.EvmKeeper.ActivateChainReferenceID(ctx, newChain.ChainReferenceID, "addr", "id")
	require.NoError(t, err)

	validators := genValidators(t, 25, 25000)
	for _, val := range validators {
		a.StakingKeeper.SetValidator(ctx, val)
		err = a.ValsetKeeper.AddExternalChainInfo(ctx, val.GetOperator(), []*valsettypes.ExternalChainInfo{
			{
				ChainType: "EVM",
				ChainReferenceID:   "bob",
				Address:   rand.ETHAddress().Hex(),
				Pubkey:    []byte("pk"),
			},
		})
		require.NoError(t, err)
	}

	queue := fmt.Sprintf("EVM/%s/%s", newChain.GetChainReferenceID(), keeper.ConsensusTurnstoneMessage)

	msgs, err := a.ConsensusKeeper.GetMessagesFromQueue(ctx, queue, 1)
	require.NoError(t, err)
	require.Empty(t, msgs)

	_, err = a.ValsetKeeper.TriggerSnapshotBuild(ctx)
	require.NoError(t, err)

	msgs, err = a.ConsensusKeeper.GetMessagesFromQueue(ctx, queue, 1)
	require.NoError(t, err)
	require.Len(t, msgs, 1)

}

func TestAddingSupportForNewChain(t *testing.T) {
	a := app.NewTestApp(t, false)
	ctx := a.NewContext(false, tmproto.Header{
		Height: 5,
	})

	t.Run("with happy path there are no errors", func(t *testing.T) {
		newChain := &types.AddChainProposal{
			ChainReferenceID:           "bob",
			Title:             "bla",
			Description:       "bla",
			BlockHeight:       uint64(123),
			BlockHashAtHeight: "0x1234",
		}
		err := a.EvmKeeper.AddSupportForNewChain(ctx, newChain)
		require.NoError(t, err)

		gotChainInfo, err := a.EvmKeeper.GetChainInfo(ctx, newChain.GetChainReferenceID())
		require.NoError(t, err)

		require.Equal(t, newChain.GetChainReferenceID(), gotChainInfo.GetChainReferenceID())
		require.Equal(t, newChain.GetBlockHashAtHeight(), gotChainInfo.GetReferenceBlockHash())
		require.Equal(t, newChain.GetBlockHeight(), gotChainInfo.GetReferenceBlockHeight())
	})

	t.Run("when chainReferenceID already exists then it returns an error", func(t *testing.T) {
		newChain := &types.AddChainProposal{
			ChainReferenceID:           "bob",
			Title:             "bla",
			Description:       "bla",
			BlockHeight:       uint64(123),
			BlockHashAtHeight: "0x1234",
		}
		err := a.EvmKeeper.AddSupportForNewChain(ctx, newChain)
		require.Error(t, err)
	})

	t.Run("activiting chain", func(t *testing.T) {
		t.Run("if the chain does not exist it returns the error", func(t *testing.T) {
			err := a.EvmKeeper.ActivateChainReferenceID(ctx, "i dont exist", "bla", "bob")
			require.Error(t, err)
		})
		t.Run("works when chain exists", func(t *testing.T) {
			err := a.EvmKeeper.ActivateChainReferenceID(ctx, "bob", "addr", "id")
			require.NoError(t, err)
			gotChainInfo, err := a.EvmKeeper.GetChainInfo(ctx, "bob")
			require.NoError(t, err)

			require.Equal(t, "addr", gotChainInfo.GetSmartContractAddr())
			require.Equal(t, "id", gotChainInfo.GetSmartContractID())
		})
	})

	t.Run("removing chain", func(t *testing.T) {
		t.Run("if the chain does not exist it returns the error", func(t *testing.T) {
			err := a.EvmKeeper.RemoveSupportForChain(ctx, &types.RemoveChainProposal{
				ChainReferenceID: "i don't exist",
			})
			require.Error(t, err)
		})
		t.Run("works when chain exists", func(t *testing.T) {
			err := a.EvmKeeper.RemoveSupportForChain(ctx, &types.RemoveChainProposal{
				ChainReferenceID: "bob",
			})
			require.NoError(t, err)
			_, err = a.EvmKeeper.GetChainInfo(ctx, "bob")
			require.Error(t, keeper.ErrChainNotFound)
		})
	})
}
