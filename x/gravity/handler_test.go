package gravity

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palomachain/paloma/x/gravity/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
)

// nolint: exhaustruct
func TestHandleMsgSendToEth(t *testing.T) {
	var (
		userCosmosAddr, e1               = sdk.AccAddressFromBech32("gravity1990z7dqsvh8gthw9pa5sn4wuy2xrsd80lcx6lv")
		blockTime                        = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		blockHeight            int64     = 200
		denom                            = "gravity0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e"
		startingCoinAmount, _            = sdk.NewIntFromString("150000000000000000000") // 150 ETH worth, required to reach above u64 limit (which is about 18 ETH)
		sendAmount, _                    = sdk.NewIntFromString("50000000000000000000")  // 50 ETH
		feeAmount, _                     = sdk.NewIntFromString("5000000000000000000")   // 5 ETH
		startingCoins          sdk.Coins = sdk.Coins{sdk.NewCoin(denom, startingCoinAmount)}
		sendingCoin            sdk.Coin  = sdk.NewCoin(denom, sendAmount)
		feeCoin                sdk.Coin  = sdk.NewCoin(denom, feeAmount)
		ethDestination                   = "0x3c9289da00b02dC623d0D8D907619890301D26d4"
		invalidEthDestinations           = []string{"obviously invalid", "0x3c9289da00b02dC623d0D8D907", "0x3c9289da00b02dC623d0D8D907dC623d0D8D907619890", "0x3c9289da00b02dC623d0D8D907619890301D26dU"}
	)
	require.NoError(t, e1)

	// we start by depositing some funds into the users balance to send
	input := keeper.CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := input.Context
	h := NewHandler(input.GravityKeeper)
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, startingCoins))
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, userCosmosAddr, startingCoins))
	balance1 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, startingCoinAmount)}, balance1)

	// send some coins
	msg := &types.MsgSendToEth{
		Sender:    userCosmosAddr.String(),
		EthDest:   ethDestination,
		Amount:    sendingCoin,
		BridgeFee: feeCoin}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err := h(ctx, msg)
	require.NoError(t, err)
	balance2 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, startingCoinAmount.Sub(sendAmount).Sub(feeAmount))}, balance2)

	// do the same thing again and make sure it works twice
	msg1 := &types.MsgSendToEth{
		Sender:    userCosmosAddr.String(),
		EthDest:   ethDestination,
		Amount:    sendingCoin,
		BridgeFee: feeCoin}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err1 := h(ctx, msg1)
	require.NoError(t, err1)
	balance3 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	finalAmount3 := startingCoinAmount.Sub(sendAmount).Sub(sendAmount).Sub(feeAmount).Sub(feeAmount)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, finalAmount3)}, balance3)

	// now we should be out of coins and error
	msg2 := &types.MsgSendToEth{
		Sender:    userCosmosAddr.String(),
		EthDest:   ethDestination,
		Amount:    sendingCoin,
		BridgeFee: feeCoin}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err2 := h(ctx, msg2)
	require.Error(t, err2)
	balance4 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, finalAmount3)}, balance4)

	// these should all produce an error
	for _, val := range invalidEthDestinations {
		msg := &types.MsgSendToEth{
			Sender:    userCosmosAddr.String(),
			EthDest:   val,
			Amount:    sendingCoin,
			BridgeFee: feeCoin}
		ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
		_, err := h(ctx, msg)
		require.Error(t, err)
		balance := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
		assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, finalAmount3)}, balance)
	}

}

