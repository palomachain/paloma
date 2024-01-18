package gravity

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/palomachain/paloma/x/gravity/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNonValidatorBatchConfirm(t *testing.T) {
	//	Test if a non-validator confirm won't panic

	input, ctx := keeper.SetupFiveValChain(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	pk := input.GravityKeeper
	params := pk.GetParams(ctx)

	// Create not nice guy with very little stake
	consPrivKey := ed25519.GenPrivKey()
	consPubKey := consPrivKey.PubKey()
	valPrivKey := secp256k1.GenPrivKey()
	valPubKey := valPrivKey.PubKey()
	valAddr := sdk.ValAddress(valPubKey.Address())
	accAddr := sdk.AccAddress(valPubKey.Address())

	// Initialize the account for the key
	acc := input.AccountKeeper.NewAccount(
		input.Context,
		authtypes.NewBaseAccount(accAddr, valPubKey, 0, 0),
	)

	require.NoError(t, input.BankKeeper.MintCoins(input.Context, types.ModuleName, keeper.InitCoins))
	err := input.BankKeeper.SendCoinsFromModuleToAccount(
		input.Context,
		types.ModuleName,
		accAddr,
		keeper.InitCoins,
	)
	require.NoError(t, err)

	// Set the account in state
	input.AccountKeeper.SetAccount(input.Context, acc)

	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(&input.StakingKeeper)

	_, err = stakingMsgSvr.CreateValidator(input.Context, keeper.NewTestMsgCreateValidator(valAddr, consPubKey, math.NewIntFromUint64(1)))
	require.NoError(t, err)

	// Run the staking endblocker to ensure valset is correct in state
	_, err = input.StakingKeeper.EndBlocker(input.Context)
	require.NoError(t, err)

	ethAddr, err := types.NewEthAddress("0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B")
	require.NoError(t, err)

	notNiceVal, found, err := pk.GetOrchestratorValidator(ctx, accAddr)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, notNiceVal.Status, stakingtypes.Unbonded)

	// First store a batch

	batch, err := types.NewInternalOutgingTxBatchFromExternalBatch(types.OutgoingTxBatch{
		BatchNonce:         1,
		BatchTimeout:       0,
		Transactions:       []types.OutgoingTransferTx{},
		TokenContract:      keeper.TokenContractAddrs[0],
		PalomaBlockCreated: uint64(sdkCtx.BlockHeight() - int64(params.SignedBatchesWindow+1)),
	})
	require.NoError(t, err)
	pk.StoreBatch(ctx, *batch)
	unslashedBatches, err := pk.GetUnSlashedBatches(ctx, uint64(sdkCtx.BlockHeight()))
	require.NoError(t, err)
	assert.True(t, len(unslashedBatches) == 1 && unslashedBatches[0].BatchNonce == 1)

	for i, orch := range keeper.OrchAddrs {
		pk.SetBatchConfirm(ctx, &types.MsgConfirmBatch{
			Nonce:         batch.BatchNonce,
			TokenContract: keeper.TokenContractAddrs[0],
			EthSigner:     keeper.EthAddrs[i].String(),
			Orchestrator:  orch.String(),
			Signature:     "",
		})
	}

	// Sign using our not nice validator
	// This is not really possible if we use confirmHandlerCommon
	pk.SetBatchConfirm(ctx, &types.MsgConfirmBatch{
		Nonce:         batch.BatchNonce,
		TokenContract: keeper.TokenContractAddrs[0],
		EthSigner:     ethAddr.GetAddress().Hex(),
		Orchestrator:  accAddr.String(),
		Signature:     "",
	})

	// Now remove all the stake
	_, err = stakingMsgSvr.Undelegate(input.Context, keeper.NewTestMsgUnDelegateValidator(valAddr, math.NewIntFromUint64(1)))
	require.NoError(t, err)

	EndBlocker(ctx, pk)
}

