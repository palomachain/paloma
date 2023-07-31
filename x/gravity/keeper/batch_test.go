package keeper

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint: exhaustruct
func TestBatches(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	var (
		now                     = time.Now().UTC()
		mySender, e1            = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver, e2          = types.NewEthAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr, e3 = types.NewEthAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5") // Pickle
		token, e4               = types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr.GetAddress().Hex())
		allVouchers             = sdk.NewCoins(token.GravityCoin())
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	require.NoError(t, e4)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// CREATE FIRST BATCH
	// ==================

	// batch should not be created if there is no txs of the given token type in tx pool
	noBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, *myTokenContractAddr, 2)
	require.Nil(t, noBatch)
	require.Error(t, err)

	// add some TX to the pool
	for i, v := range []uint64{2, 3, 2, 1} {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr.GetAddress().Hex())
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr.GetAddress().Hex())
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *myReceiver, amount, fee)
		require.NoError(t, err)
		ctx.Logger().Info(fmt.Sprintf("Created transaction %v with amount %v and fee %v", i, amount, fee))
		// Should create:
		// 1: tx amount is 100, fee is 2, id is 1
		// 2: tx amount is 101, fee is 3, id is 2
		// 3: tx amount is 102, fee is 2, id is 3
		// 4: tx amount is 103, fee is 1, id is 4
	}

	// when
	ctx = ctx.WithBlockTime(now)
	input.GravityKeeper.SetLastObservedEthereumBlockHeight(ctx, 1234567)
	// maxElements must be greater then 0, otherwise the batch would not be created
	noBatch, err = input.GravityKeeper.BuildOutgoingTXBatch(ctx, *myTokenContractAddr, 0)
	require.Nil(t, noBatch)
	require.Error(t, err)

	// tx batch size is 2, so that some of them stay behind
	firstBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, *myTokenContractAddr, 2)
	require.NoError(t, err)

	// then batch is persisted
	gotFirstBatch := input.GravityKeeper.GetOutgoingTXBatch(ctx, firstBatch.TokenContract, firstBatch.BatchNonce)
	require.NotNil(t, gotFirstBatch)
	// Should have txs 2: and 3: from above, as ties in fees are broken by transaction index
	ctx.Logger().Info(fmt.Sprintf("found batch %+v", gotFirstBatch))

	expFirstBatch := types.OutgoingTxBatch{
		BatchNonce: 1,
		Transactions: []types.OutgoingTransferTx{
			{
				Id:          2,
				Erc20Fee:    types.NewERC20Token(3, myTokenContractAddr.GetAddress().Hex()),
				Sender:      mySender.String(),
				DestAddress: myReceiver.GetAddress().Hex(),
				Erc20Token:  types.NewERC20Token(101, myTokenContractAddr.GetAddress().Hex()),
			},
			{
				Id:          3,
				Erc20Fee:    types.NewERC20Token(2, myTokenContractAddr.GetAddress().Hex()),
				Sender:      mySender.String(),
				DestAddress: myReceiver.GetAddress().Hex(),
				Erc20Token:  types.NewERC20Token(102, myTokenContractAddr.GetAddress().Hex()),
			},
		},
		TokenContract:      myTokenContractAddr.GetAddress().Hex(),
		CosmosBlockCreated: 1234567,
		BatchTimeout:       input.GravityKeeper.getBatchTimeoutHeight(ctx),
	}
	assert.Equal(t, expFirstBatch.BatchTimeout, gotFirstBatch.BatchTimeout)
	assert.Equal(t, expFirstBatch.BatchNonce, gotFirstBatch.BatchNonce)
	assert.Equal(t, expFirstBatch.CosmosBlockCreated, gotFirstBatch.CosmosBlockCreated)
	assert.Equal(t, expFirstBatch.TokenContract, gotFirstBatch.TokenContract.GetAddress().Hex())
	assert.Equal(t, len(expFirstBatch.Transactions), len(gotFirstBatch.Transactions))
	for i := 0; i < len(expFirstBatch.Transactions); i++ {
		assert.Equal(t, expFirstBatch.Transactions[i], gotFirstBatch.Transactions[i].ToExternal())
	}

	// persist confirmations for first batch
	for i, orch := range OrchAddrs {
		ethAddr, err := types.NewEthAddress(EthAddrs[i].String())
		require.NoError(t, err)

		conf := &types.MsgConfirmBatch{
			Nonce:         firstBatch.BatchNonce,
			TokenContract: firstBatch.TokenContract.GetAddress().Hex(),
			EthSigner:     ethAddr.GetAddress().Hex(),
			Orchestrator:  orch.String(),
			Signature:     "dummysig",
		}

		input.GravityKeeper.SetBatchConfirm(ctx, conf)
	}

	// verify that confirms are persisted
	firstBatchConfirms := input.GravityKeeper.GetBatchConfirmByNonceAndTokenContract(ctx, firstBatch.BatchNonce, firstBatch.TokenContract)
	require.Equal(t, len(OrchAddrs), len(firstBatchConfirms))

	// and verify remaining available Tx in the pool
	// Should still have 1: and 4: above
	gotUnbatchedTx := input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *myTokenContractAddr)
	oneFee, err := types.NewInternalERC20Token(sdk.NewInt(1), myTokenContractAddr.GetAddress().Hex())
	require.NoError(t, err)
	oneHundredTok, err := types.NewInternalERC20Token(sdk.NewInt(100), myTokenContractAddr.GetAddress().Hex())
	require.NoError(t, err)
	twoFee, err := types.NewInternalERC20Token(sdk.NewInt(2), myTokenContractAddr.GetAddress().Hex())
	require.NoError(t, err)
	oneHundredThreeTok, err := types.NewInternalERC20Token(sdk.NewInt(103), myTokenContractAddr.GetAddress().Hex())
	require.NoError(t, err)
	expUnbatchedTx := []*types.InternalOutgoingTransferTx{
		{
			Id:          1,
			Erc20Fee:    twoFee,
			Sender:      mySender,
			DestAddress: myReceiver,
			Erc20Token:  oneHundredTok,
		},
		{
			Id:          4,
			Erc20Fee:    oneFee,
			Sender:      mySender,
			DestAddress: myReceiver,
			Erc20Token:  oneHundredThreeTok,
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)

	// CREATE SECOND, MORE PROFITABLE BATCH
	// ====================================

	// first check that less profitable batch cannot be created
	noBatch, err = input.GravityKeeper.BuildOutgoingTXBatch(ctx, *myTokenContractAddr, 2)
	require.Nil(t, noBatch)
	require.Error(t, err)

	// add some more TX to the pool to create a more profitable batch
	for i, v := range []uint64{4, 5} {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr.GetAddress().Hex())
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr.GetAddress().Hex())
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *myReceiver, amount, fee)
		require.NoError(t, err)
		// Creates the following:
		// 5: amount 100, fee 4, id 5
		// 6: amount 101, fee 5, id 6
	}

	// create the more profitable batch
	ctx = ctx.WithBlockTime(now)
	// tx batch size is 2, so that some of them stay behind
	secondBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, *myTokenContractAddr, 2)
	require.NoError(t, err)

	input.GravityKeeper.SetLastObservedEthereumBlockHeight(ctx, 1234567)
	// check that the more profitable batch has the right txs in it
	// Should only have 5: and 6: above
	expSecondBatch := types.OutgoingTxBatch{
		BatchNonce: 2,
		Transactions: []types.OutgoingTransferTx{
			{
				Id:          6,
				Erc20Fee:    types.NewERC20Token(5, myTokenContractAddr.GetAddress().Hex()),
				Sender:      mySender.String(),
				DestAddress: myReceiver.GetAddress().Hex(),
				Erc20Token:  types.NewERC20Token(101, myTokenContractAddr.GetAddress().Hex()),
			},
			{
				Id:          5,
				Erc20Fee:    types.NewERC20Token(4, myTokenContractAddr.GetAddress().Hex()),
				Sender:      mySender.String(),
				DestAddress: myReceiver.GetAddress().Hex(),
				Erc20Token:  types.NewERC20Token(100, myTokenContractAddr.GetAddress().Hex()),
			},
		},
		TokenContract:      myTokenContractAddr.GetAddress().Hex(),
		CosmosBlockCreated: 1234567,
		BatchTimeout:       input.GravityKeeper.getBatchTimeoutHeight(ctx),
	}

	assert.Equal(t, expSecondBatch.BatchTimeout, secondBatch.BatchTimeout)
	assert.Equal(t, expSecondBatch.BatchNonce, secondBatch.BatchNonce)
	assert.Equal(t, expSecondBatch.CosmosBlockCreated, secondBatch.CosmosBlockCreated)
	assert.Equal(t, expSecondBatch.TokenContract, secondBatch.TokenContract.GetAddress().Hex())
	assert.Equal(t, len(expSecondBatch.Transactions), len(secondBatch.Transactions))
	for i := 0; i < len(expSecondBatch.Transactions); i++ {
		assert.Equal(t, expSecondBatch.Transactions[i], secondBatch.Transactions[i].ToExternal())
	}

	// persist confirmations for second batch
	for i, orch := range OrchAddrs {
		ethAddr, err := types.NewEthAddress(EthAddrs[i].String())
		require.NoError(t, err)

		conf := &types.MsgConfirmBatch{
			Nonce:         secondBatch.BatchNonce,
			TokenContract: secondBatch.TokenContract.GetAddress().Hex(),
			EthSigner:     ethAddr.GetAddress().Hex(),
			Orchestrator:  orch.String(),
			Signature:     "dummysig",
		}

		input.GravityKeeper.SetBatchConfirm(ctx, conf)
	}

	// verify that confirms are persisted
	secondBatchConfirms := input.GravityKeeper.GetBatchConfirmByNonceAndTokenContract(ctx, secondBatch.BatchNonce, secondBatch.TokenContract)
	require.Equal(t, len(OrchAddrs), len(secondBatchConfirms))

	// check that last added batch is the one with the biggest nonce
	lastOutgoingBatch := input.GravityKeeper.GetLastOutgoingBatchByTokenType(ctx, *myTokenContractAddr)
	require.NotNil(t, lastOutgoingBatch)
	assert.Equal(t, lastOutgoingBatch.BatchNonce, secondBatch.BatchNonce)

	// EXECUTE THE MORE PROFITABLE BATCH
	// =================================

	// Execute the batch
	fakeBlock := secondBatch.CosmosBlockCreated // A fake ethereum block used for testing only
	msg := types.MsgBatchSendToEthClaim{EthBlockHeight: fakeBlock, BatchNonce: secondBatch.BatchNonce}
	input.GravityKeeper.OutgoingTxBatchExecuted(ctx, secondBatch.TokenContract, msg)

	// check batch has been deleted
	gotSecondBatch := input.GravityKeeper.GetOutgoingTXBatch(ctx, secondBatch.TokenContract, secondBatch.BatchNonce)
	require.Nil(t, gotSecondBatch)
	// check batch confirmations have been deleted
	secondBatchConfirms = input.GravityKeeper.GetBatchConfirmByNonceAndTokenContract(ctx, secondBatch.BatchNonce, secondBatch.TokenContract)
	require.Equal(t, 0, len(secondBatchConfirms))

	// check that txs from first batch have been freed
	gotUnbatchedTx = input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *myTokenContractAddr)
	threeFee, err := types.NewInternalERC20Token(sdk.NewInt(3), myTokenContractAddr.GetAddress().Hex())
	require.NoError(t, err)
	oneHundredOneTok, err := types.NewInternalERC20Token(sdk.NewInt(101), myTokenContractAddr.GetAddress().Hex())
	require.NoError(t, err)
	oneHundredTwoTok, err := types.NewInternalERC20Token(sdk.NewInt(102), myTokenContractAddr.GetAddress().Hex())
	require.NoError(t, err)
	expUnbatchedTx = []*types.InternalOutgoingTransferTx{
		{
			Id:          2,
			Erc20Fee:    threeFee,
			Sender:      mySender,
			DestAddress: myReceiver,
			Erc20Token:  oneHundredOneTok,
		},
		{
			Id:          3,
			Erc20Fee:    twoFee,
			Sender:      mySender,
			DestAddress: myReceiver,
			Erc20Token:  oneHundredTwoTok,
		},
		{
			Id:          1,
			Erc20Fee:    twoFee,
			Sender:      mySender,
			DestAddress: myReceiver,
			Erc20Token:  oneHundredTok,
		},
		{
			Id:          4,
			Erc20Fee:    oneFee,
			Sender:      mySender,
			DestAddress: myReceiver,
			Erc20Token:  oneHundredThreeTok,
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)

	// check that first batch has been deleted
	gotFirstBatch = input.GravityKeeper.GetOutgoingTXBatch(ctx, firstBatch.TokenContract, firstBatch.BatchNonce)
	require.Nil(t, gotFirstBatch)
	// check that first batch confirmations have been deleted
	firstBatchConfirms = input.GravityKeeper.GetBatchConfirmByNonceAndTokenContract(ctx, firstBatch.BatchNonce, firstBatch.TokenContract)
	require.Equal(t, 0, len(firstBatchConfirms))
}

