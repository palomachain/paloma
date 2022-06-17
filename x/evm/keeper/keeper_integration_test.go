package keeper_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/palomachain/paloma/testutil/rand"
	"github.com/tendermint/tendermint/crypto/ed25519"
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

// func TestEndToEndForEvmArbitraryCall(t *testing.T) {
// 	chainType, chainID := consensustypes.ChainTypeEVM, "eth-main"
// 	a := app.NewTestApp(t, false)
// 	ctx := a.NewContext(false, tmproto.Header{
// 		Height: 5,
// 	})

// 	validators := genValidators(t, 25, 25000)
// 	for _, val := range validators {
// 		a.StakingKeeper.SetValidator(ctx, val)
// 	}

// 	smartContractAddr := common.BytesToAddress(rand.Bytes(5))
// 	err := a.EvmKeeper.AddSmartContractExecutionToConsensus(
// 		ctx,
// 		&types.ArbitrarySmartContractCall{
// 			Payload: func() []byte {
// 				evm := whoops.Must(abi.JSON(strings.NewReader(sample.SimpleABI)))
// 				return whoops.Must(evm.Pack("store", big.NewInt(1337)))
// 			}(),
// 			HexAddress: smartContractAddr.Hex(),
// 			Abi:        []byte(sample.SimpleABI),
// 		},
// 	)

// 	require.NoError(t, err)

// 	private, err := crypto.GenerateKey()
// 	require.NoError(t, err)

// 	accAddr := crypto.PubkeyToAddress(private.PublicKey)
// 	err = a.ValsetKeeper.AddExternalChainInfo(ctx, validators[0].GetOperator(), []*valsettypes.ExternalChainInfo{
// 		{
// 			ChainType: chainType,
// 			ChainID:   chainID,
// 			Address:   accAddr.Hex(),
// 			Pubkey:    accAddr[:],
// 		},
// 	})

// 	require.NoError(t, err)
// 	queue := consensustypes.Queue(keeper.ConsensusArbitraryContractCall, chainType, chainID)
// 	msgs, err := a.ConsensusKeeper.GetMessagesForSigning(ctx, queue, validators[0].GetOperator())

// 	for _, msg := range msgs {
// 		sigbz, err := crypto.Sign(msg.GetBytesToSign(), private)
// 		require.NoError(t, err)
// 		err = a.ConsensusKeeper.AddMessageSignature(
// 			ctx,
// 			validators[0].GetOperator(),
// 			[]*consensustypes.MsgAddMessagesSignatures_MsgSignedMessage{
// 				{
// 					Id:              msg.GetId(),
// 					QueueTypeName:   queue,
// 					Signature:       sigbz,
// 					SignedByAddress: accAddr.Hex(),
// 				},
// 			},
// 		)
// 		require.NoError(t, err)
// 	}

// }