// nolint: exhaustruct
func TestMsgSendToCosmosClaim(t *testing.T) {
	var (
		myCosmosAddr, e1 = sdk.AccAddressFromBech32("gravity16ahjkfqxpp6lvfy9fpfnfjg39xr96qet0l08hu")
		anyETHAddr       = "0xf9613b532673Cc223aBa451dFA8539B87e1F666D"
		tokenETHAddr     = "0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"
		myBlockTime      = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		amountA, _       = sdk.NewIntFromString("50000000000000000000")  // 50 ETH
		amountB, _       = sdk.NewIntFromString("100000000000000000000") // 100 ETH
	)
	require.NoError(t, e1)
	input, ctx := keeper.SetupFiveValChain(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	h := NewHandler(input.GravityKeeper)

	myErc20 := types.ERC20Token{
		Amount:   amountA,
		Contract: tokenETHAddr,
	}

	// send attestations from all five validators
	for _, v := range keeper.OrchAddrs {
		ethClaim := types.MsgSendToCosmosClaim{
			EventNonce:     uint64(1),
			TokenContract:  myErc20.Contract,
			Amount:         myErc20.Amount,
			EthereumSender: anyETHAddr,
			CosmosReceiver: myCosmosAddr.String(),
			Orchestrator:   v.String(),
		}
		// each msg goes into it's own block
		ctx = ctx.WithBlockTime(myBlockTime)
		_, err := h(ctx, &ethClaim)
		EndBlocker(ctx, input.GravityKeeper)
		require.NoError(t, err)

		// and attestation persisted
		hash, err := ethClaim.ClaimHash()
		require.NoError(t, err)
		a := input.GravityKeeper.GetAttestation(ctx, uint64(1), hash)
		require.NotNil(t, a)

		// Test to reject duplicate deposit
		// when
		ctx = ctx.WithBlockTime(myBlockTime)
		_, err = h(ctx, &ethClaim)
		EndBlocker(ctx, input.GravityKeeper)
		// then
		require.Error(t, err)
	}

	// and vouchers added to the account
	balance := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("gravity0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", amountA)}, balance)

	// send attestations from all five validators
	for _, v := range keeper.OrchAddrs {
		// Test to reject skipped nonce
		ethClaim := types.MsgSendToCosmosClaim{
			EventNonce:     uint64(3),
			TokenContract:  tokenETHAddr,
			Amount:         amountA,
			EthereumSender: anyETHAddr,
			CosmosReceiver: myCosmosAddr.String(),
			Orchestrator:   v.String(),
		}

		// when
		ctx = ctx.WithBlockTime(myBlockTime)
		_, err := h(ctx, &ethClaim)
		EndBlocker(ctx, input.GravityKeeper)
		// then
		require.Error(t, err)
	}

	balance = input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("gravity0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", amountA)}, balance)

	// send attestations from all five validators
	for _, v := range keeper.OrchAddrs {
		// Test to finally accept consecutive nonce
		ethClaim := types.MsgSendToCosmosClaim{
			EventNonce:     uint64(2),
			Amount:         amountA,
			TokenContract:  tokenETHAddr,
			EthereumSender: anyETHAddr,
			CosmosReceiver: myCosmosAddr.String(),
			Orchestrator:   v.String(),
		}

		// when
		ctx = ctx.WithBlockTime(myBlockTime)
		_, err := h(ctx, &ethClaim)
		EndBlocker(ctx, input.GravityKeeper)

		// then
		require.NoError(t, err)
	}

	balance = input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("gravity0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", amountB)}, balance)
}

