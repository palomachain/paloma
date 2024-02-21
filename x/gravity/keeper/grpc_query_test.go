package keeper

import (
	"bytes"
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
func TestLastPendingBatchRequest(t *testing.T) {
	specs := map[string]struct {
		expResp types.QueryLastPendingBatchRequestByAddrResponse
	}{
		"find batch": {
			expResp: types.QueryLastPendingBatchRequestByAddrResponse{
				Batch: []types.OutgoingTxBatch{
					{
						BatchNonce: 1,
						Transactions: []types.OutgoingTransferTx{
							{
								Id:          4,
								Sender:      "paloma1qyqszqgpqyqszqgpqyqszqgpqyqszqgp2kvale",
								DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
								Erc20Token: types.ERC20Token{
									Amount:           math.NewInt(103),
									Contract:         testERC20Address,
									ChainReferenceId: "test-chain",
								},
							},
							{
								Id:          3,
								Sender:      "paloma1qyqszqgpqyqszqgpqyqszqgpqyqszqgp2kvale",
								DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
								Erc20Token: types.ERC20Token{
									Amount:           math.NewInt(102),
									Contract:         testERC20Address,
									ChainReferenceId: "test-chain",
								},
							},
						},
						TokenContract:      testERC20Address,
						PalomaBlockCreated: 1235067,
						ChainReferenceId:   "test-chain",
					},
				},
			},
		},
	}
	// any lower than this and a validator won't be created
	const minStake = 1000000
	input, _ := SetupTestChain(t, []uint64{minStake, minStake, minStake, minStake, minStake})
	sdkCtx := sdk.UnwrapSDKContext(input.Context)

	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := sdk.WrapSDKContext(sdkCtx)
	var valAddr sdk.AccAddress = bytes.Repeat([]byte{byte(1)}, 20)
	createTestBatch(t, input, 2)
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			req := new(types.QueryLastPendingBatchRequestByAddrRequest)
			req.Address = valAddr.String()
			got, err := input.GravityKeeper.LastPendingBatchRequestByAddr(ctx, req)
			require.NoError(t, err)

			// Don't bother comparing some computed values
			got.Batch[0].BatchTimeout = 0
			got.Batch[0].BytesToSign = nil
			got.Batch[0].Assignee = ""

			assert.Equal(t, &spec.expResp, got, got)
		})
	}
}

