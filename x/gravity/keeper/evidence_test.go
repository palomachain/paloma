package keeper

import (
	"encoding/hex"
	"testing"
	"time"

	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/require"
)

// nolint: exhaustruct
func TestSubmitBadSignatureEvidenceBatchExists(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	var (
		now          = time.Now().UTC()
		mySender, e1 = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver   = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		token, e2    = types.NewInternalERC20Token(math.NewInt(99999), testERC20Address, "test-chain")
		allVouchers  = sdk.NewCoins(sdk.NewCoin(testDenom, token.Amount))
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract, err := types.NewEthAddress(testERC20Address)
	require.NoError(t, err)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// CREATE BATCH

	// add some TX to the pool
	for i := 0; i < 4; i++ {
		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(i+100)), testERC20Address, "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)

		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, "test-chain")
		require.NoError(t, err)
	}

	// when
	ctx = sdkCtx.WithBlockTime(now)

	goodBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", *tokenContract, 2)
	require.NoError(t, err)
	goodBatchExternal := goodBatch.ToExternal()

	any, err := codectypes.NewAnyWithValue(&goodBatchExternal)
	require.NoError(t, err)

	msg := types.MsgSubmitBadSignatureEvidence{
		Subject:   any,
		Signature: "foo",
	}

	err = input.GravityKeeper.CheckBadSignatureEvidence(ctx, &msg, "test-chain")
	require.EqualError(t, err, "Checkpoint exists, cannot slash: invalid")
}

// nolint: exhaustruct
func TestSubmitBadSignatureEvidenceSlash(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	defer func() {
		sdk.UnwrapSDKContext(input.Context).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	batch := types.OutgoingTxBatch{
		TokenContract:    "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
		BatchTimeout:     420,
		ChainReferenceId: "test-chain",
	}

	checkpoint, err := batch.GetCheckpoint("")
	require.NoError(t, err)

	any, err := codectypes.NewAnyWithValue(&batch)
	require.NoError(t, err)

	ethSignature, err := types.NewEthereumSignature(checkpoint, EthPrivKeys[0])
	require.NoError(t, err)

	msg := types.MsgSubmitBadSignatureEvidence{
		Subject:   any,
		Signature: hex.EncodeToString(ethSignature),
	}

	err = input.GravityKeeper.CheckBadSignatureEvidence(ctx, &msg, "test-chain")
	require.NoError(t, err)

	val, err := input.StakingKeeper.Validator(ctx, ValAddrs[0])
	require.NoError(t, err)
	require.True(t, val.IsJailed())
}
