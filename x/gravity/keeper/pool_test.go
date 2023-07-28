package keeper

import (
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palomachain/paloma/x/gravity/types"
)

// Tests that the pool is populated with the created transactions before any batch is created
func TestAddToOutgoingPool(t *testing.T) {
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
	tokenContract, err := types.NewEthAddress(myTokenContractAddr)
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

	// when
	for i, v := range []uint64{2, 3, 2, 1} {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr)
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr)
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, fee)
		require.NoError(t, err)
		t.Logf("___ response: %#v", r)
		// Should create:
		// 1: amount 100, fee 2
		// 2: amount 101, fee 3
		// 3: amount 102, fee 2
		// 4: amount 103, fee 1

	}
	// then
	got := input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *tokenContract)

	receiverAddr, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	threeTok, err := types.NewInternalERC20Token(sdk.NewInt(3), myTokenContractAddr)
	require.NoError(t, err)
	twoTok, err := types.NewInternalERC20Token(sdk.NewInt(2), myTokenContractAddr)
	require.NoError(t, err)
	oneTok, err := types.NewInternalERC20Token(sdk.NewInt(1), myTokenContractAddr)
	require.NoError(t, err)
	oneHundredTok, err := types.NewInternalERC20Token(sdk.NewInt(100), myTokenContractAddr)
	require.NoError(t, err)
	oneHundredOneTok, err := types.NewInternalERC20Token(sdk.NewInt(101), myTokenContractAddr)
	require.NoError(t, err)
	oneHundredTwoTok, err := types.NewInternalERC20Token(sdk.NewInt(102), myTokenContractAddr)
	require.NoError(t, err)
	oneHundredThreeTok, err := types.NewInternalERC20Token(sdk.NewInt(103), myTokenContractAddr)
	require.NoError(t, err)
	exp := []*types.InternalOutgoingTransferTx{
		{
			Id:          2,
			Erc20Fee:    threeTok,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  oneHundredOneTok,
		},
		{
			Id:          3,
			Erc20Fee:    twoTok,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  oneHundredTwoTok,
		},
		{
			Id:          1,
			Erc20Fee:    twoTok,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  oneHundredTok,
		},
		{
			Id:          4,
			Erc20Fee:    oneTok,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  oneHundredThreeTok,
		},
	}
	assert.Equal(t, exp, got)
}

// Checks some common edge cases like invalid inputs, user doesn't have enough tokens, token doesn't exist, inconsistent entry
func TestAddToOutgoingPoolEdgeCases(t *testing.T) {
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
	amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(100)), myTokenContractAddr)
	require.NoError(t, err)
	amount := amountToken.GravityCoin()
	feeToken, err := types.NewInternalERC20Token(sdk.NewInt(2), myTokenContractAddr)
	require.NoError(t, err)
	fee := feeToken.GravityCoin()

	//////// Nonexistant Token ////////
	r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, fee)
	require.Error(t, err)
	require.Zero(t, r)

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

	//////// Insufficient Balance from Amount ////////
	badAmountToken, err := types.NewInternalERC20Token(sdk.NewInt(999999), myTokenContractAddr)
	require.NoError(t, err)
	badAmount := badAmountToken.GravityCoin()
	r, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, badAmount, fee)
	require.Error(t, err)
	require.Zero(t, r)

	//////// Insufficient Balance from Fee ////////
	badFeeToken, err := types.NewInternalERC20Token(sdk.NewInt(999999), myTokenContractAddr)
	require.NoError(t, err)
	badFee := badFeeToken.GravityCoin()
	r, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, badFee)
	require.Error(t, err)
	require.Zero(t, r)

	//////// Insufficient Balance from Amount and Fee ////////
	// Amount is 100, fee is the current balance - 99
	badFeeToken, err = types.NewInternalERC20Token(sdk.NewInt(99999-99), myTokenContractAddr)
	require.NoError(t, err)
	badFee = badFeeToken.GravityCoin()
	r, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, badFee)
	require.Error(t, err)
	require.Zero(t, r)

	//////// Zero inputs ////////
	mtCtx := new(sdk.Context)
	mtSend := new(sdk.AccAddress)
	var mtRecieve = types.ZeroAddress() // This address should not actually cause an issue
	mtCoin := new(sdk.Coin)
	r, err = input.GravityKeeper.AddToOutgoingPool(*mtCtx, *mtSend, mtRecieve, *mtCoin, *mtCoin)
	require.Error(t, err)
	require.Zero(t, r)

	//////// Inconsistent Entry ////////
	badFeeContractAddr := "0x429881672b9AE42b8eBA0e26cd9c73711b891ca6"
	badFeeToken, err = types.NewInternalERC20Token(sdk.NewInt(100), badFeeContractAddr)
	require.NoError(t, err)
	badFee = badFeeToken.GravityCoin()
	r, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, badFee)
	require.Error(t, err)
	require.Zero(t, r)
}