// tests that batches work with large token amounts, mostly a duplicate of the above
// tests but using much bigger numbers
// nolint: exhaustruct
func TestBatchesFullCoins(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	var (
		now                 = time.Now().UTC()
		mySender, e1        = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver          = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		receiverAddr, e2    = types.NewEthAddress(myReceiver)
		myTokenContractAddr = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5"   // Pickle
		totalCoins, _       = sdk.NewIntFromString("1500000000000000000000") // 1,500 ETH worth
		oneEth, _           = sdk.NewIntFromString("1000000000000000000")
		token, e3           = types.NewInternalERC20Token(totalCoins, myTokenContractAddr)
		allVouchers         = sdk.NewCoins(token.GravityCoin())
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	tokenContract, err := types.NewEthAddress(myTokenContractAddr)
	require.NoError(t, err)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// CREATE FIRST BATCH
	// ==================

	// add some TX to the pool
	for _, v := range []uint64{20, 300, 25, 10} {
		vAsSDKInt := sdk.NewIntFromUint64(v)
		amountToken, err := types.NewInternalERC20Token(oneEth.Mul(vAsSDKInt), myTokenContractAddr)
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(oneEth.Mul(vAsSDKInt), myTokenContractAddr)
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiverAddr, amount, fee)
		require.NoError(t, err)
	}

	// when
	ctx = ctx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	firstBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, *tokenContract, 2)
	require.NoError(t, err)

	// then batch is persisted
	gotFirstBatch := input.GravityKeeper.GetOutgoingTXBatch(ctx, firstBatch.TokenContract, firstBatch.BatchNonce)
	require.NotNil(t, gotFirstBatch)

	expFirstBatch := &types.OutgoingTxBatch{
		BatchNonce: 1,
		Transactions: []types.OutgoingTransferTx{
			{
				Id:          2,
				Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(300)), myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver,
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(300)), myTokenContractAddr),
			},
			{
				Id:          3,
				Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(25)), myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver,
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(25)), myTokenContractAddr),
			},
		},
		TokenContract:      myTokenContractAddr,
		CosmosBlockCreated: 1234567,
	}
	assert.Equal(t, expFirstBatch.BatchTimeout, gotFirstBatch.BatchTimeout)
	assert.Equal(t, expFirstBatch.BatchNonce, gotFirstBatch.BatchNonce)
	assert.Equal(t, expFirstBatch.CosmosBlockCreated, gotFirstBatch.CosmosBlockCreated)
	assert.Equal(t, expFirstBatch.TokenContract, gotFirstBatch.TokenContract.GetAddress().Hex())
	assert.Equal(t, len(expFirstBatch.Transactions), len(gotFirstBatch.Transactions))
	for i := 0; i < len(expFirstBatch.Transactions); i++ {
		assert.Equal(t, expFirstBatch.Transactions[i], gotFirstBatch.Transactions[i].ToExternal())
	}

	// and verify remaining available Tx in the pool
	gotUnbatchedTx := input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *tokenContract)
	twentyTok, err := types.NewInternalERC20Token(oneEth.Mul(sdk.NewIntFromUint64(20)), myTokenContractAddr)
	require.NoError(t, err)
	tenTok, err := types.NewInternalERC20Token(oneEth.Mul(sdk.NewIntFromUint64(10)), myTokenContractAddr)
	require.NoError(t, err)
	expUnbatchedTx := []*types.InternalOutgoingTransferTx{
		{
			Id:          1,
			Erc20Fee:    twentyTok,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  twentyTok,
		},
		{
			Id:          4,
			Erc20Fee:    tenTok,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  tenTok,
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)

	// CREATE SECOND, MORE PROFITABLE BATCH
	// ====================================

	// add some more TX to the pool to create a more profitable batch
	for _, v := range []uint64{200, 150} {
		vAsSDKInt := sdk.NewIntFromUint64(v)
		amountToken, err := types.NewInternalERC20Token(oneEth.Mul(vAsSDKInt), myTokenContractAddr)
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(oneEth.Mul(vAsSDKInt), myTokenContractAddr)
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiverAddr, amount, fee)
		require.NoError(t, err)
	}

	// create the more profitable batch
	ctx = ctx.WithBlockTime(now)
	input.GravityKeeper.SetLastObservedEthereumBlockHeight(ctx, 1234567)
	// tx batch size is 2, so that some of them stay behind
	secondBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, *tokenContract, 2)
	require.NoError(t, err)

	// check that the more profitable batch has the right txs in it
	expSecondBatch := &types.OutgoingTxBatch{
		BatchNonce: 2,
		Transactions: []types.OutgoingTransferTx{
			{
				Id:          5,
				Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(200)), myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver,
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(200)), myTokenContractAddr),
			},
			{
				Id:          6,
				Erc20Fee:    types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(150)), myTokenContractAddr),
				Sender:      mySender.String(),
				DestAddress: myReceiver,
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(sdk.NewIntFromUint64(150)), myTokenContractAddr),
			},
		},
		TokenContract:      myTokenContractAddr,
		CosmosBlockCreated: 1234567,
		BatchTimeout:       input.GravityKeeper.getBatchTimeoutHeight(ctx),
	}

	assert.Equal(t, expSecondBatch.BatchTimeout, secondBatch.BatchTimeout)
	assert.Equal(t, expSecondBatch.BatchNonce, secondBatch.BatchNonce)
	assert.Equal(t, expSecondBatch.CosmosBlockCreated, secondBatch.CosmosBlockCreated)
	assert.Equal(t, expSecondBatch.TokenContract, secondBatch.TokenContract.GetAddress().Hex())
	assert.Equal(t, len(expSecondBatch.Transactions), len(secondBatch.Transactions))
	for i := 0; i < len(expSecondBatch.Transactions); i++ {
		assert.Equal(t, expSecondBatch.Transactions[i], secondBatch.Transactions[i].ToExternal())
	}

	// EXECUTE THE MORE PROFITABLE BATCH
	// =================================

	// Execute the batch
	fakeBlock := secondBatch.CosmosBlockCreated // A fake ethereum block used for testing only
	msg := types.MsgBatchSendToEthClaim{EthBlockHeight: fakeBlock, BatchNonce: secondBatch.BatchNonce}
	input.GravityKeeper.OutgoingTxBatchExecuted(ctx, secondBatch.TokenContract, msg)

	// check batch has been deleted
	gotSecondBatch := input.GravityKeeper.GetOutgoingTXBatch(ctx, secondBatch.TokenContract, secondBatch.BatchNonce)
	require.Nil(t, gotSecondBatch)

	// check that txs from first batch have been freed
	gotUnbatchedTx = input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *tokenContract)
	threeHundredTok, err := types.NewInternalERC20Token(oneEth.Mul(sdk.NewIntFromUint64(300)), myTokenContractAddr)
	require.NoError(t, err)
	twentyFiveTok, err := types.NewInternalERC20Token(oneEth.Mul(sdk.NewIntFromUint64(25)), myTokenContractAddr)
	require.NoError(t, err)
	expUnbatchedTx = []*types.InternalOutgoingTransferTx{
		{
			Id:          2,
			Erc20Fee:    threeHundredTok,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  threeHundredTok,
		},
		{
			Id:          3,
			Erc20Fee:    twentyFiveTok,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  twentyFiveTok,
		},
		{
			Id:          1,
			Erc20Fee:    twentyTok,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  twentyTok,
		},
		{
			Id:          4,
			Erc20Fee:    tenTok,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  tenTok,
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)
}