// Test batch timeout
func TestBatchTimeout(t *testing.T) {
	input, ctx := keeper.SetupFiveValChain(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	pk := input.GravityKeeper
	params := pk.GetParams(ctx)
	var (
		testTime         = time.Unix(1693424690, 0)
		mySender, e1     = sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
		myReceiver       = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		testERC20Address = "0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"
		testDenom        = "ugrain"
		token, e2        = types.NewInternalERC20Token(math.NewInt(99999), testERC20Address, "test-chain")
		allVouchers      = sdk.NewCoins(sdk.NewCoin(testDenom, token.Amount))
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract, err := types.NewEthAddress(testERC20Address)
	require.NoError(t, err)

	require.Greater(t, params.AverageBlockTime, uint64(0))
	require.Greater(t, params.AverageEthereumBlockTime, uint64(0))

	// mint some vouchers first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mySender, allVouchers))

	// add some TX to the pool
	for i := 0; i < 6; i++ {
		amountToken, err := types.NewInternalERC20Token(math.NewInt(int64(i+100)), testERC20Address, "test-chain")
		require.NoError(t, err)
		amount := sdk.NewCoin(testDenom, amountToken.Amount)

		_, err = input.GravityKeeper.AddToOutgoingPool(ctx, mySender, *receiver, amount, "test-chain")
		require.NoError(t, err)
	}

	// when
	ctx = sdkCtx.WithBlockTime(testTime)

	// check that we can make a batch without first setting an ethereum block height
	b1, err1 := pk.BuildOutgoingTXBatch(ctx, "test-chain", *tokenContract, 1)
	require.NoError(t, err1)
	require.Equal(t, b1.BatchTimeout, uint64(1693425290))

	pk.SetLastObservedEthereumBlockHeight(ctx, 500)

	// create another batch
	b2, err2 := pk.BuildOutgoingTXBatch(ctx, "test-chain", *tokenContract, 2)
	require.NoError(t, err2)
	// this is exactly block 500 plus twelve hours
	require.Equal(t, b2.BatchTimeout, uint64(1693425290))

	// make sure the batches got stored in the first place
	gotFirstBatch, err := input.GravityKeeper.GetOutgoingTXBatch(ctx, b1.TokenContract, b1.BatchNonce)
	require.NoError(t, err)
	require.NotNil(t, gotFirstBatch)
	gotSecondBatch, err := input.GravityKeeper.GetOutgoingTXBatch(ctx, b2.TokenContract, b2.BatchNonce)
	require.NoError(t, err)
	require.NotNil(t, gotSecondBatch)

	// persist confirmations for second batch to test their deletion on batch timeout
	for i, orch := range keeper.OrchAddrs {
		ethAddr, err := types.NewEthAddress(keeper.EthAddrs[i].String())
		require.NoError(t, err)

		conf := &types.MsgConfirmBatch{
			Nonce:         b2.BatchNonce,
			TokenContract: b2.TokenContract.GetAddress().Hex(),
			EthSigner:     ethAddr.GetAddress().Hex(),
			Orchestrator:  orch.String(),
			Signature:     "dummysig",
		}

		input.GravityKeeper.SetBatchConfirm(ctx, conf)
	}

	// verify that confirms are persisted
	secondBatchConfirms, err := input.GravityKeeper.GetBatchConfirmByNonceAndTokenContract(ctx, b2.BatchNonce, b2.TokenContract)
	require.NoError(t, err)
	require.Equal(t, len(keeper.OrchAddrs), len(secondBatchConfirms))

	// when, beyond the timeout
	ctx = sdkCtx.WithBlockTime(testTime.Add(20 * time.Minute))

	EndBlocker(ctx, pk)

	// this had a timeout of zero should be deleted.
	gotFirstBatch, err = input.GravityKeeper.GetOutgoingTXBatch(ctx, b1.TokenContract, b1.BatchNonce)
	require.NoError(t, err)
	require.Nil(t, gotFirstBatch)
	// make sure the end blocker does not delete these, as the block height has not officially
	// been updated by a relay event
	gotSecondBatch, err = input.GravityKeeper.GetOutgoingTXBatch(ctx, b2.TokenContract, b2.BatchNonce)
	require.NoError(t, err)
	require.Nil(t, gotSecondBatch)

	// verify that second batch confirms are deleted
	secondBatchConfirms, err = input.GravityKeeper.GetBatchConfirmByNonceAndTokenContract(ctx, b2.BatchNonce, b2.TokenContract)
	require.NoError(t, err)
	require.Equal(t, 0, len(secondBatchConfirms))
}
