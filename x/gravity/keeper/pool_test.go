package keeper

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests that the pool is populated with the created transactions before any batch is created
func TestAddToOutgoingPool(t *testing.T) {
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
	tokenContract, err := types.NewEthAddress(testERC20Address)
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

	// when
	for i := 0; i < 4; i++ {
		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(i+100)), testERC20Address, "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)

		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, "test-chain")
		require.NoError(t, err)
		t.Logf("___ response: %#v", r)
		// Should create:
		// 1: amount 100
		// 2: amount 101
		// 3: amount 102
		// 4: amount 103

	}
	// then
	got, err := input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *tokenContract)
	require.NoError(t, err)

	receiverAddr, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	oneHundredTok, err := types.NewInternalERC20Token(math.NewInt(100), testERC20Address, "test-chain")
	require.NoError(t, err)
	oneHundredOneTok, err := types.NewInternalERC20Token(math.NewInt(101), testERC20Address, "test-chain")
	require.NoError(t, err)
	oneHundredTwoTok, err := types.NewInternalERC20Token(math.NewInt(102), testERC20Address, "test-chain")
	require.NoError(t, err)
	oneHundredThreeTok, err := types.NewInternalERC20Token(math.NewInt(103), testERC20Address, "test-chain")
	require.NoError(t, err)
	exp := []*types.InternalOutgoingTransferTx{
		{
			Id:          4,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  oneHundredThreeTok,
		},
		{
			Id:          3,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  oneHundredTwoTok,
		},
		{
			Id:          2,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  oneHundredOneTok,
		},
		{
			Id:          1,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  oneHundredTok,
		},
	}
	assert.Equal(t, exp, got)
}

// Checks some common edge cases like invalid inputs, user doesn't have enough tokens, token doesn't exist, inconsistent entry
func TestAddToOutgoingPoolEdgeCases(t *testing.T) {
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
	amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(100)), testERC20Address, "test-chain")
	require.NoError(t, err)
	amount := sdk.NewCoin(testDenom, amountToken.Amount)

	//////// Nonexistant Token ////////
	r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, "test-chain")
	require.Error(t, err)
	require.Zero(t, r)

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

	//////// Insufficient Balance from Amount ////////
	badAmountToken, err := types.NewInternalERC20Token(math.NewInt(999999), testERC20Address, "test-chain")
	require.NoError(t, err)
	badAmount := sdk.NewCoin(testDenom, badAmountToken.Amount)
	r, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, badAmount, "test-chain")
	require.Error(t, err)
	require.Zero(t, r)

	//////// Zero inputs ////////

	mtSend := new(sdk.AccAddress)
	mtRecieve := types.ZeroAddress() // This address should not actually cause an issue
	mtCoin := new(sdk.Coin)
	r, err = input.GravityKeeper.AddToOutgoingPool(input.Context, *mtSend, mtRecieve, *mtCoin, "test-chain")
	require.Error(t, err)
	require.Zero(t, r)
}

