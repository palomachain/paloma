package keeper

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
	vtypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint: exhaustruct
func TestBatches(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	var (
		now                     = time.Now().UTC()
		mySender, e1            = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver, e2          = types.NewEthAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr, e3 = types.NewEthAddress(testERC20Address)
		token, e4               = types.NewInternalERC20Token(math.NewInt(99999), myTokenContractAddr.GetAddress().Hex(), "test-chain")
		allVouchers             = sdk.NewCoins(sdk.NewCoin(testDenom, token.Amount))
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
	noBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", *myTokenContractAddr, 2)
	require.Nil(t, noBatch)
	require.Nil(t, err)

	// add some TX to the pool
	for i := 0; i < 4; i++ {
		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(i+100)), myTokenContractAddr.GetAddress().Hex(), "test-chain")
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

	// when
	ctx = sdkCtx.WithBlockTime(now)
	err = input.GravityKeeper.SetLastObservedEthereumBlockHeight(ctx, 1234567)
	require.NoError(t, err)
	// maxElements must be greater than 0, otherwise the batch would not be created
	noBatch, err = input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", *myTokenContractAddr, 0)
	require.Nil(t, noBatch)
	require.Error(t, err)

	// tx batch size is 2, so that some of them stay behind
	firstBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", *myTokenContractAddr, 2)
	require.NoError(t, err)

	// then batch is persisted
	gotFirstBatch, err := input.GravityKeeper.GetOutgoingTXBatch(ctx, firstBatch.TokenContract, firstBatch.BatchNonce)
	require.NoError(t, err)
	require.NotNil(t, gotFirstBatch)
	// Should have txs 4 and 3 from above, because the store is ordered by amount
	sdkCtx.Logger().Info(fmt.Sprintf("found batch %+v", gotFirstBatch))

	expFirstBatch := types.OutgoingTxBatch{
		BatchNonce: 1,
		Transactions: []types.OutgoingTransferTx{
			{
				Id:          4,
				Sender:      mySender.String(),
				DestAddress: myReceiver.GetAddress().Hex(),
				Erc20Token:  types.NewERC20Token(103, myTokenContractAddr.GetAddress().Hex(), "test-chain"),
			},
			{
				Id:          3,
				Sender:      mySender.String(),
				DestAddress: myReceiver.GetAddress().Hex(),
				Erc20Token:  types.NewERC20Token(102, myTokenContractAddr.GetAddress().Hex(), "test-chain"),
			},
		},
		TokenContract:      myTokenContractAddr.GetAddress().Hex(),
		PalomaBlockCreated: 1234567,
		BatchTimeout:       input.GravityKeeper.getBatchTimeoutHeight(ctx),
		ChainReferenceId:   "test-chain",
	}
	assert.Equal(t, expFirstBatch.BatchTimeout, gotFirstBatch.BatchTimeout)
	assert.Equal(t, expFirstBatch.BatchNonce, gotFirstBatch.BatchNonce)
	assert.Equal(t, expFirstBatch.PalomaBlockCreated, gotFirstBatch.PalomaBlockCreated)
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
			Signature:     "1234",
			Metadata: vtypes.MsgMetadata{
				Creator: orch.String(),
				Signers: []string{orch.String()},
			},
		}

		_, err = input.GravityKeeper.SetBatchConfirm(ctx, conf)
		require.NoError(t, err)
	}

	// verify that confirms are persisted
	firstBatchConfirms, err := input.GravityKeeper.GetBatchConfirmByNonceAndTokenContract(ctx, firstBatch.BatchNonce, firstBatch.TokenContract)
	require.NoError(t, err)
	require.Equal(t, len(OrchAddrs), len(firstBatchConfirms))

	gotUnbatchedTx, err := input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *myTokenContractAddr)
	require.NoError(t, err)
	oneHundredTok, err := types.NewInternalERC20Token(math.NewInt(100), myTokenContractAddr.GetAddress().Hex(), "test-chain")
	require.NoError(t, err)
	oneHundredOneTok, err := types.NewInternalERC20Token(math.NewInt(101), myTokenContractAddr.GetAddress().Hex(), "test-chain")
	require.NoError(t, err)

	// and verify remaining available Tx in the pool
	// Should still have 1: and 2: above
	expUnbatchedTx := []*types.InternalOutgoingTransferTx{
		{
			Id:          2,
			Sender:      mySender,
			DestAddress: myReceiver,
			Erc20Token:  oneHundredOneTok,
		},
		{
			Id:          1,
			Sender:      mySender,
			DestAddress: myReceiver,
			Erc20Token:  oneHundredTok,
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)

	// CREATE SECOND BATCH
	// ====================================

	ctx = sdkCtx.WithBlockTime(now)

	secondBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", *myTokenContractAddr, 2)
	require.NoError(t, err)

	err = input.GravityKeeper.SetLastObservedEthereumBlockHeight(ctx, 1234567)
	require.NoError(t, err)

	// Should have the remaining transactions
	expSecondBatch := types.OutgoingTxBatch{
		BatchNonce: 2,
		Transactions: []types.OutgoingTransferTx{
			{
				Id:          2,
				Sender:      mySender.String(),
				DestAddress: myReceiver.GetAddress().Hex(),
				Erc20Token:  types.NewERC20Token(101, myTokenContractAddr.GetAddress().Hex(), "test-chain"),
			},
			{
				Id:          1,
				Sender:      mySender.String(),
				DestAddress: myReceiver.GetAddress().Hex(),
				Erc20Token:  types.NewERC20Token(100, myTokenContractAddr.GetAddress().Hex(), "test-chain"),
			},
		},
		TokenContract:      myTokenContractAddr.GetAddress().Hex(),
		PalomaBlockCreated: 1234567,
		BatchTimeout:       input.GravityKeeper.getBatchTimeoutHeight(ctx),
		ChainReferenceId:   "test-chain",
	}

	assert.Equal(t, expSecondBatch.BatchTimeout, secondBatch.BatchTimeout)
	assert.Equal(t, expSecondBatch.BatchNonce, secondBatch.BatchNonce)
	assert.Equal(t, expSecondBatch.PalomaBlockCreated, secondBatch.PalomaBlockCreated)
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
			Signature:     "1234",
			Metadata: vtypes.MsgMetadata{
				Creator: orch.String(),
				Signers: []string{orch.String()},
			},
		}

		_, err = input.GravityKeeper.SetBatchConfirm(ctx, conf)
		require.NoError(t, err)
	}

	// verify that confirms are persisted
	secondBatchConfirms, err := input.GravityKeeper.GetBatchConfirmByNonceAndTokenContract(ctx, secondBatch.BatchNonce, secondBatch.TokenContract)
	require.NoError(t, err)
	require.Equal(t, len(OrchAddrs), len(secondBatchConfirms))

	// check that last added batch is the one with the highest nonce
	lastOutgoingBatch, err := input.GravityKeeper.GetLastOutgoingBatchByTokenType(ctx, *myTokenContractAddr)
	require.NoError(t, err)
	require.NotNil(t, lastOutgoingBatch)
	assert.Equal(t, lastOutgoingBatch.BatchNonce, secondBatch.BatchNonce)

	// EXECUTE THE SECOND BATCH
	// =================================

	// Execute the batch
	fakeBlock := secondBatch.PalomaBlockCreated // A fake ethereum block used for testing only
	msg := types.MsgBatchSendToEthClaim{
		EthBlockHeight:   fakeBlock,
		BatchNonce:       secondBatch.BatchNonce,
		TokenContract:    secondBatch.TokenContract.GetAddress().String(),
		ChainReferenceId: secondBatch.GetChainReferenceID(),
		Metadata: vtypes.MsgMetadata{
			Creator: OrchAddrs[0].String(),
			Signers: []string{OrchAddrs[0].String()},
		},
	}
	err = input.GravityKeeper.OutgoingTxBatchExecuted(ctx, secondBatch.TokenContract, msg)
	require.NoError(t, err)

	// check batch has been deleted
	gotSecondBatch, err := input.GravityKeeper.GetOutgoingTXBatch(ctx, secondBatch.TokenContract, secondBatch.BatchNonce)
	require.NoError(t, err)
	require.Nil(t, gotSecondBatch)
	// check batch confirmations have been deleted
	secondBatchConfirms, err = input.GravityKeeper.GetBatchConfirmByNonceAndTokenContract(ctx, secondBatch.BatchNonce, secondBatch.TokenContract)
	require.NoError(t, err)
	require.Equal(t, 0, len(secondBatchConfirms))
}

