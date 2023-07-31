package keeper

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/require"
)

// nolint: exhaustruct
func TestSubmitBadSignatureEvidenceBatchExists(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context

	var (
		now                 = time.Now().UTC()
		mySender, e1        = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver          = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		myTokenContractAddr = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5" // Pickle
		token, e2           = types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr)
		allVouchers         = sdk.NewCoins(token.GravityCoin())
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract, err := types.NewEthAddress(myTokenContractAddr)
	require.NoError(t, err)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// CREATE BATCH

	// add some TX to the pool
	for i, v := range []uint64{2, 3, 2, 1} {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr)
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr)
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, fee)
		require.NoError(t, err)
	}

	// when
	ctx = ctx.WithBlockTime(now)

	goodBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, *tokenContract, 2)
	goodBatchExternal := goodBatch.ToExternal()
	require.NoError(t, err)

	any, err := codectypes.NewAnyWithValue(&goodBatchExternal)
	require.NoError(t, err)

	msg := types.MsgSubmitBadSignatureEvidence{
		Subject:   any,
		Signature: "foo",
	}

	err = input.GravityKeeper.CheckBadSignatureEvidence(ctx, &msg)
	require.EqualError(t, err, "Checkpoint exists, cannot slash: invalid")
}

// nolint: exhaustruct
func TestSubmitBadSignatureEvidenceValsetExists(t *testing.T) {
	// input := CreateTestEnv(t)
	input, ctx := SetupFiveValChain(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	// ctx := input.Context

	valset, err := input.GravityKeeper.SetValsetRequest(ctx)
	require.NoError(t, err)

	any, err := codectypes.NewAnyWithValue(&valset)
	require.NoError(t, err)

	msg := types.MsgSubmitBadSignatureEvidence{
		Subject:   any,
		Signature: "foo",
	}

	err = input.GravityKeeper.CheckBadSignatureEvidence(ctx, &msg)
	require.EqualError(t, err, "Checkpoint exists, cannot slash: invalid")
}

// nolint: exhaustruct
func TestSubmitBadSignatureEvidenceLogicCallExists(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context

	contract := common.BytesToAddress(bytes.Repeat([]byte{0x1}, 20)).String()
	logicCall := types.OutgoingLogicCall{
		Transfers:            []types.ERC20Token{},
		Fees:                 []types.ERC20Token{},
		LogicContractAddress: contract,
		Payload:              []byte{},
		Timeout:              420,
		InvalidationId:       []byte{},
		InvalidationNonce:    0,
		CosmosBlockCreated:   0,
	}

	input.GravityKeeper.SetOutgoingLogicCall(ctx, logicCall)

	any, err := codectypes.NewAnyWithValue(&logicCall)
	require.NoError(t, err)

	msg := types.MsgSubmitBadSignatureEvidence{
		Subject:   any,
		Signature: "foo",
	}

	err = input.GravityKeeper.CheckBadSignatureEvidence(ctx, &msg)
	require.EqualError(t, err, "Checkpoint exists, cannot slash: invalid")
}

// nolint: exhaustruct
func TestSubmitBadSignatureEvidenceSlash(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	batch := types.OutgoingTxBatch{
		TokenContract: "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
		BatchTimeout:  420,
	}

	checkpoint := batch.GetCheckpoint(input.GravityKeeper.GetGravityID(ctx))

	any, err := codectypes.NewAnyWithValue(&batch)
	require.NoError(t, err)

	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	ethAddress, err := types.NewEthAddress(crypto.PubkeyToAddress(privKey.PublicKey).String())
	require.NoError(t, err)

	input.GravityKeeper.SetEthAddressForValidator(ctx, ValAddrs[0], *ethAddress)

	ethSignature, err := types.NewEthereumSignature(checkpoint, privKey)
	require.NoError(t, err)

	msg := types.MsgSubmitBadSignatureEvidence{
		Subject:   any,
		Signature: hex.EncodeToString(ethSignature),
	}

	err = input.GravityKeeper.CheckBadSignatureEvidence(ctx, &msg)
	require.NoError(t, err)

	val := input.StakingKeeper.Validator(ctx, ValAddrs[0])
	require.True(t, val.IsJailed())
}