// TestManyBatches handles test cases around batch execution, specifically executing multiple batches
// out of sequential order, which is exactly what happens on the
// nolint: exhaustruct
func TestManyBatches(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	var (
		now                = time.Now().UTC()
		mySender, e1       = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver         = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		tokenContractAddr1 = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5"
		tokenContractAddr2 = "0xF815240800ddf3E0be80e0d848B13ecaa504BF37"
		tokenContractAddr3 = "0xd086dDA7BccEB70e35064f540d07E4baED142cB3"
		tokenContractAddr4 = "0x384981B9d133701c4bD445F77bF61C3d80e79D46"
		totalCoins, _      = sdk.NewIntFromString("1500000000000000000000000")
		oneEth, _          = sdk.NewIntFromString("1000000000000000000")
		token1, e2         = types.NewInternalERC20Token(totalCoins, tokenContractAddr1)
		token2, e3         = types.NewInternalERC20Token(totalCoins, tokenContractAddr2)
		token3, e4         = types.NewInternalERC20Token(totalCoins, tokenContractAddr3)
		token4, e5         = types.NewInternalERC20Token(totalCoins, tokenContractAddr4)
		allVouchers        = sdk.NewCoins(
			token1.GravityCoin(),
			token2.GravityCoin(),
			token3.GravityCoin(),
			token4.GravityCoin(),
		)
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	require.NoError(t, e4)
	require.NoError(t, e5)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)

	// mint vouchers first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))
	input.GravityKeeper.SetLastObservedEthereumBlockHeight(ctx, 1234567)

	// CREATE FIRST BATCH
	// ==================

	tokens := [4]string{tokenContractAddr1, tokenContractAddr2, tokenContractAddr3, tokenContractAddr4}

	// when
	ctx = ctx.WithBlockTime(now)
	var batches []types.OutgoingTxBatch

	for _, contract := range tokens {
		contractAddr, err := types.NewEthAddress(contract)
		require.NoError(t, err)
		for v := 1; v < 500; v++ {
			vAsSDKInt := sdk.NewIntFromUint64(uint64(v))
			amountToken, err := types.NewInternalERC20Token(oneEth.Mul(vAsSDKInt), contract)
			require.NoError(t, err)
			amount := amountToken.GravityCoin()
			feeToken, err := types.NewInternalERC20Token(oneEth.Mul(vAsSDKInt), contract)
			require.NoError(t, err)
			fee := feeToken.GravityCoin()

			_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, fee)
			require.NoError(t, err)
			// create batch after every 100 txs to be able to create more profitable batches
			if (v+1)%100 == 0 {
				batch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, *contractAddr, 100)
				require.NoError(t, err)
				batches = append(batches, batch.ToExternal())
			}
		}
	}

	for _, batch := range batches {
		// then batch is persisted
		contractAddr, err := types.NewEthAddress(batch.TokenContract)
		require.NoError(t, err)
		gotBatch := input.GravityKeeper.GetOutgoingTXBatch(ctx, *contractAddr, batch.BatchNonce)
		require.NotNil(t, gotBatch)
	}

	// EXECUTE BOTH BATCHES
	// =================================

	// shuffle batches to simulate out of order execution on Ethereum
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(batches), func(i, j int) { batches[i], batches[j] = batches[j], batches[i] })

	// Execute the batches, if there are any problems OutgoingTxBatchExecuted will panic
	for _, batch := range batches {
		contractAddr, err := types.NewEthAddress(batch.TokenContract)
		require.NoError(t, err)
		gotBatch := input.GravityKeeper.GetOutgoingTXBatch(ctx, *contractAddr, batch.BatchNonce)
		// we may have already deleted some of the batches in this list by executing later ones
		if gotBatch != nil {
			fakeBlock := batch.CosmosBlockCreated // A fake ethereum block used for testing only
			msg := types.MsgBatchSendToEthClaim{EthBlockHeight: fakeBlock, BatchNonce: batch.BatchNonce}
			input.GravityKeeper.OutgoingTxBatchExecuted(ctx, *contractAddr, msg)
		}
	}
}

