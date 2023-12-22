package keeper

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/require"
)

// Tests that the gravity module's balance is accounted for with unbatched txs, including tx cancellation
func TestModuleBalanceUnbatchedTxs(t *testing.T) {
	////////////////// SETUP //////////////////
	input := CreateTestEnv(t)
	defer func() {
		sdk.UnwrapSDKContext(input.Context).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	ctx := input.Context
	var (
		mySender, e1 = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver   = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
	)
	require.NoError(t, e1)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	// mint some voucher first
	allVouchersToken, err := types.NewInternalERC20Token(math.NewInt(99999), testERC20Address, "test-chain")
	require.NoError(t, err)
	allVouchers := sdk.Coins{sdk.NewCoin(testDenom, allVouchersToken.Amount)}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	////////////////// EXECUTE //////////////////
	// Check the invariant without any transactions
	checkInvariant(t, ctx, input.GravityKeeper, true)

	// Create some unbatched transactions
	for i := 0; i < 4; i++ {
		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(i+100)), testERC20Address, "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)

		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, "test-chain")
		require.NotZero(t, r)
		require.NoError(t, err)
		// Should create:
		// 1: amount 100
		// 2: amount 101
		// 3: amount 102
		// 4: amount 103
	}
	checkInvariant(t, ctx, input.GravityKeeper, true)

	// Remove one of the transactions
	err = input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 1, mySender)
	require.NoError(t, err)
	checkInvariant(t, ctx, input.GravityKeeper, true)

	// Ensure an error is returned for a mismatched balance
	oneVoucher, err := types.NewInternalERC20Token(math.NewInt(1), testERC20Address, "test-chain")
	require.NoError(t, err)

	checkImbalancedModule(t, ctx, input.GravityKeeper, input.BankKeeper, mySender, sdk.NewCoins(sdk.NewCoin(testDenom, oneVoucher.Amount)))
}

// Tests that the gravity module's balance is accounted for with batches of txs, including unbatched txs and tx cancellation
func TestModuleBalanceBatchedTxs(t *testing.T) {
	////////////////// SETUP //////////////////
	input, ctx := SetupFiveValChain(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	var (
		now                     = time.Now().UTC()
		mySender, e1            = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver, e2          = types.NewEthAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr, e3 = types.NewEthAddress(testERC20Address)
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	token, err := types.NewInternalERC20Token(math.NewInt(150000000000000), myTokenContractAddr.GetAddress().Hex(), "test-chain")
	require.NoError(t, err)
	voucher, err := types.NewInternalERC20Token(math.NewInt(1), myTokenContractAddr.GetAddress().Hex(), "test-chain")
	require.NoError(t, err)
	voucherCoin := sdk.NewCoins(sdk.NewCoin(testDenom, voucher.Amount))
	tokenCoin := sdk.NewCoins(sdk.NewCoin(testDenom, token.Amount))

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, tokenCoin))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, tokenCoin))

	err = input.GravityKeeper.SetLastObservedEthereumBlockHeight(ctx, 1234567)
	require.NoError(t, err)

	////////////////// EXECUTE //////////////////
	// Check the invariant without any transactions
	checkInvariant(t, ctx, input.GravityKeeper, true)

	// add some TX to the pool
	for i := 0; i < 8; i++ {
		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(i+100)), token.Contract.GetAddress().Hex(), "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)

		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *myReceiver, amount, "test-chain")
		require.NoError(t, err)
		// Should create:
		// 1: tx amount is 100, id is 1
		// 2: tx amount is 101, id is 2
		// 3: tx amount is 102, id is 3
		// 4: tx amount is 103, id is 4
	}
	// The module should be balanced with these unbatched txs
	checkInvariant(t, ctx, input.GravityKeeper, true)

	// Create a batch, perform some checks
	// when
	ctx = sdkCtx.WithBlockTime(now)
	// tx batch size is 3, so that some of them stay behind
	batch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", token.Contract, 3)
	require.NoError(t, err)
	// then check the batch persists
	batch, err = input.GravityKeeper.GetOutgoingTXBatch(ctx, batch.TokenContract, batch.BatchNonce)
	require.NotNil(t, batch)
	require.NoError(t, err)

	// The module should be balanced with the new unobserved batch + leftover unbatched txs
	checkInvariant(t, ctx, input.GravityKeeper, true)
	checkImbalancedModule(t, ctx, input.GravityKeeper, input.BankKeeper, mySender, voucherCoin)

	// Remove a tx from the pool
	require.NoError(t, input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 4, mySender))

	// Simulate one batch being relayed and observed
	fakeBlock := batch.PalomaBlockCreated // A fake ethereum block used for the test only
	msg := types.MsgBatchSendToEthClaim{
		EventNonce:       0,
		EthBlockHeight:   fakeBlock,
		BatchNonce:       batch.BatchNonce,
		TokenContract:    testERC20Address,
		ChainReferenceId: "test-chain",
		Orchestrator:     "",
	}
	err = input.GravityKeeper.OutgoingTxBatchExecuted(ctx, batch.TokenContract, msg)
	require.NoError(t, err)
	// The module should be balanced with the batch now being observed + one leftover unbatched tx still in the pool
	checkInvariant(t, ctx, input.GravityKeeper, true)
	checkImbalancedModule(t, ctx, input.GravityKeeper, input.BankKeeper, mySender, voucherCoin)
}

func checkInvariant(t *testing.T, ctx context.Context, k Keeper, succeed bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	res, ok := ModuleBalanceInvariant(k)(sdkCtx)
	if succeed {
		require.False(t, ok, "Invariant should have returned false")
		require.Empty(t, res, "Invariant should have returned no message")
	} else {
		require.True(t, ok, "Invariant should have returned true")
		require.NotEmpty(t, res, "Invariant should have returned a message")
	}
}

func checkImbalancedModule(t *testing.T, ctx context.Context, gravityKeeper Keeper, bankKeeper bankkeeper.BaseKeeper, sender sdk.AccAddress, coins sdk.Coins) {
	// Imbalance the module
	require.NoError(t, bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins))
	checkInvariant(t, ctx, gravityKeeper, false)
	// Rebalance the module
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, coins))
}
