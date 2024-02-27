package gravity

import (
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint: exhaustruct
func TestHandleMsgSendToEth(t *testing.T) {
	input := keeper.CreateTestEnv(t)
	var (
		userCosmosAddr, e1               = sdk.AccAddressFromBech32("paloma1l2j8vaykh03zenzytntj3cza6zfxwlj68dd0l3")
		blockTime                        = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		blockHeight            int64     = 200
		testDenom                        = "ugrain"
		startingCoinAmount, _            = math.NewIntFromString("150000000000000000000")
		sendAmount, _                    = math.NewIntFromString("60000000000000000000")
		startingCoins          sdk.Coins = sdk.Coins{sdk.NewCoin(testDenom, startingCoinAmount)}
		sendingCoin            sdk.Coin  = sdk.NewCoin(testDenom, sendAmount)
		ethDestination                   = "0x3c9289da00b02dC623d0D8D907619890301D26d4"
		invalidEthDestinations           = []string{"obviously invalid", "0x3c9289da00b02dC623d0D8D907", "0x3c9289da00b02dC623d0D8D907dC623d0D8D907619890", "0x3c9289da00b02dC623d0D8D907619890301D26dU"}
	)
	require.NoError(t, e1)

	// we start by depositing some funds into the users balance to send
	sdkCtx := sdk.UnwrapSDKContext(input.Context)
	defer func() { sdkCtx.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	h := NewHandler(input.GravityKeeper)
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, startingCoins))
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, userCosmosAddr, startingCoins))
	balance1 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(testDenom, startingCoinAmount)}, balance1)

	// send some coins
	msg := &types.MsgSendToEth{
		EthDest:          ethDestination,
		Amount:           sendingCoin,
		ChainReferenceId: "test-chain",
		Metadata: valsettypes.MsgMetadata{
			Creator: userCosmosAddr.String(),
		},
	}
	ctx = sdkCtx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err := h(sdkCtx, msg)
	require.NoError(t, err)
	balance2 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(testDenom, startingCoinAmount.Sub(sendAmount))}, balance2)

	// do the same thing again and make sure it works twice
	msg1 := &types.MsgSendToEth{
		EthDest:          ethDestination,
		Amount:           sendingCoin,
		ChainReferenceId: "test-chain",
		Metadata: valsettypes.MsgMetadata{
			Creator: userCosmosAddr.String(),
		},
	}
	ctx = sdkCtx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err1 := h(sdkCtx, msg1)
	require.NoError(t, err1)
	balance3 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	finalAmount3 := startingCoinAmount.Sub(sendAmount).Sub(sendAmount)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(testDenom, finalAmount3)}, balance3)

	// now we should be out of coins and error
	msg2 := &types.MsgSendToEth{
		EthDest:          ethDestination,
		Amount:           sendingCoin,
		ChainReferenceId: "test-chain",
		Metadata: valsettypes.MsgMetadata{
			Creator: userCosmosAddr.String(),
		},
	}
	ctx = sdkCtx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err2 := h(sdkCtx, msg2)
	require.Error(t, err2)
	balance4 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(testDenom, finalAmount3)}, balance4)

	// these should all produce an error
	for _, val := range invalidEthDestinations {
		msg := &types.MsgSendToEth{
			EthDest:          val,
			Amount:           sendingCoin,
			ChainReferenceId: "test-chain",
			Metadata: valsettypes.MsgMetadata{
				Creator: userCosmosAddr.String(),
			},
		}
		ctx = sdkCtx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
		_, err := h(sdkCtx, msg)
		require.Error(t, err)
		balance := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
		assert.Equal(t, sdk.Coins{sdk.NewCoin(testDenom, finalAmount3)}, balance)
	}
}

const biggestInt = "115792089237316195423570985008687907853269984665640564039457584007913129639935" // 2^256 - 1

// We rely on BitLen() to detect Uint256 overflow, here we ensure BitLen() returns what we expect
func TestUint256BitLen(t *testing.T) {
	biggestBigInt, _ := new(big.Int).SetString(biggestInt, 10)
	require.Equal(t, 256, biggestBigInt.BitLen(), "expected 2^256 - 1 to be represented in 256 bits")
	biggerThanUint256 := biggestBigInt.Add(biggestBigInt, new(big.Int).SetInt64(1)) // add 1 to the max value of Uint256
	require.Equal(t, 257, biggerThanUint256.BitLen(), "expected 2^256 to be represented in 257 bits")
}

// nolint: exhaustruct
//func TestMsgSetOrchestratorAddresses(t *testing.T) {
//	var (
//		ethAddress, e1                 = types.NewEthAddress("0xb462864E395d88d6bc7C5dd5F3F5eb4cc2599255")
//		cosmosAddress   sdk.AccAddress = bytes.Repeat([]byte{0x1}, 20)
//		ethAddress2, e2                = types.NewEthAddress("0x26126048c706fB45a5a6De8432F428e794d0b952")
//		cosmosAddress2  sdk.AccAddress = bytes.Repeat([]byte{0x2}, 20)
//		blockTime                      = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
//		blockTime2                     = time.Date(2020, 9, 15, 15, 20, 10, 0, time.UTC)
//		blockHeight     int64          = 200
//		blockHeight2    int64          = 210
//	)
//	require.NoError(t, e1)
//	require.NoError(t, e2)
//	input, ctx := keeper.SetupTestChain(t, []uint64{1000000000}, false)
//	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()
//
//	wctx := sdk.WrapSDKContext(ctx)
//	k := input.GravityKeeper
//	h := NewHandler(input.GravityKeeper)
//	ctx = ctx.WithBlockTime(blockTime)
//	valAddress, err := sdk.ValAddressFromBech32(input.StakingKeeper.GetValidators(ctx, 10)[0].OperatorAddress)
//	require.NoError(t, err)
//
//	// test setting keys
//	msg := types.NewMsgSetOrchestratorAddress(valAddress, cosmosAddress, *ethAddress)
//	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
//	_, err = h(ctx, msg)
//	require.NoError(t, err)
//
//	// test all lookup methods
//
//	// individual lookups
//	ethLookup, found, err := k.GetEthAddressByValidator(ctx, valAddress)
//	require.NoError(t, err)
//	assert.True(t, found)
//	assert.Equal(t, ethLookup, ethAddress)
//
//	valLookup, found, err := k.GetOrchestratorValidator(ctx, cosmosAddress)
//	require.NoError(t, err)
//	assert.True(t, found)
//	assert.Equal(t, valLookup.GetOperator(), valAddress)
//
//	// query endpoints
//	queryO := types.QueryDelegateKeysByOrchestratorAddress{
//		OrchestratorAddress: cosmosAddress.String(),
//	}
//	_, err = k.GetDelegateKeyByOrchestrator(wctx, &queryO)
//	require.NoError(t, err)
//
//	queryE := types.QueryDelegateKeysByEthAddress{
//		EthAddress: ethAddress.GetAddress().Hex(),
//	}
//	_, err = k.GetDelegateKeyByEth(wctx, &queryE)
//	require.NoError(t, err)
//
//	// try to set values again. This should fail see issue #344 for why allowing this
//	// would require keeping a history of all validators delegate keys forever
//	msg = types.NewMsgSetOrchestratorAddress(valAddress, cosmosAddress2, *ethAddress2)
//	ctx = ctx.WithBlockTime(blockTime2).WithBlockHeight(blockHeight2)
//	_, err = h(ctx, msg)
//	require.Error(t, err)
//}