// nolint: exhaustruct
func TestPoolTxRefund(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	var (
		now                 = time.Now().UTC()
		mySender, e1        = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		notMySender, e2     = sdk.AccAddressFromBech32("gravity1add7f8wyertuus9r20284ej0asrs085c8ajr0y")
		myReceiver          = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		myTokenContractAddr = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5" // Pickle
		token, e3           = types.NewInternalERC20Token(sdk.NewInt(414), myTokenContractAddr)
		allVouchers         = sdk.NewCoins(token.GravityCoin())
		denomToken, e4      = types.NewInternalERC20Token(sdk.NewInt(1), myTokenContractAddr)
		myDenom             = denomToken.GravityCoin().Denom
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	require.NoError(t, e4)
	contract, err := types.NewEthAddress(myTokenContractAddr)
	require.NoError(t, err)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// CREATE FIRST BATCH
	// ==================

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
		// Should have created:
		// 1: amount 100, fee 2
		// 2: amount 101, fee 3
		// 3: amount 102, fee 2
		// 4: amount 103, fee 1
	}

	// when
	ctx = ctx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	// Should have 2: and 3: from above
	_, err = input.GravityKeeper.BuildOutgoingTXBatch(ctx, *contract, 2)
	require.NoError(t, err)

	// try to refund a tx that's in a batch
	err1 := input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 3, mySender)
	require.Error(t, err1)

	// try to refund somebody else's tx
	err2 := input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 4, notMySender)
	require.Error(t, err2)

	// try to refund a tx that's in the pool
	err3 := input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 4, mySender)
	require.NoError(t, err3)

	// make sure refund was issued
	balances := input.BankKeeper.GetAllBalances(ctx, mySender)
	require.Equal(t, sdk.NewInt(104), balances.AmountOf(myDenom))
}