// tests that batches work with large token amounts, mostly a duplicate of the above
// tests but using much bigger numbers
// nolint: exhaustruct
func TestBatchesFullCoins(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	var (
		now              = time.Now().UTC()
		mySender, e1     = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver       = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		receiverAddr, e2 = types.NewEthAddress(myReceiver)
		totalCoins, _    = math.NewIntFromString("1500000000000000000000") // 1,500 ETH worth
		oneEth, _        = math.NewIntFromString("1000000000000000000")
		token, e3        = types.NewInternalERC20Token(totalCoins, testERC20Address, "test-chain")
		allVouchers      = sdk.NewCoins(sdk.NewCoin(testDenom, token.Amount))
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	tokenContract, err := types.NewEthAddress(testERC20Address)
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
		vAsSDKInt := math.NewIntFromUint64(v)
		amountToken, err := types.NewInternalERC20Token(oneEth.Mul(vAsSDKInt), testERC20Address, "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)

		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiverAddr, amount, "test-chain")
		require.NoError(t, err)
	}

	// when
	ctx = sdkCtx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	firstBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", *tokenContract, 2)
	require.NoError(t, err)

	// then batch is persisted
	gotFirstBatch, err := input.GravityKeeper.GetOutgoingTXBatch(ctx, firstBatch.TokenContract, firstBatch.BatchNonce)
	require.NoError(t, err)
	require.NotNil(t, gotFirstBatch)

	expFirstBatch := &types.OutgoingTxBatch{
		BatchNonce: 1,
		Transactions: []types.OutgoingTransferTx{
			{
				Id:          2,
				Sender:      mySender.String(),
				DestAddress: myReceiver,
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(math.NewIntFromUint64(300)), testERC20Address, "test-chain"),
			},
			{
				Id:          3,
				Sender:      mySender.String(),
				DestAddress: myReceiver,
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(math.NewIntFromUint64(25)), testERC20Address, "test-chain"),
			},
		},
		TokenContract:      testERC20Address,
		PalomaBlockCreated: 1234567,
		ChainReferenceId:   "test-chain",
	}
	assert.Equal(t, expFirstBatch.BatchNonce, gotFirstBatch.BatchNonce)
	assert.Equal(t, expFirstBatch.PalomaBlockCreated, gotFirstBatch.PalomaBlockCreated)
	assert.Equal(t, expFirstBatch.TokenContract, gotFirstBatch.TokenContract.GetAddress().Hex())
	assert.Equal(t, len(expFirstBatch.Transactions), len(gotFirstBatch.Transactions))
	for i := 0; i < len(expFirstBatch.Transactions); i++ {
		assert.Equal(t, expFirstBatch.Transactions[i], gotFirstBatch.Transactions[i].ToExternal())
	}

	// and verify remaining available Tx in the pool
	gotUnbatchedTx, err := input.GravityKeeper.GetUnbatchedTransactionsByContract(ctx, *tokenContract)
	require.NoError(t, err)
	twentyTok, err := types.NewInternalERC20Token(oneEth.Mul(math.NewIntFromUint64(20)), testERC20Address, "test-chain")
	require.NoError(t, err)
	tenTok, err := types.NewInternalERC20Token(oneEth.Mul(math.NewIntFromUint64(10)), testERC20Address, "test-chain")
	require.NoError(t, err)
	expUnbatchedTx := []*types.InternalOutgoingTransferTx{
		{
			Id:          1,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  twentyTok,
		},
		{
			Id:          4,
			Sender:      mySender,
			DestAddress: receiverAddr,
			Erc20Token:  tenTok,
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)

	// Create second batch
	// ====================================

	// add some more TX to the pool to create a more profitable batch
	for _, v := range []uint64{200, 150} {
		vAsSDKInt := math.NewIntFromUint64(v)
		amountToken, err := types.NewInternalERC20Token(oneEth.Mul(vAsSDKInt), testERC20Address, "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)

		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiverAddr, amount, "test-chain")
		require.NoError(t, err)
	}

	ctx = sdkCtx.WithBlockTime(now)
	err = input.GravityKeeper.SetLastObservedEthereumBlockHeight(ctx, 1234567)
	require.NoError(t, err)
	// tx batch size is 2, so that some of them stay behind
	secondBatch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", *tokenContract, 2)
	require.NoError(t, err)

	// check that the batch has the right txs in it
	expSecondBatch := &types.OutgoingTxBatch{
		BatchNonce: 2,
		Transactions: []types.OutgoingTransferTx{
			{
				Id:          5,
				Sender:      mySender.String(),
				DestAddress: myReceiver,
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(math.NewIntFromUint64(200)), testERC20Address, "test-chain"),
			},
			{
				Id:          6,
				Sender:      mySender.String(),
				DestAddress: myReceiver,
				Erc20Token:  types.NewSDKIntERC20Token(oneEth.Mul(math.NewIntFromUint64(150)), testERC20Address, "test-chain"),
			},
		},
		TokenContract:      testERC20Address,
		PalomaBlockCreated: 1234567,
		BatchTimeout:       input.GravityKeeper.getBatchTimeoutHeight(ctx),
		ChainReferenceId:   "test-chain",
	}

	assert.Equal(t, expSecondBatch.BatchTimeout, secondBatch.BatchTimeout)
	assert.Equal(t, expSecondBatch.BatchNonce, secondBatch.BatchNonce)
	assert.Equal(t, expSecondBatch.PalomaBlockCreated, secondBatch.PalomaBlockCreated)
	assert.Equal(t, expSecondBatch.TokenContract, secondBatch.TokenContract.GetAddress().Hex())
	assert.Equal(t, len(expSecondBatch.Transactions), len(secondBatch.Transactions))
	for i := 0; i < len(expSecondBatch.Transactions); i++ {
		assert.Equal(t, expSecondBatch.Transactions[i], secondBatch.Transactions[i].ToExternal())
	}

	// Execute the batch
	fakeBlock := secondBatch.PalomaBlockCreated // A fake ethereum block used for testing only
	msg := types.MsgBatchSendToEthClaim{
		EthBlockHeight:   fakeBlock,
		BatchNonce:       secondBatch.BatchNonce,
		TokenContract:    testERC20Address,
		ChainReferenceId: secondBatch.GetChainReferenceID(),
	}
	err = input.GravityKeeper.OutgoingTxBatchExecuted(ctx, secondBatch.TokenContract, msg)
	require.NoError(t, err)

	// check batch has been deleted
	gotSecondBatch, err := input.GravityKeeper.GetOutgoingTXBatch(ctx, secondBatch.TokenContract, secondBatch.BatchNonce)
	require.NoError(t, err)
	require.Nil(t, gotSecondBatch)
}

