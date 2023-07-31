package keeper

import (
	"fmt"
	"sort"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/palomachain/paloma/x/bech32ibc"
	bech32ibctypes "github.com/palomachain/paloma/x/bech32ibc/types"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/require"
)

// Tests that batches and transactions are preserved during chain restart, including pending forwards
func TestBatchAndTxImportExport(t *testing.T) {
	// SETUP ENV + DATA
	// ==================
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	batchSize := 100
	accAddresses := []string{ // Warning: this must match the length of ctrAddresses

		"gravity1dg55rtevlfxh46w88yjpdd08sqhh5cc3z8yqu6",
		"gravity164knshrzuuurf05qxf3q5ewpfnwzl4gj3t84vv",
		"gravity193fw83ynn76328pty4yl7473vg9x86aly623wk",
		"gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm",
		"gravity1ees2tqhhhm9ahlhceh2zdguww9lqn2ckcxpllh",
	}
	ethAddresses := []string{
		"0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
		"0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD8",
		"0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD9",
		"0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD0",
		"0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD1",
		"0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD2",
		"0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD3",
	}
	ctrAddresses := []string{ // Warning: this must match the length of accAddresses
		"0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5",
		"0x429881672b9AE42b8eBA0e26cd9c73711b891ca6",
		"0x429881672b9aE42b8eba0e26cD9c73711B891Ca7",
		"0x429881672B9AE42b8EbA0E26cD9C73711b891Ca8",
		"0x429881672B9AE42b8EbA0E26cD9C73711b891Ca9",
	}

	// SETUP ACCOUNTS
	// ==================
	senders := make([]*sdk.AccAddress, len(accAddresses))
	for i := range senders {
		sender, err := sdk.AccAddressFromBech32(accAddresses[i])
		require.NoError(t, err)
		senders[i] = &sender
	}
	receivers := make([]*types.EthAddress, len(ethAddresses))
	for i := range receivers {
		receiver, err := types.NewEthAddress(ethAddresses[i])
		require.NoError(t, err)
		receivers[i] = receiver
	}
	contracts := make([]*types.EthAddress, len(ctrAddresses))
	for i := range contracts {
		contract, err := types.NewEthAddress(ctrAddresses[i])
		require.NoError(t, err)
		contracts[i] = contract
	}
	tokens := make([]*types.InternalERC20Token, len(contracts))
	vouchers := make([]*sdk.Coins, len(contracts))
	for i, v := range contracts {
		token, err := types.NewInternalERC20Token(sdk.NewInt(99999999), v.GetAddress().Hex())
		tokens[i] = token
		allVouchers := sdk.NewCoins(token.GravityCoin())
		vouchers[i] = &allVouchers
		require.NoError(t, err)

		// Mint the vouchers
		require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	}

	// give sender i a balance of token i
	for i, v := range senders {
		input.AccountKeeper.NewAccountWithAddress(ctx, *v)
		require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, *v, *vouchers[i]))
	}

	// CREATE TRANSACTIONS
	// ==================
	numTxs := 5000 // should end up with 1000 txs per contract
	txs := make([]*types.InternalOutgoingTransferTx, numTxs)
	fees := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	amounts := []int{51, 52, 53, 54, 55, 56, 57, 58, 59, 60}
	for i := 0; i < numTxs; i++ {
		// Pick fee, amount, sender, receiver, and contract for the ith transaction
		// Sender and contract will always match up (they must since sender i controls the whole balance of the ith token)
		// Receivers should get a balance of many token types since i % len(receivers) is usually different than i % len(contracts)
		fee := fees[i%len(fees)] // fee for this transaction
		amount := amounts[i%len(amounts)]
		sender := senders[i%len(senders)]
		receiver := receivers[i%len(receivers)]
		contract := contracts[i%len(contracts)]
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(amount)), contract.GetAddress().Hex())
		require.NoError(t, err)
		feeToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(fee)), contract.GetAddress().Hex())
		require.NoError(t, err)

		// add transaction to the pool
		id, err := input.GravityKeeper.AddToOutgoingPool(ctx, *sender, *receiver, amountToken.GravityCoin(), feeToken.GravityCoin())
		require.NoError(t, err)
		ctx.Logger().Info(fmt.Sprintf("Created transaction %v with amount %v and fee %v of contract %v from %v to %v", i, amount, fee, contract, sender, receiver))

		// Record the transaction for later testing
		tx, err := types.NewInternalOutgoingTransferTx(id, sender.String(), receiver.GetAddress().Hex(), amountToken.ToExternal(), feeToken.ToExternal())
		require.NoError(t, err)
		txs[i] = tx
	}

	// when

	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now)

	// CREATE BATCHES
	// ==================
	// Want to create batches for half of the transactions for each contract
	// with 100 tx in each batch, 1000 txs per contract, we want 5 batches per contract to batch 500 txs per contract
	batches := make([]*types.InternalOutgoingTxBatch, 5*len(contracts))
	for i, v := range contracts {
		batch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, *v, uint(batchSize))
		require.NoError(t, err)
		batches[i] = batch
		ctx.Logger().Info(fmt.Sprintf("Created batch %v for contract %v with %v transactions", i, v.GetAddress(), batchSize))
	}

	// CREATE FORWARDS
	// Setup ibc auto-forwarding for a connection which doesn't exist
	stake := "stake"
	foreignHrp := "astro"
	sourceChannel := "channel-0"
	rec := bech32ibctypes.HrpIbcRecord{
		Hrp:               foreignHrp,
		SourceChannel:     sourceChannel,
		IcsToHeightOffset: 1000,
		IcsToTimeOffset:   1000,
	}
	input.GravityKeeper.bech32IbcKeeper.SetHrpIbcRecords(ctx, []bech32ibctypes.HrpIbcRecord{rec})
	hrpRecords := input.GravityKeeper.bech32IbcKeeper.GetHrpIbcRecords(ctx)
	require.Equal(t, len(hrpRecords), 1)
	require.Equal(t, hrpRecords[0], rec)
	forwards := input.GravityKeeper.PendingIbcAutoForwards(ctx, 0)
	require.Equal(t, 0, len(forwards))

	// Create pending forwards which must be preserved
	forwards = make([]*types.PendingIbcAutoForward, len(senders))
	for i, v := range senders {
		foreignRcv, err := bech32.ConvertAndEncode(foreignHrp, *v)
		require.NoError(t, err)
		c := contracts[i%len(contracts)]
		token, err := types.NewInternalERC20Token(sdk.NewInt(99999999), c.GetAddress().Hex())
		require.NoError(t, err)
		coins := sdk.NewCoins(token.GravityCoin())
		// Mint the coins to be forwarded, since the gravity module should already hold the tokens
		require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, coins))
		fwd := types.PendingIbcAutoForward{
			ForeignReceiver: foreignRcv,
			Token:           &coins[0],
			IbcChannel:      sourceChannel,
			EventNonce:      uint64(i + 1),
		}
		input.GravityKeeper.setLastObservedEventNonce(ctx, fwd.EventNonce)
		err = input.GravityKeeper.addPendingIbcAutoForward(ctx, fwd, stake)
		require.NoError(t, err)
		forwards[i] = &fwd
	}

	checkAllTransactionsExist(t, input.GravityKeeper, ctx, txs, forwards)
	exportImport(t, &input)
	checkAllTransactionsExist(t, input.GravityKeeper, ctx, txs, forwards)

	// Clear the pending ibc auto forwards so the invariant won't fail
	input.GravityKeeper.IteratePendingIbcAutoForwards(ctx, func(_ []byte, fwd *types.PendingIbcAutoForward) bool {
		require.NoError(t, input.GravityKeeper.deletePendingIbcAutoForward(ctx, fwd.EventNonce))
		require.NoError(t, input.BankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(*fwd.Token)))
		return false
	})
}