// nolint: exhaustruct
func TestBatchesNotCreatedWhenBridgePaused(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context

	// pause the bridge
	params := input.GravityKeeper.GetParams(ctx)
	params.BridgeActive = false
	input.GravityKeeper.SetParams(ctx, params)

	var (
		now                     = time.Now().UTC()
		mySender, e1            = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver, e2          = types.NewEthAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr, e3 = types.NewEthAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5") // Pickle
		token, e4               = types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr.GetAddress().Hex())
		allVouchers             = sdk.NewCoins(token.GravityCoin())
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	require.NoError(t, e4)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// CREATE FIRST BATCH
	// ==================

	// add some TX to the pool
	for i, v := range []uint64{2, 3, 2, 1} {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr.GetAddress().Hex())
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr.GetAddress().Hex())
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *myReceiver, amount, fee)
		require.NoError(t, err)
		ctx.Logger().Info(fmt.Sprintf("Created transaction %v with amount %v and fee %v", i, amount, fee))
		// Should create:
		// 1: tx amount is 100, fee is 2, id is 1
		// 2: tx amount is 101, fee is 3, id is 2
		// 3: tx amount is 102, fee is 2, id is 3
		// 4: tx amount is 103, fee is 1, id is 4
	}

	// when
	ctx = ctx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	_, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, *myTokenContractAddr, 2)
	require.Error(t, err)

	// then batch is persisted
	gotFirstBatch := input.GravityKeeper.GetOutgoingTXBatch(ctx, *myTokenContractAddr, 1)
	require.Nil(t, gotFirstBatch)

	// resume the bridge
	params.BridgeActive = true
	input.GravityKeeper.SetParams(ctx, params)

	// when
	ctx = ctx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	firstBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, *myTokenContractAddr, 2)
	require.NoError(t, err)

	// then batch is persisted
	gotFirstBatch = input.GravityKeeper.GetOutgoingTXBatch(ctx, firstBatch.TokenContract, firstBatch.BatchNonce)
	require.NotNil(t, gotFirstBatch)
}

