package keeper

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/require"
)

// Tests that the gravity module's balance is accounted for with unbatched txs, including tx cancellation
func TestModuleBalanceUnbatchedTxs(t *testing.T) {
	////////////////// SETUP //////////////////
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	var (
		mySender, e1        = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver          = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		myTokenContractAddr = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5"
	)
	require.NoError(t, e1)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	// mint some voucher first
	allVouchersToken, err := types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr)
	require.NoError(t, err)
	allVouchers := sdk.Coins{allVouchersToken.GravityCoin()}
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
	for i, v := range []uint64{2, 3, 2, 1} {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr)
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr)
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, fee)
		require.NotZero(t, r)
		require.NoError(t, err)
		// Should create:
		// 1: amount 100, fee 2
		// 2: amount 101, fee 3
		// 3: amount 102, fee 2
		// 4: amount 103, fee 1
	}
	checkInvariant(t, ctx, input.GravityKeeper, true)

	// Remove one of the transactions
	err = input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 1, mySender)
	require.NoError(t, err)
	checkInvariant(t, ctx, input.GravityKeeper, true)

	// Ensure an error is returned for a mismatched balance
	oneVoucher, err := types.NewInternalERC20Token(sdk.NewInt(1), myTokenContractAddr)
	require.NoError(t, err)

	checkImbalancedModule(t, ctx, input.GravityKeeper, input.BankKeeper, mySender, sdk.NewCoins(oneVoucher.GravityCoin()))
}

// Tests that the gravity module's balance is accounted for with batches of txs, including unbatched txs and tx cancellation
func TestModuleBalanceBatchedTxs(t *testing.T) {
	////////////////// SETUP //////////////////
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	var (
		now                      = time.Now().UTC()
		mySender, e1             = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver, e2           = types.NewEthAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr1, e3 = types.NewEthAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5")
		myTokenContractAddr2, e4 = types.NewEthAddress("0xF815240800ddf3E0be80e0d848B13ecaa504BF37")
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	require.NoError(t, e4)
	tokens := make([]*types.InternalERC20Token, 2)
	var err error
	tokens[0], err = types.NewInternalERC20Token(sdk.NewInt(150000000000000), myTokenContractAddr1.GetAddress().Hex())
	require.NoError(t, err)
	tokens[1], err = types.NewInternalERC20Token(sdk.NewInt(150000000000000), myTokenContractAddr2.GetAddress().Hex())
	require.NoError(t, err)
	voucher1, err := types.NewInternalERC20Token(sdk.NewInt(1), myTokenContractAddr1.GetAddress().Hex())
	require.NoError(t, err)
	voucher2, err := types.NewInternalERC20Token(sdk.NewInt(1), myTokenContractAddr2.GetAddress().Hex())
	require.NoError(t, err)
	voucherCoins := []sdk.Coins{
		sdk.NewCoins(voucher1.GravityCoin()),
		sdk.NewCoins(voucher2.GravityCoin()),
	}
	allVouchers := []sdk.Coins{
		sdk.NewCoins(tokens[0].GravityCoin()),
		sdk.NewCoins(tokens[1].GravityCoin()),
	}

	// mint some voucher first
	for _, v := range allVouchers {
		require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, v))
		// set senders balance
		input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
		require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, v))
	}
	input.GravityKeeper.SetLastObservedEthereumBlockHeight(ctx, 1234567)

	////////////////// EXECUTE //////////////////
	// Check the invariant without any transactions
	checkInvariant(t, ctx, input.GravityKeeper, true)

	for _, tok := range tokens {
		// add some TX to the pool
		for i, v := range []uint64{2, 3, 2, 1, 2, 4, 5, 1} {
			amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), tok.Contract.GetAddress().Hex())
			require.NoError(t, err)
			amount := amountToken.GravityCoin()
			feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), tok.Contract.GetAddress().Hex())
			require.NoError(t, err)
			fee := feeToken.GravityCoin()

			r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *myReceiver, amount, fee)
			require.NoError(t, err)
			ctx.Logger().Info(fmt.Sprintf("Created transaction %v with amount %v and fee %v", r, amount, fee))
			// Should create:
			// 1: tx amount is 100, fee is 2, id is 1
			// 2: tx amount is 101, fee is 3, id is 2
			// 3: tx amount is 102, fee is 2, id is 3
			// 4: tx amount is 103, fee is 1, id is 4
		}
	}
	// The module should be balanced with these unbatched txs
	checkInvariant(t, ctx, input.GravityKeeper, true)

	batches := []*types.InternalOutgoingTxBatch{nil, nil}
	// Create a batch for each token, perform some checks
	for i, tok := range tokens {
		// when
		ctx = ctx.WithBlockTime(now)
		// tx batch size is 3, so that some of them stay behind
		batch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, tok.Contract, 3)
		require.NoError(t, err)
		// then check the batch persists
		gotBatch := input.GravityKeeper.GetOutgoingTXBatch(ctx, batch.TokenContract, batch.BatchNonce)
		require.NotNil(t, gotBatch)
		batches[i] = gotBatch
		// The module should be balanced with the new unobserved batch + leftover unbatched txs
		checkInvariant(t, ctx, input.GravityKeeper, true)
		checkImbalancedModule(t, ctx, input.GravityKeeper, input.BankKeeper, mySender, voucherCoins[i])
	}
	// Remove a tx from the pool for each contract (both of these have fee = 1 and won't be batched
	require.NoError(t, input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 4, mySender))
	require.NoError(t, input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 8, mySender))

	// Here we execute the most recently created batch to test the module's balance is correct after deletion of the first batch
	// All of the batch's transactions need to end up back in the unbatched tx pool and should be counted there for us

	// The module should be balanced with the unobserved batch + one leftover unbatched tx
	checkInvariant(t, ctx, input.GravityKeeper, true)
	checkImbalancedModule(t, ctx, input.GravityKeeper, input.BankKeeper, mySender, voucherCoins[1])

	// Simulate one batch being relayed and observed
	fakeBlock := batches[1].CosmosBlockCreated // A fake ethereum block used for the test only
	msg := types.MsgBatchSendToEthClaim{
		EventNonce:     0,
		EthBlockHeight: fakeBlock,
		BatchNonce:     batches[1].BatchNonce,
		TokenContract:  "",
		Orchestrator:   "",
	}
	input.GravityKeeper.OutgoingTxBatchExecuted(ctx, batches[1].TokenContract, msg)
	// The module should be balanced with the batch now being observed + one leftover unbatched tx still in the pool
	checkInvariant(t, ctx, input.GravityKeeper, true)
	checkImbalancedModule(t, ctx, input.GravityKeeper, input.BankKeeper, mySender, voucherCoins[0])
	checkImbalancedModule(t, ctx, input.GravityKeeper, input.BankKeeper, mySender, voucherCoins[1])
}

func checkInvariant(t *testing.T, ctx sdk.Context, k Keeper, succeed bool) {
	res, ok := ModuleBalanceInvariant(k)(ctx)
	if succeed {
		require.False(t, ok, "Invariant should have returned false")
		require.Empty(t, res, "Invariant should have returned no message")
	} else {
		require.True(t, ok, "Invariant should have returned true")
		require.NotEmpty(t, res, "Invariant should have returned a message")
	}
}

func checkImbalancedModule(t *testing.T, ctx sdk.Context, gravityKeeper Keeper, bankKeeper bankkeeper.BaseKeeper, sender sdk.AccAddress, coins sdk.Coins) {
	// Imbalance the module
	require.NoError(t, bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins))
	checkInvariant(t, ctx, gravityKeeper, false)
	// Rebalance the module
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, coins))
}