func TestRemoveFromOutgoingPoolAndRefund(t *testing.T) {
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
	originalBal := uint64(99999)

	allVouchersToken, err := types.NewInternalERC20Token(math.NewIntFromUint64(originalBal), testERC20Address, "test-chain")
	require.NoError(t, err)
	allVouchers := sdk.Coins{sdk.NewCoin(testDenom, allVouchersToken.Amount)}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	// Create unbatched transactions
	unbatchedTransactions, err := input.GravityKeeper.GetUnbatchedTransactions(ctx)
	require.NoError(t, err)
	require.Empty(t, unbatchedTransactions)
	ids := make([]uint64, 4)
	amountSum := uint64(0)
	amounts := []uint64{100, 101, 102, 103}
	for i, v := range amounts {
		amountToken, err := types.NewInternalERC20Token(math.NewIntFromUint64(amounts[i]), testERC20Address, "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)
		amountSum += v
		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, "test-chain")
		require.NoError(t, err)
		t.Logf("___ response: %#v", r)
		ids[i] = r
		// Should create:
		// 1: amount 100
		// 2: amount 101
		// 3: amount 102
		// 4: amount 103

	}
	// Check balance
	currentBal := input.BankKeeper.GetBalance(ctx, mySender, testDenom).Amount.Uint64()
	require.Equal(t, currentBal, originalBal-amountSum)

	// Check that removing a transaction refunds the costs and the tx no longer exists in the pool
	checkRemovedTx(t, input, ctx, ids[2], amounts[2], &amountSum, originalBal, mySender, testERC20Address, testDenom)
	checkRemovedTx(t, input, ctx, ids[3], amounts[3], &amountSum, originalBal, mySender, testERC20Address, testDenom)
	checkRemovedTx(t, input, ctx, ids[1], amounts[1], &amountSum, originalBal, mySender, testERC20Address, testDenom)
	checkRemovedTx(t, input, ctx, ids[0], amounts[0], &amountSum, originalBal, mySender, testERC20Address, testDenom)
	unbatchedTransactions, err = input.GravityKeeper.GetUnbatchedTransactions(ctx)
	require.NoError(t, err)
	require.Empty(t, unbatchedTransactions)
}