// nolint: exhaustruct
func createTestBatch(t *testing.T, input TestInput, maxTxElements uint) {
	sdkCtx := sdk.UnwrapSDKContext(input.Context)
	var (
		mySender   = bytes.Repeat([]byte{1}, 20)
		myReceiver = "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934"
		now        = time.Now().UTC()
	)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract, err := types.NewEthAddress(testERC20Address)
	require.NoError(t, err)
	// mint some voucher first
	token, err := types.NewInternalERC20Token(math.NewInt(99999), testERC20Address, "test-chain")
	require.NoError(t, err)
	allVouchers := sdk.Coins{sdk.NewCoin(testDenom, token.Amount)}
	err = input.BankKeeper.MintCoins(input.Context, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(input.Context, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(input.Context, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	// add some TX to the pool
	for i := 0; i < 4; i++ {
		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(i+100)), testERC20Address, "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)
		_, err = input.GravityKeeper.AddToOutgoingPool(input.Context, mySender, *receiver, amount, "test-chain")
		require.NoError(t, err)
		// Should create:
		// 1: amount 100
		// 2: amount 101
		// 3: amount 102
		// 4: amount 103
	}
	// when
	input.Context = sdkCtx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	_, err = input.GravityKeeper.BuildOutgoingTXBatch(input.Context, "test-chain", *tokenContract, maxTxElements)
	require.NoError(t, err)
	// Should have 2 and 3 from above
	// 1 and 4 should be unbatched
}

// nolint: exhaustruct
func TestQueryAllBatchConfirms(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	k := input.GravityKeeper

	var (
		tokenContract      = "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"
		validatorAddr, err = sdk.AccAddressFromBech32("paloma1mgamdcs9dah0vn0gqupl05up7pedg2mvyy5e9j")
	)
	require.NoError(t, err)

	_, err = input.GravityKeeper.SetBatchConfirm(sdkCtx, &types.MsgConfirmBatch{
		Nonce:         1,
		TokenContract: tokenContract,
		EthSigner:     "0xf35e2cc8e6523d683ed44870f5b7cc785051a77d",
		Orchestrator:  validatorAddr.String(),
		Signature:     "d34db33f",
		Metadata: vtypes.MsgMetadata{
			Creator: validatorAddr.String(),
			Signers: []string{validatorAddr.String()},
		},
	})
	require.NoError(t, err)

	batchConfirms, err := k.BatchConfirms(ctx, &types.QueryBatchConfirmsRequest{Nonce: 1, ContractAddress: tokenContract})
	require.NoError(t, err)

	expectedRes := types.QueryBatchConfirmsResponse{
		Confirms: []types.MsgConfirmBatch{
			{
				Nonce:         1,
				TokenContract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
				EthSigner:     "0xf35e2cc8e6523d683ed44870f5b7cc785051a77d",
				Orchestrator:  "paloma1mgamdcs9dah0vn0gqupl05up7pedg2mvyy5e9j",
				Signature:     "d34db33f",
				Metadata: vtypes.MsgMetadata{
					Creator: "paloma1mgamdcs9dah0vn0gqupl05up7pedg2mvyy5e9j",
					Signers: []string{"paloma1mgamdcs9dah0vn0gqupl05up7pedg2mvyy5e9j"},
				},
			},
		},
	}

	assert.Equal(t, &expectedRes, batchConfirms, "json is equal")
}

// nolint: exhaustruct
// Check with multiple nonces and tokenContracts
func TestQueryBatch(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	k := input.GravityKeeper

	createTestBatch(t, input, 2)

	batch, err := k.BatchRequestByNonce(ctx, &types.QueryBatchRequestByNonceRequest{Nonce: 1, ContractAddress: testERC20Address})
	require.NoError(t, err)

	expectedRes := types.QueryBatchRequestByNonceResponse{
		Batch: types.OutgoingTxBatch{
			BatchTimeout: 0,
			Transactions: []types.OutgoingTransferTx{
				{
					DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
					Erc20Token: types.ERC20Token{
						Amount:           math.NewInt(103),
						Contract:         testERC20Address,
						ChainReferenceId: "test-chain",
					},
					Sender: "paloma1qyqszqgpqyqszqgpqyqszqgpqyqszqgp2kvale",
					Id:     4,
				},
				{
					DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
					Erc20Token: types.ERC20Token{
						Amount:           math.NewInt(102),
						Contract:         testERC20Address,
						ChainReferenceId: "test-chain",
					},
					Sender: "paloma1qyqszqgpqyqszqgpqyqszqgpqyqszqgp2kvale",
					Id:     3,
				},
			},
			BatchNonce:         1,
			PalomaBlockCreated: 1234567,
			TokenContract:      testERC20Address,
			ChainReferenceId:   "test-chain",
		},
	}

	// Don't bother comparing some computed values
	batch.Batch.BatchTimeout = 0
	batch.Batch.BytesToSign = nil
	batch.Batch.Assignee = ""

	assert.Equal(t, &expectedRes, batch, batch)
}

// nolint: exhaustruct
func TestLastBatchesRequest(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	k := input.GravityKeeper

	createTestBatch(t, input, 2)
	createTestBatch(t, input, 3)

	lastBatches, err := k.OutgoingTxBatches(ctx, &types.QueryOutgoingTxBatchesRequest{})
	require.NoError(t, err)

	expectedRes := types.QueryOutgoingTxBatchesResponse{
		Batches: []types.OutgoingTxBatch{
			{
				Transactions: []types.OutgoingTransferTx{
					{
						DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
						Erc20Token: types.ERC20Token{
							Amount:           math.NewInt(103),
							Contract:         testERC20Address,
							ChainReferenceId: "test-chain",
						},
						Sender: "paloma1qyqszqgpqyqszqgpqyqszqgpqyqszqgp2kvale",
						Id:     8,
					},
					{
						DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
						Erc20Token: types.ERC20Token{
							Amount:           math.NewInt(102),
							Contract:         testERC20Address,
							ChainReferenceId: "test-chain",
						},
						Sender: "paloma1qyqszqgpqyqszqgpqyqszqgpqyqszqgp2kvale",
						Id:     7,
					},
					{
						DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
						Erc20Token: types.ERC20Token{
							Amount:           math.NewInt(101),
							Contract:         testERC20Address,
							ChainReferenceId: "test-chain",
						},
						Sender: "paloma1qyqszqgpqyqszqgpqyqszqgpqyqszqgp2kvale",
						Id:     6,
					},
				},
				BatchNonce:         2,
				PalomaBlockCreated: 1234567,
				TokenContract:      testERC20Address,
				ChainReferenceId:   "test-chain",
			},
			{
				Transactions: []types.OutgoingTransferTx{
					{
						DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
						Erc20Token: types.ERC20Token{
							Amount:           math.NewInt(103),
							Contract:         testERC20Address,
							ChainReferenceId: "test-chain",
						},
						Sender: "paloma1qyqszqgpqyqszqgpqyqszqgpqyqszqgp2kvale",
						Id:     4,
					},
					{
						DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
						Erc20Token: types.ERC20Token{
							Amount:           math.NewInt(102),
							Contract:         testERC20Address,
							ChainReferenceId: "test-chain",
						},
						Sender: "paloma1qyqszqgpqyqszqgpqyqszqgpqyqszqgp2kvale",
						Id:     3,
					},
				},
				BatchNonce:         1,
				PalomaBlockCreated: 1234567,
				TokenContract:      testERC20Address,
				ChainReferenceId:   "test-chain",
			},
		},
	}

	// Don't bother comparing some computed values
	lastBatches.Batches[0].BatchTimeout = 0
	lastBatches.Batches[0].BytesToSign = nil
	lastBatches.Batches[0].Assignee = ""
	lastBatches.Batches[1].BatchTimeout = 0
	lastBatches.Batches[1].BytesToSign = nil
	lastBatches.Batches[1].Assignee = ""

	assert.Equal(t, &expectedRes, lastBatches, "json is equal")
}

// nolint: exhaustruct
func TestQueryERC20ToDenom(t *testing.T) {
	var (
		chainReferenceID = "test-chain"
		erc20, err       = types.NewEthAddress(testERC20Address)
	)
	require.NoError(t, err)
	response := types.QueryERC20ToDenomResponse{
		Denom: testDenom,
	}
	input := CreateTestEnv(t)
	sdkCtx := sdk.UnwrapSDKContext(input.Context)
	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := sdk.UnwrapSDKContext(sdkCtx)
	k := input.GravityKeeper
	err = input.GravityKeeper.setDenomToERC20(sdkCtx, chainReferenceID, testDenom, *erc20)
	require.NoError(t, err)

	queriedDenom, err := k.ERC20ToDenom(ctx, &types.QueryERC20ToDenomRequest{
		Erc20:            erc20.GetAddress().Hex(),
		ChainReferenceId: chainReferenceID,
	})
	require.NoError(t, err)

	assert.Equal(t, &response, queriedDenom)
}

// nolint: exhaustruct
func TestQueryDenomToERC20(t *testing.T) {
	var (
		chainReferenceID = "test-chain"
		erc20, err       = types.NewEthAddress(testERC20Address)
	)
	require.NoError(t, err)
	response := types.QueryDenomToERC20Response{
		Erc20: erc20.GetAddress().Hex(),
	}
	input := CreateTestEnv(t)
	defer func() {
		sdk.UnwrapSDKContext(input.Context).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	sdkCtx := input.Context

	k := input.GravityKeeper

	err = input.GravityKeeper.setDenomToERC20(sdkCtx, chainReferenceID, testDenom, *erc20)
	require.NoError(t, err)

	queriedERC20, err := k.DenomToERC20(input.Context, &types.QueryDenomToERC20Request{
		Denom:            testDenom,
		ChainReferenceId: chainReferenceID,
	})
	require.NoError(t, err)

	assert.Equal(t, &response, queriedERC20)
}

// nolint: exhaustruct
func TestQueryPendingSendToEth(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	sdkCtx := sdk.UnwrapSDKContext(input.Context)
	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()
	k := input.GravityKeeper
	var (
		now            = time.Now().UTC()
		mySender, err1 = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver     = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		token, err2    = types.NewInternalERC20Token(math.NewInt(99999), testERC20Address, "test-chain")
		allVouchers    = sdk.NewCoins(sdk.NewCoin(testDenom, token.Amount))
	)
	require.NoError(t, err1)
	require.NoError(t, err2)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract, err := types.NewEthAddress(testERC20Address)
	require.NoError(t, err)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(sdkCtx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(sdkCtx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, mySender, allVouchers))

	// CREATE FIRST BATCH
	// ==================

	// add some TX to the pool
	for i := 0; i < 4; i++ {
		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(i+100)), testERC20Address, "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)
		_, err = input.GravityKeeper.AddToOutgoingPool(sdkCtx, mySender, *receiver, amount, "test-chain")
		require.NoError(t, err)
		// Should create:
		// 1: amount 100
		// 2: amount 101
		// 3: amount 102
		// 4: amount 104
	}

	// when
	sdkCtx = sdkCtx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	// Should contain 2 and 3 from above
	_, err = input.GravityKeeper.BuildOutgoingTXBatch(sdkCtx, "test-chain", *tokenContract, 2)
	require.NoError(t, err)

	// Should receive 1 and 4 unbatched, 2 and 3 batched in response
	response, err := k.GetPendingSendToEth(ctx, &types.QueryPendingSendToEth{SenderAddress: mySender.String()})
	require.NoError(t, err)
	expectedRes := types.QueryPendingSendToEthResponse{
		TransfersInBatches: []types.OutgoingTransferTx{
			{
				Id:          4,
				Sender:      "paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk",
				DestAddress: "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
				Erc20Token: types.ERC20Token{
					Contract:         testERC20Address,
					Amount:           math.NewInt(103),
					ChainReferenceId: "test-chain",
				},
			},
			{
				Id:          3,
				Sender:      "paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk",
				DestAddress: "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
				Erc20Token: types.ERC20Token{
					Contract:         testERC20Address,
					Amount:           math.NewInt(102),
					ChainReferenceId: "test-chain",
				},
			},
		},

		UnbatchedTransfers: []types.OutgoingTransferTx{
			{
				Id:          2,
				Sender:      "paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk",
				DestAddress: "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
				Erc20Token: types.ERC20Token{
					Contract:         testERC20Address,
					Amount:           math.NewInt(101),
					ChainReferenceId: "test-chain",
				},
			},
			{
				Id:          1,
				Sender:      "paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk",
				DestAddress: "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
				Erc20Token: types.ERC20Token{
					Contract:         testERC20Address,
					Amount:           math.NewInt(100),
					ChainReferenceId: "test-chain",
				},
			},
		},
	}

	assert.Equal(t, &expectedRes, response, "json is equal")
}