// nolint: exhaustruct
// test that tokens on the blacklist do not enter batches
func TestEthereumBlacklistBatches(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	var (
		now                     = time.Now().UTC()
		mySender, e1            = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver, e2          = types.NewEthAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		blacklistedReceiver, e3 = types.NewEthAddress("0x4d16b9E4a27c3313440923fEfCd013178149A5bD")
		myTokenContractAddr, e4 = types.NewEthAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5") // Pickle
		token, e5               = types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr.GetAddress().Hex())
		allVouchers             = sdk.NewCoins(token.GravityCoin())
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	require.NoError(t, e4)
	require.NoError(t, e5)

	// add the blacklisted address to the blacklist
	params := input.GravityKeeper.GetParams(ctx)
	params.EthereumBlacklist = append(params.EthereumBlacklist, blacklistedReceiver.GetAddress().Hex())
	input.GravityKeeper.SetParams(ctx, params)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// CREATE FIRST BATCH
	// ==================

	// add some TX to the pool
	for i, v := range []uint64{2, 3, 2, 1, 5} {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr.GetAddress().Hex())
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr.GetAddress().Hex())
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		// one of the transactions should go to the blacklisted address
		if i == 4 {
			_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *blacklistedReceiver, amount, fee)
		} else {
			_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *myReceiver, amount, fee)
		}
		require.NoError(t, err)
		ctx.Logger().Info(fmt.Sprintf("Created transaction %v with amount %v and fee %v", i, amount, fee))
		// Should create:
		// 1: tx amount is 100, fee is 2, id is 1
		// 2: tx amount is 101, fee is 3, id is 2
		// 3: tx amount is 102, fee is 2, id is 3
		// 4: tx amount is 103, fee is 1, id is 4
		// 5: tx amount is 104, fee is 5, id is 5
	}

	// check that blacklisted tx fee is not insluded in profitability calculation
	currentFees := input.GravityKeeper.GetBatchFeeByTokenType(ctx, *myTokenContractAddr, 10)
	assert.NotNil(t, currentFees)
	assert.True(t, currentFees.TotalFees.Equal(sdk.NewInt(8)))

	// when
	ctx = ctx.WithBlockTime(now)

	// tx batch size is 10
	firstBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, *myTokenContractAddr, 10)
	require.NoError(t, err)

	// then batch is persisted
	gotFirstBatch := input.GravityKeeper.GetOutgoingTXBatch(ctx, firstBatch.TokenContract, firstBatch.BatchNonce)
	require.NotNil(t, gotFirstBatch)
	// Should have all from above except the banned dest
	ctx.Logger().Info(fmt.Sprintf("found batch %+v", gotFirstBatch))

	// should be 4 not 5 transactions
	assert.Equal(t, 4, len(gotFirstBatch.Transactions))
	// should not contain id 5
	for i := 0; i < len(gotFirstBatch.Transactions); i++ {
		assert.NotEqual(t, gotFirstBatch.Transactions[i].Id, 5)
	}

	// and verify remaining available Tx in the pool
	// should only be 5
	gotUnbatchedTx := input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *myTokenContractAddr)
	assert.Equal(t, gotUnbatchedTx[0].Id, uint64(5))

}