func TestRemoveFromOutgoingPoolAndRefundCosmosOriginated(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() {
		sdk.UnwrapSDKContext(input.Context).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	ctx := input.Context

	var (
		mySender, e1        = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver          = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		myTokenContractAddr = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5"
		myTokenDenom        = "grav"
	)
	require.NoError(t, e1)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	// mint some voucher first
	originalBal := uint64(99999)

	tokenAddr, err := types.NewEthAddress(myTokenContractAddr)
	require.NoError(t, err)

	// add it to the ERC20 registry
	err = input.GravityKeeper.setDenomToERC20(ctx, "test-chain", myTokenDenom, *tokenAddr)
	require.NoError(t, err)

	addr, err := input.GravityKeeper.GetERC20OfDenom(ctx, "test-chain", myTokenDenom)
	require.NoError(t, err)
	require.Equal(t, tokenAddr.GetAddress().Hex(), myTokenContractAddr)
	require.Equal(t, tokenAddr, addr)

	allVouchers := sdk.Coins{sdk.NewCoin(myTokenDenom, math.NewIntFromUint64(originalBal))}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	// Create unbatched transactions
	unbatchedTransactions, err := input.GravityKeeper.GetUnbatchedTransactions(ctx)
	require.NoError(t, err)
	require.Empty(t, unbatchedTransactions)
	amountSum := uint64(0)
	ids := make([]uint64, 4)
	amounts := []uint64{100, 101, 102, 103}
	for i, v := range amounts {
		amount := sdk.NewCoin(myTokenDenom, math.NewIntFromUint64(amounts[i]))

		amountSum += v
		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, "test-chain")
		require.NoError(t, err)
		t.Logf("___ response: %#v", r)
		ids[i] = r
		// Should create:
		// 1: amount 100
		// 2: amount 101
		// 3: amount 102
		// 4: amount 103

	}
	// Check balance
	currentBal := input.BankKeeper.GetBalance(ctx, mySender, myTokenDenom).Amount.Uint64()
	require.Equal(t, currentBal, originalBal-amountSum)

	// Check that removing a transaction refunds the costs and the tx no longer exists in the pool
	checkRemovedTx(t, input, ctx, ids[2], amounts[2], &amountSum, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	checkRemovedTx(t, input, ctx, ids[3], amounts[3], &amountSum, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	checkRemovedTx(t, input, ctx, ids[1], amounts[1], &amountSum, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	checkRemovedTx(t, input, ctx, ids[0], amounts[0], &amountSum, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	unbatchedTransactions, err = input.GravityKeeper.GetUnbatchedTransactions(ctx)
	require.NoError(t, err)
	require.Empty(t, unbatchedTransactions)
}

// Helper method to:
// 1. Remove the transaction specified by `id`, `myTokenContractAddr` and `amount`
// 2. Update the amount tracker by subtracting the refunded `amount`
// 3. Require that `mySender` has been refunded the correct amount for the cancelled transaction
// 4. Require that the unbatched transaction pool does not contain the refunded transaction via iterating its elements
func checkRemovedTx(
	t *testing.T,
	input TestInput,
	ctx context.Context,
	id uint64,
	amount uint64,
	amountSum *uint64,
	originalBal uint64,
	mySender sdk.AccAddress,
	myTokenContractAddr string,
	myTokenDenom string,
) {
	err := input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, id, mySender)
	require.NoError(t, err)
	*amountSum -= amount // user should have regained the locked amounts from tx
	currentBal := input.BankKeeper.GetBalance(ctx, mySender, myTokenDenom).Amount.Uint64()
	require.Equal(t, currentBal, originalBal-*amountSum)
	expectedKey := myTokenContractAddr + fmt.Sprint(amount) + fmt.Sprint(id)
	err = input.GravityKeeper.IterateUnbatchedTransactions(ctx, func(key []byte, tx *types.InternalOutgoingTransferTx) bool {
		require.NotEqual(t, []byte(expectedKey), key)
		found := id == tx.Id &&
			amount == tx.Erc20Token.Amount.Uint64()
		require.False(t, found)
		return false
	})
	require.NoError(t, err)
}

// ======================== Edge case tests for RemoveFromOutgoingPoolAndRefund() =================================== //

// Checks some common edge cases like invalid inputs, user didn't submit the transaction, tx doesn't exist, inconsistent entry
func TestRefundInconsistentTx(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() {
		sdk.UnwrapSDKContext(input.Context).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	ctx := input.Context
	var (
		mySender, e1            = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver, e2          = types.NewEthAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr, e3 = types.NewEthAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5")
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)

	//////// Refund an inconsistent tx ////////
	amountToken, err := types.NewInternalERC20Token(math.NewInt(100), myTokenContractAddr.GetAddress().Hex(), "test-chain")
	require.NoError(t, err)

	// This unsafe override won't fail
	err = input.GravityKeeper.addUnbatchedTX(ctx, &types.InternalOutgoingTransferTx{
		Id:          uint64(5),
		Sender:      mySender,
		DestAddress: myReceiver,
		Erc20Token:  amountToken,
	})
	origBalances := input.BankKeeper.GetAllBalances(ctx, mySender)
	require.NoError(t, err, "someone added validation to addUnbatchedTx")
	err = input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, uint64(5), mySender)
	require.Error(t, err)
	newBalances := input.BankKeeper.GetAllBalances(ctx, mySender)
	require.Equal(t, origBalances, newBalances)
}

func TestRefundNonexistentTx(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() {
		sdk.UnwrapSDKContext(input.Context).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	ctx := input.Context
	mySender, e1 := sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
	require.NoError(t, e1)

	//////// Refund a tx which never existed ////////
	origBalances := input.BankKeeper.GetAllBalances(ctx, mySender)
	err := input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, uint64(1), mySender)
	require.Error(t, err)
	newBalances := input.BankKeeper.GetAllBalances(ctx, mySender)
	require.Equal(t, origBalances, newBalances)
}

func TestRefundTwice(t *testing.T) {
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

	//////// Refund a tx twice ////////

	// mint some voucher first
	originalBal := uint64(99999)
	allVouchersToken, err := types.NewInternalERC20Token(math.NewIntFromUint64(originalBal), testERC20Address, "test-chain")
	require.NoError(t, err)
	allVouchers := sdk.Coins{sdk.NewCoin(testDenom, allVouchersToken.Amount)}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	amountToken, err := types.NewInternalERC20Token(math.NewInt(100), testERC20Address, "test-chain")
	require.NoError(t, err)
	origBalances := input.BankKeeper.GetAllBalances(ctx, mySender)

	txId, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, sdk.NewCoin(testDenom, amountToken.Amount), "test-chain")
	require.NoError(t, err)
	afterAddBalances := input.BankKeeper.GetAllBalances(ctx, mySender)

	// First refund goes through
	err = input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, txId, mySender)
	require.NoError(t, err)
	afterRefundBalances := input.BankKeeper.GetAllBalances(ctx, mySender)

	// Second fails
	err = input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, txId, mySender)
	require.Error(t, err)
	afterSecondRefundBalances := input.BankKeeper.GetAllBalances(ctx, mySender)

	require.NotEqual(t, origBalances, afterAddBalances)
	require.Equal(t, origBalances, afterRefundBalances)
	require.Equal(t, origBalances, afterSecondRefundBalances)
}