func TestTotalBatchFeeInPool(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context

	// token1
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

	// create outgoing pool
	for i, v := range []uint64{2, 3, 2, 1} {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr)
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr)
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		r, err2 := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, fee)
		require.NoError(t, err2)
		t.Logf("___ response: %#v", r)
	}

	// token 2 - Only top 100
	var (
		myToken2ContractAddr = "0x7D1AfA7B718fb893dB30A3aBc0Cfc608AaCfeBB0"
	)
	// mint some voucher first
	allVouchersToken, err = types.NewInternalERC20Token(sdk.NewIntFromUint64(uint64(18446744073709551615)), myToken2ContractAddr)
	require.NoError(t, err)
	allVouchers = sdk.Coins{allVouchersToken.GravityCoin()}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	// Add

	// create outgoing pool
	for i := 0; i < 110; i++ {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myToken2ContractAddr)
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(5), myToken2ContractAddr)
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, fee)
		require.NoError(t, err)
		t.Logf("___ response: %#v", r)
	}

	batchFees := input.GravityKeeper.GetAllBatchFees(ctx, OutgoingTxBatchSize)
	/*
		tokenFeeMap should be
		map[0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5:8 0x7D1AfA7B718fb893dB30A3aBc0Cfc608AaCfeBB0:500]
		**/
	assert.Equal(t, batchFees[0].TotalFees.BigInt(), big.NewInt(int64(8)))
	assert.Equal(t, batchFees[0].TxCount, uint64(4))
	assert.Equal(t, batchFees[1].TotalFees.BigInt(), big.NewInt(int64(500)))
	assert.Equal(t, batchFees[1].TxCount, uint64(100))

}