// tests total batch fee collected from all of the txs in the batch
// nolint: exhaustruct
func TestGetFees(t *testing.T) {

	txs := []types.OutgoingTransferTx{
		{Erc20Fee: types.ERC20Token{Amount: sdk.NewInt(1)}},
		{Erc20Fee: types.ERC20Token{Amount: sdk.NewInt(2)}},
		{Erc20Fee: types.ERC20Token{Amount: sdk.NewInt(3)}},
	}

	type batchFeesTuple struct {
		batch        types.OutgoingTxBatch
		expectedFees sdk.Int
	}

	batches := []batchFeesTuple{
		{types.OutgoingTxBatch{
			Transactions: []types.OutgoingTransferTx{}},
			sdk.NewInt(0),
		},
		{types.OutgoingTxBatch{
			Transactions: []types.OutgoingTransferTx{txs[0]}},
			sdk.NewInt(1),
		},
		{types.OutgoingTxBatch{
			Transactions: []types.OutgoingTransferTx{txs[0], txs[1], txs[2]}},
			sdk.NewInt(6),
		},
	}

	for _, val := range batches {
		if !val.batch.GetFees().Equal(val.expectedFees) {
			t.Errorf("Invalid total batch fees!")
		}
	}
}

func TestBatchConfirms(t *testing.T) {
	input := CreateTestEnv(t)
	ctx := input.Context
	var (
		now                     = time.Now().UTC()
		mySender, e1            = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver, e2          = types.NewEthAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr, e3 = types.NewEthAddress("0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5") // Pickle
		token, e4               = types.NewInternalERC20Token(sdk.NewInt(1000000), myTokenContractAddr.GetAddress().Hex())
		allVouchers             = sdk.NewCoins(token.GravityCoin())
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	require.NoError(t, e4)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// when
	ctx = ctx.WithBlockTime(now)

	// add batches with 1 tx to the pool
	for i := 1; i < 200; i++ {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr.GetAddress().Hex())
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(uint64(i+10)), myTokenContractAddr.GetAddress().Hex())
		require.NoError(t, err)
		fee := feeToken.GravityCoin()

		// add tx to the pool
		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *myReceiver, amount, fee)
		require.NoError(t, err)
		ctx.Logger().Info(fmt.Sprintf("Created transaction %v with amount %v and fee %v", i, amount, fee))

		// create batch
		_, err = input.GravityKeeper.BuildOutgoingTXBatch(ctx, *myTokenContractAddr, 1)
		require.NoError(t, err)
	}

	outogoingBatches := input.GravityKeeper.GetOutgoingTxBatches(ctx)

	// persist confirmations
	for i, orch := range OrchAddrs {
		for _, batch := range outogoingBatches {
			ethAddr, err := types.NewEthAddress(EthAddrs[i].String())
			require.NoError(t, err)

			conf := &types.MsgConfirmBatch{
				Nonce:         batch.BatchNonce,
				TokenContract: batch.TokenContract.GetAddress().Hex(),
				EthSigner:     ethAddr.GetAddress().Hex(),
				Orchestrator:  orch.String(),
				Signature:     "dummysig",
			}

			input.GravityKeeper.SetBatchConfirm(ctx, conf)
		}
	}

	// try to set connfirm with invalid address
	conf := &types.MsgConfirmBatch{
		Nonce:         outogoingBatches[0].BatchNonce,
		TokenContract: outogoingBatches[0].TokenContract.GetAddress().Hex(),
		EthSigner:     EthAddrs[0].String(),
		Orchestrator:  "invalid address",
		Signature:     "dummysig",
	}
	assert.Panics(t, func() { input.GravityKeeper.SetBatchConfirm(ctx, conf) })

	// try to set connfirm with invalid token contract
	conf = &types.MsgConfirmBatch{
		Nonce:         outogoingBatches[0].BatchNonce,
		TokenContract: "invalid token",
		EthSigner:     EthAddrs[0].String(),
		Orchestrator:  OrchAddrs[0].String(),
		Signature:     "dummysig",
	}
	assert.Panics(t, func() { input.GravityKeeper.SetBatchConfirm(ctx, conf) })

	// verify that confirms are persisted for each orchestrator address
	var batchConfirm *types.MsgConfirmBatch
	for _, batch := range outogoingBatches {
		for i, addr := range OrchAddrs {
			batchConfirm = input.GravityKeeper.GetBatchConfirm(ctx, batch.BatchNonce, batch.TokenContract, addr)
			require.Equal(t, batch.BatchNonce, batchConfirm.Nonce)
			require.Equal(t, batch.TokenContract.GetAddress().Hex(), batchConfirm.TokenContract)
			require.Equal(t, EthAddrs[i].String(), batchConfirm.EthSigner)
			require.Equal(t, addr.String(), batchConfirm.Orchestrator)
			require.Equal(t, "dummysig", batchConfirm.Signature)
		}
	}
}

func TestLastSlashedBatchBlock(t *testing.T) {
	input := CreateTestEnv(t)
	ctx := input.Context

	assert.Equal(t, uint64(0), input.GravityKeeper.GetLastSlashedBatchBlock(ctx))
	assert.NotPanics(t, func() { input.GravityKeeper.SetLastSlashedBatchBlock(ctx, 2) })
	assert.Equal(t, uint64(2), input.GravityKeeper.GetLastSlashedBatchBlock(ctx))
	// LastSlashedBatchBlock cannot be set to lower than the current LastSlashedBatchBlock value
	assert.Panics(t, func() { input.GravityKeeper.SetLastSlashedBatchBlock(ctx, 1) })
	assert.Equal(t, uint64(2), input.GravityKeeper.GetLastSlashedBatchBlock(ctx))
	assert.NotPanics(t, func() { input.GravityKeeper.SetLastSlashedBatchBlock(ctx, 129) })
	assert.Equal(t, uint64(129), input.GravityKeeper.GetLastSlashedBatchBlock(ctx))
}