// Check the various getter methods for the pool
func TestGetUnbatchedTransactions(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	defer func() {
		sdk.UnwrapSDKContext(ctx).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	// token1
	var (
		mySender, e1 = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver   = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
	)
	require.NoError(t, e1)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract1, err := types.NewEthAddress(testERC20Address)
	require.NoError(t, err)
	// mint some vouchers first
	allVouchersToken, err := types.NewInternalERC20Token(math.NewInt(99999), testERC20Address, "test-chain")
	require.NoError(t, err)
	allVouchers := sdk.Coins{sdk.NewCoin(testDenom, allVouchersToken.Amount)}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	ids := make([]uint64, 4)
	amounts := []uint64{100, 101, 102, 103}
	idToTxMap := make(map[uint64]*types.OutgoingTransferTx)
	for i := 0; i < 4; i++ {
		amountToken, err := types.NewInternalERC20Token(math.NewIntFromUint64(amounts[i]), testERC20Address, "test-chain")
		require.NoError(t, err)
		amount1 := sdk.NewCoin(testDenom, amountToken.Amount)

		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount1, "test-chain")
		require.NoError(t, err)
		ids[i] = r
		idToTxMap[r] = &types.OutgoingTransferTx{
			Id:          r,
			Sender:      mySender.String(),
			DestAddress: myReceiver,
			Erc20Token:  amountToken.ToExternal(),
		}
	}

	// GetUnbatchedTxByAmountAndId
	tokenAmount, err := types.NewInternalERC20Token(math.NewIntFromUint64(amounts[0]), testERC20Address, "test-chain")
	require.NoError(t, err)
	tokenId := ids[0]
	tx, err := input.GravityKeeper.GetUnbatchedTxByAmountAndId(ctx, *tokenAmount, tokenId)
	require.NoError(t, err)
	expTx, err := types.NewInternalOutgoingTransferTx(tokenId, mySender.String(), myReceiver, tokenAmount.ToExternal())
	require.NoError(t, err)
	require.Equal(t, *expTx, *tx)

	// GetUnbatchedTxById
	tx, err = input.GravityKeeper.GetUnbatchedTxById(ctx, tokenId)
	require.NoError(t, err)
	require.Equal(t, *expTx, *tx)

	// GetUnbatchedTransactionsByContract
	tokenTxs, err := input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *tokenContract1)
	require.NoError(t, err)
	for _, v := range tokenTxs {
		expTx := idToTxMap[v.Id]
		require.NotNil(t, expTx)
		require.Equal(t, testERC20Address, v.Erc20Token.Contract.GetAddress().Hex())
		require.Equal(t, expTx.DestAddress, v.DestAddress.GetAddress().Hex())
		require.Equal(t, expTx.Sender, v.Sender.String())
	}
	// GetUnbatchedTransactions
	allTxs, err := input.GravityKeeper.GetUnbatchedTransactions(ctx)
	require.NoError(t, err)
	for _, v := range allTxs {
		expTx := idToTxMap[v.Id]
		require.NotNil(t, expTx)
		require.Equal(t, expTx.DestAddress, v.DestAddress.GetAddress().Hex())
		require.Equal(t, expTx.Sender, v.Sender.String())
		require.Equal(t, expTx.Erc20Token.Contract, v.Erc20Token.Contract.GetAddress().Hex())
	}
}