func TestGetBatchFeeByTokenType(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context

	// token1
	var (
		mySender1, e1                       = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		mySender2            sdk.AccAddress = []byte("gravity1ahx7f8wyertus")
		mySender3            sdk.AccAddress = []byte("gravity1ahx7f8wyertut")
		myReceiver                          = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		myTokenContractAddr1                = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5"
		myTokenContractAddr2                = "0x429881672b9AE42b8eBA0e26cd9c73711b891ca6"
		myTokenContractAddr3                = "0x429881672b9aE42b8eba0e26cD9c73711B891Ca7"
	)
	require.NoError(t, e1)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract1, err := types.NewEthAddress(myTokenContractAddr1)
	require.NoError(t, err)
	tokenContract2, err := types.NewEthAddress(myTokenContractAddr2)
	require.NoError(t, err)
	tokenContract3, err := types.NewEthAddress(myTokenContractAddr3)
	require.NoError(t, err)
	// mint some vouchers first
	allVouchersToken1, err := types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr1)
	require.NoError(t, err)
	allVouchers1 := sdk.Coins{allVouchersToken1.GravityCoin()}
	allVouchersToken2, err := types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr2)
	require.NoError(t, err)
	allVouchers2 := sdk.Coins{allVouchersToken2.GravityCoin()}
	allVouchersToken3, err := types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr3)
	require.NoError(t, err)
	allVouchers3 := sdk.Coins{allVouchersToken3.GravityCoin()}

	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers1)
	require.NoError(t, err)
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers2)
	require.NoError(t, err)
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers3)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender1)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender1, allVouchers1)
	require.NoError(t, err)
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender2)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender2, allVouchers2)
	require.NoError(t, err)
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender3)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender3, allVouchers3)
	require.NoError(t, err)

	totalFee1 := int64(0)
	totalFee2 := int64(0)
	totalFee3 := int64(0)
	// create outgoing pool
	for i := 0; i < 110; i++ {
		amountToken1, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr1)
		require.NoError(t, err)
		amount1 := amountToken1.GravityCoin()
		feeAmt1 := int64(i + 1) // fees can't be 0
		feeToken1, err := types.NewInternalERC20Token(sdk.NewInt(feeAmt1), myTokenContractAddr1)
		require.NoError(t, err)
		fee1 := feeToken1.GravityCoin()
		amountToken2, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr2)
		require.NoError(t, err)
		amount2 := amountToken2.GravityCoin()
		feeAmt2 := int64(2*i + 1) // fees can't be 0
		feeToken2, err := types.NewInternalERC20Token(sdk.NewInt(feeAmt2), myTokenContractAddr2)
		require.NoError(t, err)
		fee2 := feeToken2.GravityCoin()
		amountToken3, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr3)
		require.NoError(t, err)
		amount3 := amountToken3.GravityCoin()
		feeAmt3 := int64(3*i + 1) // fees can't be 0
		feeToken3, err := types.NewInternalERC20Token(sdk.NewInt(feeAmt3), myTokenContractAddr3)
		require.NoError(t, err)
		fee3 := feeToken3.GravityCoin()

		if i >= 10 {
			totalFee1 += feeAmt1
		}
		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender1, *receiver, amount1, fee1)
		require.NoError(t, err)
		t.Logf("___ response: %d", r)

		if i >= 10 {
			totalFee2 += feeAmt2
		}
		r, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender2, *receiver, amount2, fee2)
		require.NoError(t, err)
		t.Logf("___ response: %d", r)

		if i >= 10 {
			totalFee3 += feeAmt3
		}
		r, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender3, *receiver, amount3, fee3)
		require.NoError(t, err)
		t.Logf("___ response: %d", r)
	}

	batchFee1 := input.GravityKeeper.GetBatchFeeByTokenType(ctx, *tokenContract1, 100)
	require.Equal(t, batchFee1.Token, myTokenContractAddr1)
	require.Equal(t, batchFee1.TotalFees.Uint64(), uint64(totalFee1), fmt.Errorf("expected total fees %d but got %d", batchFee1.TotalFees.Uint64(), uint64(totalFee1)))
	require.Equal(t, batchFee1.TxCount, uint64(100), fmt.Errorf("expected tx count %d but got %d", batchFee1.TxCount, uint64(100)))
	batchFee2 := input.GravityKeeper.GetBatchFeeByTokenType(ctx, *tokenContract2, 100)
	require.Equal(t, batchFee2.Token, myTokenContractAddr2)
	require.Equal(t, batchFee2.TotalFees.Uint64(), uint64(totalFee2), fmt.Errorf("expected total fees %d but got %d", batchFee2.TotalFees.Uint64(), uint64(totalFee2)))
	require.Equal(t, batchFee2.TxCount, uint64(100), fmt.Errorf("expected tx count %d but got %d", batchFee2.TxCount, uint64(100)))
	batchFee3 := input.GravityKeeper.GetBatchFeeByTokenType(ctx, *tokenContract3, 100)
	require.Equal(t, batchFee3.Token, myTokenContractAddr3)
	require.Equal(t, batchFee3.TotalFees.Uint64(), uint64(totalFee3), fmt.Errorf("expected total fees %d but got %d", batchFee3.TotalFees.Uint64(), uint64(totalFee3)))
	require.Equal(t, batchFee3.TxCount, uint64(100), fmt.Errorf("expected tx count %d but got %d", batchFee3.TxCount, uint64(100)))

}