// TestManyBatches handles test cases around batch execution, specifically executing multiple batches
// out of sequential order, which is exactly what happens on the
// nolint: exhaustruct
//func TestManyBatches(t *testing.T) {
//	input := CreateTestEnv(t)
//	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()
//
//	ctx := input.Context
//	var (
//		now                = time.Now().UTC()
//		mySender, e1       = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
//		myReceiver         = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
//		tokenContractAddr1 = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5"
//		tokenContractAddr2 = "0xF815240800ddf3E0be80e0d848B13ecaa504BF37"
//		tokenContractAddr3 = "0xd086dDA7BccEB70e35064f540d07E4baED142cB3"
//		tokenContractAddr4 = "0x384981B9d133701c4bD445F77bF61C3d80e79D46"
//		totalCoins, _      = math.NewIntFromString("1500000000000000000000000")
//		oneEth, _          = math.NewIntFromString("1000000000000000000")
//		token1, e2         = types.NewInternalERC20Token(totalCoins, tokenContractAddr1, "test-chain")
//		token2, e3         = types.NewInternalERC20Token(totalCoins, tokenContractAddr2, "test-chain")
//		token3, e4         = types.NewInternalERC20Token(totalCoins, tokenContractAddr3, "test-chain")
//		token4, e5         = types.NewInternalERC20Token(totalCoins, tokenContractAddr4, "test-chain")
//		allVouchers        = sdk.NewCoins(sdk.NewCoin(testDenom, token1.Amount.Add(token2.Amount).Add(token3.Amount).Add(token4.Amount)))
//	)
//	require.NoError(t, e1)
//	require.NoError(t, e2)
//	require.NoError(t, e3)
//	require.NoError(t, e4)
//	require.NoError(t, e5)
//	receiver, err := types.NewEthAddress(myReceiver)
//	require.NoError(t, err)
//
//	// mint vouchers first
//	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
//	// set senders balance
//	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
//	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))
//	input.GravityKeeper.SetLastObservedEthereumBlockHeight(ctx, 1234567)
//
//	// CREATE FIRST BATCH
//	// ==================
//
//	tokens := [4]string{tokenContractAddr1, tokenContractAddr2, tokenContractAddr3, tokenContractAddr4}
//
//	// when
//	ctx = ctx.WithBlockTime(now)
//	var batches []types.OutgoingTxBatch
//
//	for _, contract := range tokens {
//		contractAddr, err := types.NewEthAddress(contract)
//		require.NoError(t, err)
//		for v := 1; v < 500; v++ {
//			vAsSDKInt := math.NewIntFromUint64(uint64(v))
//			amountToken, err := types.NewInternalERC20Token(oneEth.Mul(vAsSDKInt), contract, "test-chain")
//			require.NoError(t, err)
//			amount := sdk.NewCoin(testDenom, amountToken.Amount)
//
//			_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, "test-chain")
//			require.NoError(t, err)
//			// create batch after every 100 txs to be able to create more profitable batches
//			if (v+1)%100 == 0 {
//				batch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", *contractAddr, 100)
//				require.NoError(t, err)
//				batches = append(batches, batch.ToExternal())
//			}
//		}
//	}
//
//	for _, batch := range batches {
//		// then batch is persisted
//		contractAddr, err := types.NewEthAddress(batch.TokenContract)
//		require.NoError(t, err)
//		gotBatch := input.GravityKeeper.GetOutgoingTXBatch(ctx, *contractAddr, batch.BatchNonce)
//		require.NotNil(t, gotBatch)
//	}
//
//	// EXECUTE BOTH BATCHES
//	// =================================
//
//	// shuffle batches to simulate out of order execution on Ethereum
//	rand.Seed(time.Now().UnixNano())
//	rand.Shuffle(len(batches), func(i, j int) { batches[i], batches[j] = batches[j], batches[i] })
//
//	// Execute the batches, if there are any problems OutgoingTxBatchExecuted will panic
//	for _, batch := range batches {
//		contractAddr, err := types.NewEthAddress(batch.TokenContract)
//		require.NoError(t, err)
//		gotBatch := input.GravityKeeper.GetOutgoingTXBatch(ctx, *contractAddr, batch.BatchNonce)
//		// we may have already deleted some of the batches in this list by executing later ones
//		if gotBatch != nil {
//			fakeBlock := batch.PalomaBlockCreated // A fake ethereum block used for testing only
//			msg := types.MsgBatchSendToEthClaim{EthBlockHeight: fakeBlock, BatchNonce: batch.BatchNonce}
//			input.GravityKeeper.OutgoingTxBatchExecuted(ctx, *contractAddr, msg)
//		}
//	}
//}
//
//// nolint: exhaustruct
//func TestPoolTxRefund(t *testing.T) {
//	input := CreateTestEnv(t)
//	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()
//
//	ctx := input.Context
//	var (
//		now             = time.Now().UTC()
//		mySender, e1    = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
//		notMySender, e2 = sdk.AccAddressFromBech32("gravity1add7f8wyertuus9r20284ej0asrs085c8ajr0y")
//		myReceiver      = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
//		token, e3       = types.NewInternalERC20Token(math.NewInt(414), testERC20Address, "test-chain")
//		allVouchers     = sdk.NewCoins(sdk.NewCoin(testDenom, token.Amount))
//		denomToken, e4  = types.NewInternalERC20Token(math.NewInt(1), testERC20Address, "test-chain")
//		myDenom         = sdk.NewCoin(testDenom, denomToken.Amount).Denom
//	)
//	require.NoError(t, e1)
//	require.NoError(t, e2)
//	require.NoError(t, e3)
//	require.NoError(t, e4)
//	contract, err := types.NewEthAddress(testERC20Address)
//	require.NoError(t, err)
//	receiver, err := types.NewEthAddress(myReceiver)
//	require.NoError(t, err)
//
//	// mint some voucher first
//	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
//	// set senders balance
//	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
//	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))
//
//	// CREATE FIRST BATCH
//	// ==================
//
//	// add some TX to the pool
//	for i := 0; i < 4; i++ {
//		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(i+100)), testERC20Address, "test-chain")
//		require.NoError(t, err)
//		amount := sdk.NewCoin(testDenom, amountToken.Amount)
//
//		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, "test-chain")
//		require.NoError(t, err)
//		// Should have created:
//		// 1: amount 100
//		// 2: amount 101
//		// 3: amount 102
//		// 4: amount 103
//	}
//
//	// when
//	ctx = ctx.WithBlockTime(now)
//
//	// Check the balance at the start
//	balances := input.BankKeeper.GetAllBalances(ctx, mySender)
//	require.Equal(t, math.NewInt(8), balances.AmountOf(myDenom))
//
//	// tx batch size is 2, so that some of them stay behind
//	// Should have 4: and 3: from above
//	_, err = input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", *contract, 2)
//	require.NoError(t, err)
//
//	// try to refund a tx that's in a batch
//	err1 := input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 3, mySender)
//	require.Error(t, err1)
//
//	// try to refund somebody else's tx
//	err2 := input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 2, notMySender)
//	require.Error(t, err2)
//
//	// try to refund a tx that's in the pool
//	err3 := input.GravityKeeper.RemoveFromOutgoingPoolAndRefund(ctx, 2, mySender)
//	require.NoError(t, err3)
//
//	// make sure refund was issued
//	balances = input.BankKeeper.GetAllBalances(ctx, mySender)
//	require.Equal(t, math.NewInt(109), balances.AmountOf(myDenom))
//}

