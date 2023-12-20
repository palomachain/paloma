package keeper

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/require"
)

// Tests that batches and transactions are preserved during chain restart, including pending forwards
func TestBatchAndTxImportExport(t *testing.T) {
	// SETUP ENV + DATA
	// ==================
	input := CreateTestEnv(t)
	sdkCtx := sdk.UnwrapSDKContext(input.Context)

	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	batchSize := 100
	accAddresses := []string{ // Warning: this must match the length of ctrAddresses

		"paloma1dg55rtevlfxh46w88yjpdd08sqhh5cc37jmmth",
		"paloma164knshrzuuurf05qxf3q5ewpfnwzl4gjd7cwmp",
		"paloma193fw83ynn76328pty4yl7473vg9x86alc042em",
		"paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk",
		"paloma1ees2tqhhhm9ahlhceh2zdguww9lqn2ckyn7yg6",
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
		token, err := types.NewInternalERC20Token(math.NewInt(99999999), v.GetAddress().Hex(), "test-chain")
		tokens[i] = token
		allVouchers := sdk.NewCoins(sdk.NewCoin(testDenom, token.Amount))
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
	amounts := []int{51, 52, 53, 54, 55, 56, 57, 58, 59, 60}
	for i := 0; i < numTxs; i++ {
		// Pick amount, sender, receiver, and contract for the ith transaction
		// Sender and contract will always match up (they must since sender i controls the whole balance of the ith token)
		// Receivers should get a balance of many token types since i % len(receivers) is usually different than i % len(contracts)
		amount := amounts[i%len(amounts)]
		sender := senders[i%len(senders)]
		receiver := receivers[i%len(receivers)]
		contract := contracts[i%len(contracts)]
		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(amount)), contract.GetAddress().Hex(), "test-chain")
		require.NoError(t, err)

		// add transaction to the pool
		id, err := input.GravityKeeper.AddToOutgoingPool(ctx, *sender, *receiver, sdk.NewCoin(testDenom, amountToken.Amount), "test-chain")
		require.NoError(t, err)

		// Record the transaction for later testing
		tx, err := types.NewInternalOutgoingTransferTx(id, sender.String(), receiver.GetAddress().Hex(), amountToken.ToExternal())
		require.NoError(t, err)
		txs[i] = tx
	}

	// when

	now := time.Now().UTC()
	ctx = sdkCtx.WithBlockTime(now)

	// CREATE BATCHES
	// ==================
	// Want to create batches for half of the transactions for each contract
	// with 100 tx in each batch, 1000 txs per contract, we want 5 batches per contract to batch 500 txs per contract
	batches := make([]*types.InternalOutgoingTxBatch, 5*len(contracts))
	for i, v := range contracts {
		batch, err := input.GravityKeeper.BuildOutgoingTXBatch(ctx, "test-chain", *v, uint(batchSize))
		require.NoError(t, err)
		batches[i] = batch
		sdkCtx.Logger().Info(fmt.Sprintf("Created batch %v for contract %v with %v transactions", i, v.GetAddress(), batchSize))
	}
}