// Check the various iteration methods for the pool
func TestIterateUnbatchedTransactions(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() {
		sdk.UnwrapSDKContext(input.Context).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	ctx := input.Context

	// token
	var (
		mySender, e1 = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver   = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
	)
	require.NoError(t, e1)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract1, err := types.NewEthAddress(testERC20Address)
	require.NoError(t, err)
	// mint some vouchers first
	token, err := types.NewInternalERC20Token(math.NewInt(99999), testERC20Address, "test-chain")
	require.NoError(t, err)
	allVouchers := sdk.Coins{sdk.NewCoin(testDenom, token.Amount)}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	ids := make([]uint64, 4)
	amounts := []uint64{100, 101, 102, 103}
	idToTxMap := make(map[uint64]*types.OutgoingTransferTx)
	for i := 0; i < 4; i++ {
		amount1, err := types.NewInternalERC20Token(math.NewIntFromUint64(amounts[i]), testERC20Address, "test-chain")
		require.NoError(t, err)
		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, sdk.NewCoin(testDenom, amount1.Amount), "test-chain")
		require.NoError(t, err)
		ids[i] = r
		idToTxMap[r] = &types.OutgoingTransferTx{
			Id:          r,
			Sender:      mySender.String(),
			DestAddress: myReceiver,
			Erc20Token:  amount1.ToExternal(),
		}
	}
	// IterateUnbatchedTransactionsByContract
	foundMap := make(map[uint64]bool)
	err = input.GravityKeeper.IterateUnbatchedTransactionsByContract(ctx, *tokenContract1, func(key []byte, tx *types.InternalOutgoingTransferTx) bool {
		require.NotNil(t, tx)
		fTx := idToTxMap[tx.Id]
		require.NotNil(t, fTx)
		require.Equal(t, fTx.Erc20Token.Contract, testERC20Address)
		require.Equal(t, fTx.DestAddress, myReceiver)
		require.Equal(t, mySender.String(), fTx.Sender)
		foundMap[fTx.Id] = true
		return false
	})
	require.NoError(t, err)

	for i := 1; i <= 4; i++ {
		require.True(t, foundMap[uint64(i)])
	}
	// filterAndIterateUnbatchedTransactions
	anotherFoundMap := make(map[uint64]bool)
	err = input.GravityKeeper.IterateUnbatchedTransactions(ctx, func(key []byte, tx *types.InternalOutgoingTransferTx) bool {
		require.NotNil(t, tx)
		fTx := idToTxMap[tx.Id]
		require.NotNil(t, fTx)
		require.Equal(t, fTx.DestAddress, tx.DestAddress.GetAddress().Hex())

		anotherFoundMap[fTx.Id] = true
		return false
	})
	require.NoError(t, err)

	for i := 1; i <= 4; i++ {
		require.True(t, anotherFoundMap[uint64(i)])
	}
}

// Ensures that any unbatched tx will make its way into the exported data from ExportGenesis
func TestAddToOutgoingPoolExportGenesis(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() {
		sdk.UnwrapSDKContext(input.Context).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	ctx := input.Context
	k := input.GravityKeeper
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

	unbatchedTxMap := make(map[uint64]types.OutgoingTransferTx)
	foundTxsMap := make(map[uint64]bool)
	// when
	for i := 0; i < 4; i++ {
		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(i+100)), testERC20Address, "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)

		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, "test-chain")
		require.NoError(t, err)

		unbatchedTxMap[r] = types.OutgoingTransferTx{
			Id:          r,
			Sender:      mySender.String(),
			DestAddress: myReceiver,
			Erc20Token:  amountToken.ToExternal(),
		}
		foundTxsMap[r] = false

	}
	// then
	got := ExportGenesis(ctx, k)
	require.NotNil(t, got)

	for _, tx := range got.UnbatchedTransfers {
		cached := unbatchedTxMap[tx.Id]
		require.NotNil(t, cached)
		require.Equal(t, cached, tx, "cached: %+v\nactual: %+v\n", cached, tx)
		foundTxsMap[tx.Id] = true
	}

	for _, v := range foundTxsMap {
		require.True(t, v)
	}
}