func TestBatchConfirms(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	var (
		now                     = time.Now().UTC()
		mySender, e1            = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver, e2          = types.NewEthAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr, e3 = types.NewEthAddress(testERC20Address)
		token, e4               = types.NewInternalERC20Token(math.NewInt(1000000), myTokenContractAddr.GetAddress().Hex(), "test-chain")
		allVouchers             = sdk.NewCoins(sdk.NewCoin(testDenom, token.Amount))
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
	ctx = sdk.UnwrapSDKContext(ctx).WithBlockTime(now)

	// add batches with 1 tx to the pool
	for i := 1; i < 200; i++ {
		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(i+100)), myTokenContractAddr.GetAddress().Hex(), "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)

		// add tx to the pool
		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *myReceiver, amount, "test-chain")
		require.NoError(t, err)

		// create batch
		_, err = input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", *myTokenContractAddr, 1)
		require.NoError(t, err)
	}

	outogoingBatches, err := input.GravityKeeper.GetOutgoingTxBatches(ctx)
	require.NoError(t, err)

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

			_, err = input.GravityKeeper.SetBatchConfirm(ctx, conf)
			require.NoError(t, err)
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
	_, err = input.GravityKeeper.SetBatchConfirm(ctx, conf)
	assert.Error(t, err)

	// try to set confirm with invalid token contract
	conf = &types.MsgConfirmBatch{
		Nonce:         outogoingBatches[0].BatchNonce,
		TokenContract: "invalid token",
		EthSigner:     EthAddrs[0].String(),
		Orchestrator:  OrchAddrs[0].String(),
		Signature:     "dummysig",
	}
	_, err = input.GravityKeeper.SetBatchConfirm(ctx, conf)
	assert.Error(t, err)

	// verify that confirms are persisted for each orchestrator address
	var batchConfirm *types.MsgConfirmBatch
	for _, batch := range outogoingBatches {
		for i, addr := range OrchAddrs {
			batchConfirm, err = input.GravityKeeper.GetBatchConfirm(ctx, batch.BatchNonce, batch.TokenContract, addr)
			require.NoError(t, err)
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

	lastSlashedBlock, err := input.GravityKeeper.GetLastSlashedBatchBlock(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), lastSlashedBlock)

	err = input.GravityKeeper.SetLastSlashedBatchBlock(ctx, 2)
	assert.NoError(t, err)

	lastSlashedBlock, err = input.GravityKeeper.GetLastSlashedBatchBlock(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), lastSlashedBlock)

	// LastSlashedBatchBlock cannot be set to lower than the current LastSlashedBatchBlock value
	err = input.GravityKeeper.SetLastSlashedBatchBlock(ctx, 1)
	assert.Error(t, err)

	lastSlashedBlock, err = input.GravityKeeper.GetLastSlashedBatchBlock(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), lastSlashedBlock)

	err = input.GravityKeeper.SetLastSlashedBatchBlock(ctx, 129)
	assert.NoError(t, err)

	lastSlashedBlock, err = input.GravityKeeper.GetLastSlashedBatchBlock(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(129), lastSlashedBlock)
}