func TestRemoveFromOutgoingPoolAndRefund(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	var (
		mySender, e1        = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver          = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		myTokenContractAddr = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5"
		myTokenDenom        = "gravity" + myTokenContractAddr
	)
	require.NoError(t, e1)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	// mint some voucher first
	originalBal := uint64(99999)

	allVouchersToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(originalBal), myTokenContractAddr)
	require.NoError(t, err)
	allVouchers := sdk.Coins{allVouchersToken.GravityCoin()}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	// Create unbatched transactions
	require.Empty(t, input.GravityKeeper.GetUnbatchedTransactions(ctx))
	feesAndAmounts := uint64(0)
	ids := make([]uint64, 4)
	fees := []uint64{2, 3, 2, 1}
	amounts := []uint64{100, 101, 102, 103}
	for i, v := range fees {
		amountToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(amounts[i]), myTokenContractAddr)
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr)
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		feesAndAmounts += v + amounts[i]
		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, fee)
		require.NoError(t, err)
		t.Logf("___ response: %#v", r)
		ids[i] = r
		// Should create:
		// 1: amount 100, fee 2
		// 2: amount 101, fee 3
		// 3: amount 102, fee 2
		// 4: amount 103, fee 1

	}
	// Check balance
	currentBal := input.BankKeeper.GetBalance(ctx, mySender, myTokenDenom).Amount.Uint64()
	require.Equal(t, currentBal, originalBal-feesAndAmounts)

	// Check that removing a transaction refunds the costs and the tx no longer exists in the pool
	checkRemovedTx(t, input, ctx, ids[2], fees[2], amounts[2], &feesAndAmounts, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	checkRemovedTx(t, input, ctx, ids[3], fees[3], amounts[3], &feesAndAmounts, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	checkRemovedTx(t, input, ctx, ids[1], fees[1], amounts[1], &feesAndAmounts, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	checkRemovedTx(t, input, ctx, ids[0], fees[0], amounts[0], &feesAndAmounts, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	require.Empty(t, input.GravityKeeper.GetUnbatchedTransactions(ctx))
}

func TestRemoveFromOutgoingPoolAndRefundCosmosOriginated(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context

	var (
		mySender, e1        = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
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
	input.GravityKeeper.setCosmosOriginatedDenomToERC20(ctx, myTokenDenom, *tokenAddr)

	isCosmosOriginated, addr, err := input.GravityKeeper.DenomToERC20Lookup(ctx, myTokenDenom)
	require.True(t, isCosmosOriginated)
	require.NoError(t, err)
	require.Equal(t, tokenAddr.GetAddress().Hex(), myTokenContractAddr)
	require.Equal(t, tokenAddr, addr)

	allVouchers := sdk.Coins{sdk.NewCoin(myTokenDenom, sdk.NewIntFromUint64(originalBal))}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	// Create unbatched transactions
	require.Empty(t, input.GravityKeeper.GetUnbatchedTransactions(ctx))
	feesAndAmounts := uint64(0)
	ids := make([]uint64, 4)
	fees := []uint64{2, 3, 2, 1}
	amounts := []uint64{100, 101, 102, 103}
	for i, v := range fees {
		amount := sdk.NewCoin(myTokenDenom, sdk.NewIntFromUint64(amounts[i]))
		fee := sdk.NewCoin(myTokenDenom, sdk.NewIntFromUint64(v))

		feesAndAmounts += v + amounts[i]
		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, fee)
		require.NoError(t, err)
		t.Logf("___ response: %#v", r)
		ids[i] = r
		// Should create:
		// 1: amount 100, fee 2
		// 2: amount 101, fee 3
		// 3: amount 102, fee 2
		// 4: amount 103, fee 1

	}
	// Check balance
	currentBal := input.BankKeeper.GetBalance(ctx, mySender, myTokenDenom).Amount.Uint64()
	require.Equal(t, currentBal, originalBal-feesAndAmounts)

	// Check that removing a transaction refunds the costs and the tx no longer exists in the pool
	checkRemovedTx(t, input, ctx, ids[2], fees[2], amounts[2], &feesAndAmounts, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	checkRemovedTx(t, input, ctx, ids[3], fees[3], amounts[3], &feesAndAmounts, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	checkRemovedTx(t, input, ctx, ids[1], fees[1], amounts[1], &feesAndAmounts, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	checkRemovedTx(t, input, ctx, ids[0], fees[0], amounts[0], &feesAndAmounts, originalBal, mySender, myTokenContractAddr, myTokenDenom)
	require.Empty(t, input.GravityKeeper.GetUnbatchedTransactions(ctx))
}

// Helper method to:
// 1. Remove the transaction specified by `id`, `myTokenContractAddr` and `fee`
// 2. Update the feesAndAmounts tracker by subtracting the refunded `fee` and `amount`
// 3. Require that `mySender` has been refunded the correct amount for the cancelled transaction
// 4. Require that the unbatched transaction pool does not contain the refunded transaction via iterating its elements
func checkRemovedTx(t *testing.T, input TestInput, ctx sdk.Context, id uint64, fee uint64, amount uint64,
	feesAndAmounts *uint64, originalBal uint64, mySender sdk.AccAddress, myTokenContractAddr string, myTokenDenom string) {
	err := input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, id, mySender)
	require.NoError(t, err)
	*feesAndAmounts -= fee + amount // user should have regained the locked amounts from tx
	currentBal := input.BankKeeper.GetBalance(ctx, mySender, myTokenDenom).Amount.Uint64()
	require.Equal(t, currentBal, originalBal-*feesAndAmounts)
	expectedKey := myTokenContractAddr + fmt.Sprint(fee) + fmt.Sprint(id)
	input.GravityKeeper.IterateUnbatchedTransactions(ctx, func(key []byte, tx *types.InternalOutgoingTransferTx) bool {
		require.NotEqual(t, []byte(expectedKey), key)
		found := id == tx.Id &&
			fee == tx.Erc20Fee.Amount.Uint64() &&
			amount == tx.Erc20Token.Amount.Uint64()
		require.False(t, found)
		return false
	})
}

// ======================== Edge case tests for RemoveFromOutgoingPoolAndRefund() =================================== //

// Checks some common edge cases like invalid inputs, user didn't submit the transaction, tx doesn't exist, inconsistent entry
func TestRefundInconsistentTx(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	var (
		mySender, e1            = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver, e2          = types.NewEthAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr, e3 = types.NewEthAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5")
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)

	//////// Refund an inconsistent tx ////////
	amountToken, err := types.NewInternalERC20Token(sdk.NewInt(100), myTokenContractAddr.GetAddress().Hex())
	require.NoError(t, err)
	badTokenContractAddr, err := types.NewEthAddress("0x429881672b9AE42b8eBA0e26cd9c73711b891ca6") // different last char
	require.NoError(t, err)
	badFeeToken, err := types.NewInternalERC20Token(sdk.NewInt(2), badTokenContractAddr.GetAddress().Hex())
	require.NoError(t, err)

	// This way should fail
	r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *myReceiver, amountToken.GravityCoin(), badFeeToken.GravityCoin())
	require.Zero(t, r)
	require.Error(t, err)
	// But this unsafe override won't fail
	err = input.GravityKeeper.addUnbatchedTX(ctx, &types.InternalOutgoingTransferTx{
		Id:          uint64(5),
		Sender:      mySender,
		DestAddress: myReceiver,
		Erc20Token:  amountToken,
		Erc20Fee:    badFeeToken,
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
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	var (
		mySender, e1 = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
	)
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

	//////// Refund a tx twice ////////

	// mint some voucher first
	originalBal := uint64(99999)
	allVouchersToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(originalBal), myTokenContractAddr)
	require.NoError(t, err)
	allVouchers := sdk.Coins{allVouchersToken.GravityCoin()}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	amountToken, err := types.NewInternalERC20Token(sdk.NewInt(100), myTokenContractAddr)
	require.NoError(t, err)
	feeToken, err := types.NewInternalERC20Token(sdk.NewInt(2), myTokenContractAddr)
	require.NoError(t, err)
	origBalances := input.BankKeeper.GetAllBalances(ctx, mySender)

	txId, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amountToken.GravityCoin(), feeToken.GravityCoin())
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
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context

	// token1
	var (
		mySender1, e1                       = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		mySender2            sdk.AccAddress = []byte("gravity1ahx7f8wyertus")
		myReceiver                          = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		myTokenContractAddr1                = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5"
		myTokenContractAddr2                = "0x429881672b9AE42b8eBA0e26cd9c73711b891ca6"
	)
	require.NoError(t, e1)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract1, err := types.NewEthAddress(myTokenContractAddr1)
	require.NoError(t, err)
	tokenContract2, err := types.NewEthAddress(myTokenContractAddr2)
	require.NoError(t, err)
	// mint some vouchers first
	allVouchersToken1, err := types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr1)
	require.NoError(t, err)
	allVouchers1 := sdk.Coins{allVouchersToken1.GravityCoin()}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers1)
	require.NoError(t, err)
	allVouchersToken2, err := types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr2)
	require.NoError(t, err)
	allVouchers2 := sdk.Coins{allVouchersToken2.GravityCoin()}
	require.NoError(t, err)
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers2)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender1)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender1, allVouchers1)
	require.NoError(t, err)
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender2)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender2, allVouchers2)
	require.NoError(t, err)

	ids1 := make([]uint64, 4)
	ids2 := make([]uint64, 4)
	fees := []uint64{2, 3, 2, 1}
	amounts := []uint64{100, 101, 102, 103}
	idToTxMap := make(map[uint64]*types.OutgoingTransferTx)
	for i, v := range fees {
		amountToken1, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(amounts[i]), myTokenContractAddr1)
		require.NoError(t, err)
		amount1 := amountToken1.GravityCoin()
		feeToken1, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr1)
		require.NoError(t, err)
		fee1 := feeToken1.GravityCoin()

		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender1, *receiver, amount1, fee1)
		require.NoError(t, err)
		ids1[i] = r
		idToTxMap[r] = &types.OutgoingTransferTx{
			Id:          r,
			Sender:      mySender1.String(),
			DestAddress: myReceiver,
			Erc20Token:  amountToken1.ToExternal(),
			Erc20Fee:    feeToken1.ToExternal(),
		}
		amountToken2, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(amounts[i]), myTokenContractAddr2)
		require.NoError(t, err)
		amount2 := amountToken2.GravityCoin()
		feeToken2, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr2)
		require.NoError(t, err)
		fee2 := feeToken2.GravityCoin()

		r, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender2, *receiver, amount2, fee2)
		require.NoError(t, err)
		ids2[i] = r
		idToTxMap[r] = &types.OutgoingTransferTx{
			Id:          r,
			Sender:      mySender2.String(),
			DestAddress: myReceiver,
			Erc20Token:  amountToken2.ToExternal(),
			Erc20Fee:    feeToken2.ToExternal(),
		}
	}

	// GetUnbatchedTxByFeeAndId
	token1Fee, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(fees[0]), myTokenContractAddr1)
	require.NoError(t, err)
	token1Amount, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(amounts[0]), myTokenContractAddr1)
	require.NoError(t, err)
	token1Id := ids1[0]
	tx1, err1 := input.GravityKeeper.GetUnbatchedTxByFeeAndId(ctx, *token1Fee, token1Id)
	require.NoError(t, err1)
	expTx1, err1 := types.NewInternalOutgoingTransferTx(token1Id, mySender1.String(), myReceiver, token1Amount.ToExternal(), token1Fee.ToExternal())
	require.NoError(t, err1)
	require.Equal(t, *expTx1, *tx1)

	token2Fee, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(fees[3]), myTokenContractAddr2)
	require.NoError(t, err)
	token2Amount, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(amounts[3]), myTokenContractAddr2)
	require.NoError(t, err)

	token2Id := ids2[3]
	tx2, err2 := input.GravityKeeper.GetUnbatchedTxByFeeAndId(ctx, *token2Fee, token2Id)
	require.NoError(t, err2)
	expTx2, err2 := types.NewInternalOutgoingTransferTx(token2Id, mySender2.String(), myReceiver, token2Amount.ToExternal(), token2Fee.ToExternal())
	require.NoError(t, err2)
	require.Equal(t, *expTx2, *tx2)

	// GetUnbatchedTxById
	tx1, err1 = input.GravityKeeper.GetUnbatchedTxById(ctx, token1Id)
	require.NoError(t, err1)
	require.Equal(t, *expTx1, *tx1)

	tx2, err2 = input.GravityKeeper.GetUnbatchedTxById(ctx, token2Id)
	require.NoError(t, err2)
	require.Equal(t, *expTx2, *tx2)

	// GetUnbatchedTransactionsByContract
	token1Txs := input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *tokenContract1)
	for _, v := range token1Txs {
		expTx := idToTxMap[v.Id]
		require.NotNil(t, expTx)
		require.Equal(t, myTokenContractAddr1, v.Erc20Fee.Contract.GetAddress().Hex())
		require.Equal(t, myTokenContractAddr1, v.Erc20Token.Contract.GetAddress().Hex())
		require.Equal(t, expTx.DestAddress, v.DestAddress.GetAddress().Hex())
		require.Equal(t, expTx.Sender, v.Sender.String())
	}
	token2Txs := input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *tokenContract2)
	for _, v := range token2Txs {
		expTx := idToTxMap[v.Id]
		require.NotNil(t, expTx)
		require.Equal(t, myTokenContractAddr2, v.Erc20Fee.Contract.GetAddress().Hex())
		require.Equal(t, myTokenContractAddr2, v.Erc20Token.Contract.GetAddress().Hex())
		require.Equal(t, expTx.DestAddress, v.DestAddress.GetAddress().Hex())
		require.Equal(t, expTx.Sender, v.Sender.String())
	}
	// GetUnbatchedTransactions
	allTxs := input.GravityKeeper.GetUnbatchedTransactions(ctx)
	for _, v := range allTxs {
		expTx := idToTxMap[v.Id]
		require.NotNil(t, expTx)
		require.Equal(t, expTx.DestAddress, v.DestAddress.GetAddress().Hex())
		require.Equal(t, expTx.Sender, v.Sender.String())
		require.Equal(t, expTx.Erc20Fee.Contract, v.Erc20Fee.Contract.GetAddress().Hex())
		require.Equal(t, expTx.Erc20Token.Contract, v.Erc20Token.Contract.GetAddress().Hex())
	}
}