// Requires that all transactions in txs exist in keeper
func checkAllTransactionsExist(t *testing.T, keeper Keeper, ctx sdk.Context, txs []*types.InternalOutgoingTransferTx, forwards []*types.PendingIbcAutoForward) {
	unbatched := keeper.GetUnbatchedTransactions(ctx)
	batches := keeper.GetOutgoingTxBatches(ctx)
	// Collect all txs into an array
	var gotTxs []*types.InternalOutgoingTransferTx
	gotTxs = append(gotTxs, unbatched...)
	for _, batch := range batches {
		gotTxs = append(gotTxs, batch.Transactions...)
	}
	require.Equal(t, len(txs), len(gotTxs))
	// Sort both arrays for simple searching
	sort.Slice(gotTxs, func(i, j int) bool {
		return gotTxs[i].Id < gotTxs[j].Id
	})
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Id < txs[j].Id
	})
	// Actually check that the txs all exist, iterate on txs in case some got lost in the import/export step
	for i, exp := range txs {
		require.Equal(t, exp.Id, gotTxs[i].Id)
		require.Equal(t, exp.Erc20Fee, gotTxs[i].Erc20Fee)
		require.Equal(t, exp.Erc20Token, gotTxs[i].Erc20Token)
		require.Equal(t, exp.DestAddress.GetAddress(), gotTxs[i].DestAddress.GetAddress())
		require.Equal(t, exp.Sender.String(), gotTxs[i].Sender.String())
	}

	gotFwds := keeper.PendingIbcAutoForwards(ctx, 0)
	require.Equal(t, len(forwards), len(gotFwds))
	for i, fwd := range gotFwds {
		// The order should be preserved because of the type of iterator and the construction of `forwards`
		require.Equal(t, fwd, forwards[i])
	}
}

// Exports and then imports all bridge state, overwrites the `input` test environment to simulate chain restart
func exportImport(t *testing.T, input *TestInput) {
	bankGenesis := input.BankKeeper.ExportGenesis(input.Context)                                    // Required for ibc auto forwards
	bech32ibcGenesis := bech32ibc.ExportGenesis(input.Context, input.GravityKeeper.bech32IbcKeeper) // Required for ibc auto forwards
	genesisState := ExportGenesis(input.Context, input.GravityKeeper)
	newEnv := CreateTestEnv(t)
	input = &newEnv
	unbatched := input.GravityKeeper.GetUnbatchedTransactions(input.Context)
	require.Empty(t, unbatched)
	batches := input.GravityKeeper.GetOutgoingTxBatches(input.Context)
	require.Empty(t, batches)
	forwards := input.GravityKeeper.PendingIbcAutoForwards(input.Context, 0)
	require.Empty(t, forwards)
	bech32ibc.InitGenesis(input.Context, input.GravityKeeper.bech32IbcKeeper, *bech32ibcGenesis)
	input.BankKeeper.InitGenesis(input.Context, bankGenesis)
	InitGenesis(input.Context, input.GravityKeeper, genesisState)
}