// nolint: exhaustruct
func TestEthereumBlacklist(t *testing.T) {
	var (
		myCosmosAddr, e1 = sdk.AccAddressFromBech32("gravity16ahjkfqxpp6lvfy9fpfnfjg39xr96qet0l08hu")
		anyETHSender     = "0xf9613b532673Cc223aBa451dFA8539B87e1F666D"
		tokenETHAddr     = "0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"
		denom            = "gravity0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e"
		myBlockTime      = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		amountA, _       = sdk.NewIntFromString("50000000000000000000") // 50 ETH
	)
	require.NoError(t, e1)
	input, ctx := keeper.SetupFiveValChain(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	h := NewHandler(input.GravityKeeper)

	myErc20 := types.ERC20Token{
		Amount:   amountA,
		Contract: tokenETHAddr,
	}

	k := input.GravityKeeper
	blockedAddress := anyETHSender
	newParams := k.GetParams(ctx)

	newParams.EthereumBlacklist = []string{blockedAddress}

	k.SetParams(ctx, newParams)

	assert.Equal(t, k.GetParams(ctx).EthereumBlacklist, []string{blockedAddress})

	// send attestations from all five validators
	for _, v := range keeper.OrchAddrs {
		ethClaim := types.MsgSendToCosmosClaim{
			EventNonce:     uint64(1),
			TokenContract:  myErc20.Contract,
			Amount:         myErc20.Amount,
			EthereumSender: anyETHSender,
			CosmosReceiver: myCosmosAddr.String(),
			Orchestrator:   v.String(),
		}
		// each msg goes into it's own block
		ctx = ctx.WithBlockTime(myBlockTime)
		_, err := h(ctx, &ethClaim)
		EndBlocker(ctx, input.GravityKeeper)
		require.NoError(t, err)

		// and attestation persisted
		hash, err := ethClaim.ClaimHash()
		require.NoError(t, err)
		a := input.GravityKeeper.GetAttestation(ctx, uint64(1), hash)
		require.NotNil(t, a)

		// Test to reject duplicate deposit
		// when
		ctx = ctx.WithBlockTime(myBlockTime)
		_, err = h(ctx, &ethClaim)
		EndBlocker(ctx, input.GravityKeeper)
		// then
		require.Error(t, err)
	}

	// and vouchers added to the account
	balance := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.NotEqual(t, sdk.Coins{sdk.NewCoin(denom, amountA)}, balance)

	// Make sure that the balance is empty since funds should be sent to the community pool
	assert.Equal(t, balance, sdk.Coins{})

	// Check community pool has received the money instead of the address
	community_pool_balance := input.DistKeeper.GetFeePool(ctx).CommunityPool
	assert.Equal(t, sdk.NewDecFromInt(amountA), community_pool_balance.AmountOf(denom))

}

const biggestInt = "115792089237316195423570985008687907853269984665640564039457584007913129639935" // 2^256 - 1

// We rely on BitLen() to detect Uint256 overflow, here we ensure BitLen() returns what we expect
func TestUint256BitLen(t *testing.T) {
	biggestBigInt, _ := new(big.Int).SetString(biggestInt, 10)
	require.Equal(t, 256, biggestBigInt.BitLen(), "expected 2^256 - 1 to be represented in 256 bits")
	biggerThanUint256 := biggestBigInt.Add(biggestBigInt, new(big.Int).SetInt64(1)) // add 1 to the max value of Uint256
	require.Equal(t, 257, biggerThanUint256.BitLen(), "expected 2^256 to be represented in 257 bits")
}

// sendSendToCosmosClaim is a minor utility function that pairs with the five validator test environment
// allowing us to easily send 5 claims easily
func sendSendToCosmosClaim(msg types.MsgSendToCosmosClaim, ctx sdk.Context, h sdk.Handler, t *testing.T) {
	// send attestations from all five validators
	for _, v := range keeper.OrchAddrs {
		msg.Orchestrator = v.String()
		_, err := h(ctx, &msg)
		require.NoError(t, err)
	}
}

func TestMsgSendToCosmosOverflow(t *testing.T) {
	const grandeInt = "115792089237316195423570985008687907853269984665640564039457584007913129639835" // 2^256 - 101
	var (
		biggestBigInt, _     = new(big.Int).SetString(biggestInt, 10)
		grandeBigInt, _      = new(big.Int).SetString(grandeInt, 10)
		myCosmosAddr, e1     = sdk.AccAddressFromBech32("gravity16ahjkfqxpp6lvfy9fpfnfjg39xr96qet0l08hu")
		myNonce              = uint64(1)
		anyETHAddr           = "0xf9613b532673Cc223aBa451dFA8539B87e1F666D"
		tokenETHAddr1        = "0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"
		tokenETHAddr2        = "0x429881672b9AE42b8eBA0e26cd9c73711b891ca6"
		myBlockTime          = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		tokenEthAddress1, e2 = types.NewEthAddress(tokenETHAddr1)
		tokenEthAddress2, e3 = types.NewEthAddress(tokenETHAddr2)
		denom1               = types.GravityDenom(*tokenEthAddress1)
		denom2               = types.GravityDenom(*tokenEthAddress2)
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)

	// Totally valid, but we're 101 away from the supply limit
	almostTooMuch := types.ERC20Token{
		Amount:   sdk.NewIntFromBigInt(grandeBigInt),
		Contract: tokenETHAddr1,
	}
	// This takes us past the supply limit of 2^256 - 1
	exactlyTooMuch := types.ERC20Token{
		Amount:   sdk.NewInt(101),
		Contract: tokenETHAddr1,
	}
	almostTooMuchClaim := types.MsgSendToCosmosClaim{
		EventNonce:     myNonce,
		EthBlockHeight: 0,
		TokenContract:  almostTooMuch.Contract,
		Amount:         almostTooMuch.Amount,
		EthereumSender: anyETHAddr,
		CosmosReceiver: myCosmosAddr.String(),
		Orchestrator:   "",
	}
	exactlyTooMuchClaim := types.MsgSendToCosmosClaim{
		EventNonce:     myNonce + 1,
		EthBlockHeight: 0,
		TokenContract:  exactlyTooMuch.Contract,
		Amount:         exactlyTooMuch.Amount,
		EthereumSender: anyETHAddr,
		CosmosReceiver: myCosmosAddr.String(),
		Orchestrator:   "",
	}
	// Absoulte max value of 2^256 - 1. Previous versions (v0.43 or v0.44) of cosmos-sdk did not support sdk.Int of this size
	maxSend := types.ERC20Token{
		Amount:   sdk.NewIntFromBigInt(biggestBigInt),
		Contract: tokenETHAddr2,
	}
	maxSendClaim := types.MsgSendToCosmosClaim{
		EventNonce:     myNonce + 2,
		EthBlockHeight: 0,
		TokenContract:  maxSend.Contract,
		Amount:         maxSend.Amount,
		EthereumSender: anyETHAddr,
		CosmosReceiver: myCosmosAddr.String(),
		Orchestrator:   "",
	}
	// Setup
	input, ctx := keeper.SetupFiveValChain(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	h := NewHandler(input.GravityKeeper)

	// Require that no tokens were bridged previously
	preSupply1 := input.BankKeeper.GetSupply(ctx, denom1)
	require.Equal(t, sdk.NewInt(0), preSupply1.Amount)

	fmt.Println("<<<<START Expecting to see 'minted coins from module account		module=x/bank amount={2^256 - 101}'")
	// Execute the 2^256 - 101 transaction
	ctx = ctx.WithBlockTime(myBlockTime)
	sendSendToCosmosClaim(almostTooMuchClaim, ctx, h, t)
	EndBlocker(ctx, input.GravityKeeper)
	// Require that the actual bridged amount is equal to the amount in our almostTooBigClaim
	middleSupply := input.BankKeeper.GetSupply(ctx, denom1)
	require.Equal(t, almostTooMuch.Amount, middleSupply.Amount.Sub(preSupply1.Amount))
	fmt.Println("END>>>>")

	fmt.Println("<<<<START Expecting to see an error about 'invalid supply after SendToCosmos attestation'")
	// Execute the 101 transaction
	ctx = ctx.WithBlockTime(myBlockTime)
	sendSendToCosmosClaim(exactlyTooMuchClaim, ctx, h, t)
	EndBlocker(ctx, input.GravityKeeper)
	// Require that the overflowing amount 101 was not bridged, and instead reverted
	endSupply := input.BankKeeper.GetSupply(ctx, denom1)
	require.Equal(t, almostTooMuch.Amount, endSupply.Amount.Sub(preSupply1.Amount))
	fmt.Println("END>>>>")

	// Require that no tokens were bridged previously
	preSupply2 := input.BankKeeper.GetSupply(ctx, denom2)
	require.Equal(t, sdk.NewInt(0), preSupply2.Amount)

	fmt.Println("<<<<START Expecting to see 'minted coins from module account		module=x/bank amount={2^256 - 1}'")
	// Execute the 2^256 - 1 transaction
	ctx = ctx.WithBlockTime(myBlockTime)
	sendSendToCosmosClaim(maxSendClaim, ctx, h, t)
	EndBlocker(ctx, input.GravityKeeper)
	// Require that the actual bridged amount is equal to the amount in our almostTooBigClaim
	maxSendSupply := input.BankKeeper.GetSupply(ctx, denom2)
	require.Equal(t, maxSend.Amount, maxSendSupply.Amount.Sub(preSupply2.Amount))
	fmt.Println("END>>>>")
}

// nolint: exhaustruct
func TestMsgSendToCosmosClaimSpreadVotes(t *testing.T) {
	var (
		myCosmosAddr, e1 = sdk.AccAddressFromBech32("gravity16ahjkfqxpp6lvfy9fpfnfjg39xr96qet0l08hu")
		myNonce          = uint64(1)
		anyETHAddr       = "0xf9613b532673Cc223aBa451dFA8539B87e1F666D"
		tokenETHAddr     = "0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"
		myBlockTime      = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
	)
	require.NoError(t, e1)
	input, ctx := keeper.SetupFiveValChain(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	h := NewHandler(input.GravityKeeper)

	myErc20 := types.ERC20Token{
		Amount:   sdk.NewInt(12),
		Contract: tokenETHAddr,
	}

	ethClaim := types.MsgSendToCosmosClaim{
		EventNonce:     myNonce,
		TokenContract:  myErc20.Contract,
		Amount:         myErc20.Amount,
		EthereumSender: anyETHAddr,
		CosmosReceiver: myCosmosAddr.String(),
		Orchestrator:   "",
	}

	for i := range []int{0, 1, 2} {
		// when
		ctx = ctx.WithBlockTime(myBlockTime)
		ethClaim.Orchestrator = keeper.OrchAddrs[i].String()
		_, err := h(ctx, &ethClaim)
		EndBlocker(ctx, input.GravityKeeper)
		require.NoError(t, err)

		// and attestation persisted
		hash, err := ethClaim.ClaimHash()
		require.NoError(t, err)
		a1 := input.GravityKeeper.GetAttestation(ctx, myNonce, hash)
		require.NotNil(t, a1)
		// and vouchers not yet added to the account
		balance1 := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
		assert.NotEqual(t, sdk.Coins{sdk.NewInt64Coin("gravity0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", 12)}, balance1)
	}

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	ethClaim.Orchestrator = keeper.OrchAddrs[3].String()
	_, err := h(ctx, &ethClaim)
	EndBlocker(ctx, input.GravityKeeper)
	require.NoError(t, err)

	// and attestation persisted
	hash, err := ethClaim.ClaimHash()
	require.NoError(t, err)
	a2 := input.GravityKeeper.GetAttestation(ctx, myNonce, hash)
	require.NotNil(t, a2)
	// and vouchers now added to the account
	balance2 := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewInt64Coin("gravity0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", 12)}, balance2)

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	ethClaim.Orchestrator = keeper.OrchAddrs[4].String()
	_, err = h(ctx, &ethClaim)
	EndBlocker(ctx, input.GravityKeeper)
	require.NoError(t, err)

	// and attestation persisted
	hash, err = ethClaim.ClaimHash()
	require.NoError(t, err)
	a3 := input.GravityKeeper.GetAttestation(ctx, myNonce, hash)
	require.NotNil(t, a3)
	// and no additional added to the account
	balance3 := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewInt64Coin("gravity0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e", 12)}, balance3)
}