// Check the various iteration methods for the pool
func TestIterateUnbatchedTransactions(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context

	// token1
	var (
		mySender1, e1                       = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		mySender2            sdk.AccAddress = []byte("gravity1ahx7f8wyertus")
		myReceiver                          = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		myTokenContractAddr1                = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5"
		myTokenContractAddr2                = "0x429881672b9AE42b8eBA0e26cd9c73711b891ca6"
	)
	require.NoError(t, e1)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract1, err := types.NewEthAddress(myTokenContractAddr1)
	require.NoError(t, err)
	tokenContract2, err := types.NewEthAddress(myTokenContractAddr2)
	require.NoError(t, err)
	// mint some vouchers first
	token1, err := types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr1)
	require.NoError(t, err)
	allVouchers1 := sdk.Coins{token1.GravityCoin()}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers1)
	require.NoError(t, err)

	token2, err := types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr2)
	require.NoError(t, err)
	allVouchers2 := sdk.Coins{token2.GravityCoin()}
	err = input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers2)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender1)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender1, allVouchers1)
	require.NoError(t, err)
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender2)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender2, allVouchers2)
	require.NoError(t, err)

	ids1 := make([]uint64, 4)
	ids2 := make([]uint64, 4)
	fees := []uint64{2, 3, 2, 1}
	amounts := []uint64{100, 101, 102, 103}
	idToTxMap := make(map[uint64]*types.OutgoingTransferTx)
	for i, v := range fees {
		amount1, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(amounts[i]), myTokenContractAddr1)
		require.NoError(t, err)
		fee1, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr1)
		require.NoError(t, err)
		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender1, *receiver, amount1.GravityCoin(), fee1.GravityCoin())
		require.NoError(t, err)
		ids1[i] = r
		idToTxMap[r] = &types.OutgoingTransferTx{
			Id:          r,
			Sender:      mySender1.String(),
			DestAddress: myReceiver,
			Erc20Token:  amount1.ToExternal(),
			Erc20Fee:    fee1.ToExternal(),
		}
		amount2, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(amounts[i]), myTokenContractAddr2)
		require.NoError(t, err)
		fee2, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr2)
		require.NoError(t, err)
		r, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender2, *receiver, amount2.GravityCoin(), fee2.GravityCoin())
		require.NoError(t, err)

		ids2[i] = r
		idToTxMap[r] = &types.OutgoingTransferTx{
			Id:          r,
			Sender:      mySender2.String(),
			DestAddress: myReceiver,
			Erc20Token:  amount2.ToExternal(),
			Erc20Fee:    fee2.ToExternal(),
		}
	}
	// IterateUnbatchedTransactionsByContract
	foundMap := make(map[uint64]bool)
	input.GravityKeeper.IterateUnbatchedTransactionsByContract(ctx, *tokenContract1, func(key []byte, tx *types.InternalOutgoingTransferTx) bool {
		require.NotNil(t, tx)
		fTx := idToTxMap[tx.Id]
		require.NotNil(t, fTx)
		require.Equal(t, fTx.Erc20Fee.Contract, myTokenContractAddr1)
		require.Equal(t, fTx.Erc20Token.Contract, myTokenContractAddr1)
		require.Equal(t, fTx.DestAddress, myReceiver)
		require.Equal(t, mySender1.String(), fTx.Sender)
		foundMap[fTx.Id] = true
		return false
	})
	input.GravityKeeper.IterateUnbatchedTransactionsByContract(ctx, *tokenContract2, func(key []byte, tx *types.InternalOutgoingTransferTx) bool {
		require.NotNil(t, tx)
		fTx := idToTxMap[tx.Id]
		require.NotNil(t, fTx)
		require.Equal(t, fTx.Erc20Fee.Contract, myTokenContractAddr2)
		require.Equal(t, fTx.Erc20Token.Contract, myTokenContractAddr2)
		require.Equal(t, fTx.DestAddress, myReceiver)
		require.Equal(t, mySender2.String(), fTx.Sender)
		foundMap[fTx.Id] = true
		return false
	})

	for i := 1; i <= 8; i++ {
		require.True(t, foundMap[uint64(i)])
	}
	// filterAndIterateUnbatchedTransactions
	anotherFoundMap := make(map[uint64]bool)
	input.GravityKeeper.IterateUnbatchedTransactions(ctx, func(key []byte, tx *types.InternalOutgoingTransferTx) bool {
		require.NotNil(t, tx)
		fTx := idToTxMap[tx.Id]
		require.NotNil(t, fTx)
		require.Equal(t, fTx.DestAddress, tx.DestAddress.GetAddress().Hex())

		anotherFoundMap[fTx.Id] = true
		return false
	})

	for i := 1; i <= 8; i++ {
		require.True(t, anotherFoundMap[uint64(i)])
	}
}

// Ensures that any unbatched tx will make its way into the exported data from ExportGenesis
func TestAddToOutgoingPoolExportGenesis(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	k := input.GravityKeeper
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

	unbatchedTxMap := make(map[uint64]types.OutgoingTransferTx)
	foundTxsMap := make(map[uint64]bool)
	// when
	for i, v := range []uint64{2, 3, 2, 1} {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr)
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr)
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		r, err := input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, fee)
		require.NoError(t, err)

		unbatchedTxMap[r] = types.OutgoingTransferTx{
			Id:          r,
			Sender:      mySender.String(),
			DestAddress: myReceiver,
			Erc20Token:  amountToken.ToExternal(),
			Erc20Fee:    feeToken.ToExternal(),
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