// Tests sending funds to a native account and to that same account with a foreign prefix
// The SendToCosmosClaims should modify the balance of the underlying account
func TestMsgSendToCosmosForeignPrefixedAddress(t *testing.T) {
	var (
		coreAddress          = "6ahjkfqxpp6lvfy9fpfnfjg39xr96qet"
		myForeignBytes, err0 = types.IBCAddressFromBech32("levity1" + coreAddress + "vanuy5")
		myForeignAddr        = sdk.AccAddress(myForeignBytes)
		myNativeAddr, err1   = sdk.AccAddressFromBech32("gravity1" + coreAddress + "0l08hu")

		myNonce      = uint64(1)
		anyETHAddr   = "0xf9613b532673Cc223aBa451dFA8539B87e1F666D"
		tokenETHAddr = "0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"
		myBlockTime  = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
	)
	require.NoError(t, err0)
	require.NoError(t, err1)
	input, ctx := keeper.SetupFiveValChain(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	k := input.GravityKeeper

	h := NewHandler(k)

	myErc20 := types.ERC20Token{
		Amount:   sdk.NewInt(12),
		Contract: tokenETHAddr,
	}

	myTokenAddress, err := types.NewEthAddress(myErc20.Contract)
	require.NoError(t, err)
	_, erc20Denom := k.ERC20ToDenomLookup(ctx, *myTokenAddress)

	foreignEthClaim := types.MsgSendToCosmosClaim{
		EventNonce:     myNonce + 0,
		EthBlockHeight: 0,
		TokenContract:  myErc20.Contract,
		Amount:         myErc20.Amount,
		EthereumSender: anyETHAddr,
		CosmosReceiver: myForeignAddr.String(),
		Orchestrator:   "",
	}

	nativeEthClaim := types.MsgSendToCosmosClaim{
		EventNonce:     myNonce + 1,
		EthBlockHeight: 0,
		TokenContract:  myErc20.Contract,
		Amount:         myErc20.Amount,
		EthereumSender: anyETHAddr,
		CosmosReceiver: myNativeAddr.String(),
		Orchestrator:   "",
	}
	fmt.Println("myForeignAddr initial balance:", input.BankKeeper.GetAllBalances(ctx, myForeignAddr))
	fmt.Println("myNativeAddr initial balance:", input.BankKeeper.GetAllBalances(ctx, myNativeAddr))

	fmt.Println("Sending", myErc20.Amount, "to", myForeignAddr)
	ctx = ctx.WithBlockTime(myBlockTime)
	sendSendToCosmosClaim(foreignEthClaim, ctx, h, t)
	EndBlocker(ctx, input.GravityKeeper)
	foreignBals := input.BankKeeper.GetAllBalances(ctx, myForeignAddr)
	require.Equal(t, foreignBals, sdk.NewCoins(sdk.NewCoin(erc20Denom, myErc20.Amount)))

	fmt.Println("Sending", myErc20.Amount, "to", myNativeAddr)
	ctx = ctx.WithBlockTime(myBlockTime)
	sendSendToCosmosClaim(nativeEthClaim, ctx, h, t)
	EndBlocker(ctx, input.GravityKeeper)
	nativeBals := input.BankKeeper.GetAllBalances(ctx, myForeignAddr)
	expectedDoubleBalance := myErc20.Amount.Mul(sdk.NewInt(2))
	require.Equal(t, nativeBals, sdk.NewCoins(sdk.NewCoin(erc20Denom, expectedDoubleBalance)))
}

// nolint: exhaustruct
func TestMsgSetOrchestratorAddresses(t *testing.T) {
	var (
		ethAddress, e1                 = types.NewEthAddress("0xb462864E395d88d6bc7C5dd5F3F5eb4cc2599255")
		cosmosAddress   sdk.AccAddress = bytes.Repeat([]byte{0x1}, 20)
		ethAddress2, e2                = types.NewEthAddress("0x26126048c706fB45a5a6De8432F428e794d0b952")
		cosmosAddress2  sdk.AccAddress = bytes.Repeat([]byte{0x2}, 20)
		blockTime                      = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		blockTime2                     = time.Date(2020, 9, 15, 15, 20, 10, 0, time.UTC)
		blockHeight     int64          = 200
		blockHeight2    int64          = 210
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	input, ctx := keeper.SetupTestChain(t, []uint64{1000000000}, false)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	wctx := sdk.WrapSDKContext(ctx)
	k := input.GravityKeeper
	h := NewHandler(input.GravityKeeper)
	ctx = ctx.WithBlockTime(blockTime)
	valAddress, err := sdk.ValAddressFromBech32(input.StakingKeeper.GetValidators(ctx, 10)[0].OperatorAddress)
	require.NoError(t, err)

	// test setting keys
	msg := types.NewMsgSetOrchestratorAddress(valAddress, cosmosAddress, *ethAddress)
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err = h(ctx, msg)
	require.NoError(t, err)

	// test all lookup methods

	// individual lookups
	ethLookup, found := k.GetEthAddressByValidator(ctx, valAddress)
	assert.True(t, found)
	assert.Equal(t, ethLookup, ethAddress)

	valLookup, found := k.GetOrchestratorValidator(ctx, cosmosAddress)
	assert.True(t, found)
	assert.Equal(t, valLookup.GetOperator(), valAddress)

	// query endpoints
	queryO := types.QueryDelegateKeysByOrchestratorAddress{
		OrchestratorAddress: cosmosAddress.String(),
	}
	_, err = k.GetDelegateKeyByOrchestrator(wctx, &queryO)
	require.NoError(t, err)

	queryE := types.QueryDelegateKeysByEthAddress{
		EthAddress: ethAddress.GetAddress().Hex(),
	}
	_, err = k.GetDelegateKeyByEth(wctx, &queryE)
	require.NoError(t, err)

	// try to set values again. This should fail see issue #344 for why allowing this
	// would require keeping a history of all validators delegate keys forever
	msg = types.NewMsgSetOrchestratorAddress(valAddress, cosmosAddress2, *ethAddress2)
	ctx = ctx.WithBlockTime(blockTime2).WithBlockHeight(blockHeight2)
	_, err = h(ctx, msg)
	require.Error(t, err)
}

// TestMsgValsetConfirm ensures that the valset confirm message sets a validator set confirm
// in the store and validates the signature
func TestMsgValsetConfirm(t *testing.T) {
	var (
		blockTime          = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		blockHeight  int64 = 200
		signature          = "7c331bd8f2f586b04a2e2cafc6542442ef52e8b8be49533fa6b8962e822bc01e295a62733abfd65a412a8de8286f2794134c160c27a2827bdb71044b94b003cc1c"
		badSignature       = "6c331bd8f2f586b04a2e2cafc6542442ef52e8b8be49533fa6b8962e822bc01e295a62733abfd65a412a8de8286f2794134c160c27a2827bdb71044b94b003cc1c"
		ethAddress         = "0xd62FF457C6165FF214C1658c993A8a203E601B03"
		wrongAddress       = "0xb9a2c7853F181C3dd4a0517FCb9470C0f709C08C"
	)
	ethAddressParsed, err := types.NewEthAddress(ethAddress)
	require.NoError(t, err)

	input, ctx := keeper.SetupFiveValChain(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	k := input.GravityKeeper
	h := NewHandler(input.GravityKeeper)

	// set a validator set in the store
	vs, err := k.GetCurrentValset(ctx)
	require.NoError(t, err)
	vs.Height = uint64(1)
	vs.Nonce = uint64(1)
	k.StoreValset(ctx, vs)
	k.SetLatestValsetNonce(ctx, vs.Nonce)
	k.SetEthAddressForValidator(input.Context, keeper.ValAddrs[0], *ethAddressParsed)

	// try wrong eth address
	msg := &types.MsgValsetConfirm{
		Nonce:        1,
		Orchestrator: keeper.OrchAddrs[0].String(),
		EthAddress:   wrongAddress,
		Signature:    signature,
	}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err = h(ctx, msg)
	require.Error(t, err)

	// try a nonexisting valset
	msg = &types.MsgValsetConfirm{
		Nonce:        10,
		Orchestrator: keeper.OrchAddrs[0].String(),
		EthAddress:   ethAddress,
		Signature:    signature,
	}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err = h(ctx, msg)
	require.Error(t, err)

	// try a bad signature
	msg = &types.MsgValsetConfirm{
		Nonce:        1,
		Orchestrator: keeper.OrchAddrs[0].String(),
		EthAddress:   ethAddress,
		Signature:    badSignature,
	}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err = h(ctx, msg)
	require.Error(t, err)

	msg = &types.MsgValsetConfirm{
		Nonce:        1,
		Orchestrator: keeper.OrchAddrs[0].String(),
		EthAddress:   ethAddress,
		Signature:    signature,
	}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err = h(ctx, msg)
	require.NoError(t, err)
	EndBlocker(ctx, k)
}
