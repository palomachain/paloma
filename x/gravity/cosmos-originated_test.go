package gravity

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palomachain/paloma/x/gravity/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Have the validators put in a erc20<>denom relation with ERC20DeployedEvent
// Send some coins of that denom into the cosmos module
// Check that the coins are locked, not burned
// Have the validators put in a deposit event for that ERC20
// Check that the coins are unlocked and sent to the right account

func TestCosmosOriginated(t *testing.T) {
	tv := initializeTestingVars(t)
	defer func() {
		tv.input.Context.Logger().Info("Asserting invariants at test end")
		tv.input.AssertInvariants()
	}()
	addDenomToERC20Relation(tv)
	// we only create a relation here, we don't perform
	// the other tests with the IBC representation as the
	// results should be the same
	addIbcDenomToERC20Relation(tv)
	lockCoinsInModule(tv)
	acceptDepositEvent(tv)
}

type testingVars struct {
	erc20 string
	denom string
	input keeper.TestInput
	ctx   sdk.Context
	h     sdk.Handler
	t     *testing.T
}

func initializeTestingVars(t *testing.T) *testingVars {
	var tv testingVars

	tv.t = t

	tv.erc20 = "0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e"
	tv.denom = "ugraviton"

	tv.input, tv.ctx = keeper.SetupFiveValChain(t)
	tv.h = NewHandler(tv.input.GravityKeeper)

	return &tv
}

func addDenomToERC20Relation(tv *testingVars) {
	tv.input.BankKeeper.SetDenomMetaData(tv.ctx, banktypes.Metadata{
		Description: "The native staking token of the Cosmos Gravity Bridge",
		Name:        "Graviton",
		Symbol:      "GRAV",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: "ugraviton", Exponent: uint32(0), Aliases: []string{"micrograviton"}},
			{Denom: "mgraviton", Exponent: uint32(3), Aliases: []string{"milligraviton"}},
			{Denom: "graviton", Exponent: uint32(6), Aliases: []string{}},
		},
		Base:    "ugraviton",
		Display: "graviton",
	})

	var (
		myNonce = uint64(1)
	)

	// have all five validators observe this event
	for _, v := range keeper.OrchAddrs {
		ethClaim := types.MsgERC20DeployedClaim{
			EventNonce:     myNonce,
			EthBlockHeight: 0,
			CosmosDenom:    tv.denom,
			TokenContract:  tv.erc20,
			Name:           "Graviton",
			Symbol:         "GRAV",
			Decimals:       6,
			Orchestrator:   v.String(),
		}
		_, err := tv.h(tv.ctx, &ethClaim)
		require.NoError(tv.t, err)

		// check if attestations persisted
		hash, err := ethClaim.ClaimHash()
		require.NoError(tv.t, err)
		a := tv.input.GravityKeeper.GetAttestation(tv.ctx, myNonce, hash)
		require.NotNil(tv.t, a)
	}

	EndBlocker(tv.ctx, tv.input.GravityKeeper)

	// check if erc20<>denom relation added to db
	isCosmosOriginated, gotERC20, err := tv.input.GravityKeeper.DenomToERC20Lookup(tv.ctx, tv.denom)
	require.NoError(tv.t, err)
	assert.True(tv.t, isCosmosOriginated)

	ethAddr, err := types.NewEthAddress(tv.erc20)
	require.NoError(tv.t, err)
	isCosmosOriginated, gotDenom := tv.input.GravityKeeper.ERC20ToDenomLookup(tv.ctx, *ethAddr)
	assert.True(tv.t, isCosmosOriginated)

	assert.Equal(tv.t, tv.denom, gotDenom)
	assert.Equal(tv.t, tv.erc20, gotERC20.GetAddress().Hex())
}

func lockCoinsInModule(tv *testingVars) {
	var (
		userCosmosAddr, err           = sdk.AccAddressFromBech32("gravity1990z7dqsvh8gthw9pa5sn4wuy2xrsd80lcx6lv")
		denom                         = "ugraviton"
		startingCoinAmount  sdk.Int   = sdk.NewIntFromUint64(150)
		sendAmount          sdk.Int   = sdk.NewIntFromUint64(50)
		feeAmount           sdk.Int   = sdk.NewIntFromUint64(5)
		startingCoins       sdk.Coins = sdk.Coins{sdk.NewCoin(denom, startingCoinAmount)}
		sendingCoin         sdk.Coin  = sdk.NewCoin(denom, sendAmount)
		feeCoin             sdk.Coin  = sdk.NewCoin(denom, feeAmount)
		ethDestination                = "0x3c9289da00b02dC623d0D8D907619890301D26d4"
	)
	assert.Nil(tv.t, err)

	// we start by depositing some funds into the users balance to send
	require.NoError(tv.t, tv.input.BankKeeper.MintCoins(tv.ctx, types.ModuleName, startingCoins))
	err = tv.input.BankKeeper.SendCoinsFromModuleToAccount(tv.ctx, types.ModuleName, userCosmosAddr, startingCoins)
	require.NoError(tv.t, err)
	balance1 := tv.input.BankKeeper.GetAllBalances(tv.ctx, userCosmosAddr)
	assert.Equal(tv.t, sdk.Coins{sdk.NewCoin(denom, startingCoinAmount)}, balance1)

	// send some coins
	// nolint: exhaustruct
	zeroCoin := sdk.Coin{}
	msg := &types.MsgSendToEth{
		Sender:    userCosmosAddr.String(),
		EthDest:   ethDestination,
		Amount:    sendingCoin,
		BridgeFee: feeCoin,
		ChainFee:  zeroCoin,
	}

	_, err = tv.h(tv.ctx, msg)
	require.NoError(tv.t, err)

	// Check that user balance has gone down
	balance2 := tv.input.BankKeeper.GetAllBalances(tv.ctx, userCosmosAddr)
	assert.Equal(tv.t, sdk.Coins{sdk.NewCoin(denom, startingCoinAmount.Sub(sendAmount).Sub(feeAmount))}, balance2)

	// Check that gravity balance has gone up
	gravityAddr := tv.input.AccountKeeper.GetModuleAddress(types.ModuleName)
	assert.Equal(tv.t,
		sdk.Coins{sdk.NewCoin(denom, sendAmount.Add(feeAmount))},
		tv.input.BankKeeper.GetAllBalances(tv.ctx, gravityAddr),
	)
}

func acceptDepositEvent(tv *testingVars) {
	var (
		myCosmosAddr, err = sdk.AccAddressFromBech32("gravity16ahjkfqxpp6lvfy9fpfnfjg39xr96qet0l08hu")
		myNonce           = uint64(3)
		anyETHAddr        = "0xf9613b532673Cc223aBa451dFA8539B87e1F666D"
	)
	require.NoError(tv.t, err)

	myErc20 := types.ERC20Token{
		Amount:   sdk.NewInt(12),
		Contract: tv.erc20,
	}

	// have all five validators observe this event
	for _, v := range keeper.OrchAddrs {
		ethClaim := types.MsgSendToCosmosClaim{
			EventNonce:     myNonce,
			EthBlockHeight: 0,
			TokenContract:  myErc20.Contract,
			Amount:         myErc20.Amount,
			EthereumSender: anyETHAddr,
			CosmosReceiver: myCosmosAddr.String(),
			Orchestrator:   v.String(),
		}

		_, err := tv.h(tv.ctx, &ethClaim)
		require.NoError(tv.t, err)
		EndBlocker(tv.ctx, tv.input.GravityKeeper)

		// check that attestation persisted
		hash, err := ethClaim.ClaimHash()
		require.NoError(tv.t, err)
		a := tv.input.GravityKeeper.GetAttestation(tv.ctx, myNonce, hash)
		require.NotNil(tv.t, a)
	}

	// Check that user balance has gone up
	assert.Equal(tv.t,
		sdk.Coins{sdk.NewCoin(tv.denom, myErc20.Amount)},
		tv.input.BankKeeper.GetAllBalances(tv.ctx, myCosmosAddr))

	// Check that gravity balance has gone down
	gravityAddr := tv.input.AccountKeeper.GetModuleAddress(types.ModuleName)
	assert.Equal(tv.t,
		sdk.Coins{sdk.NewCoin(tv.denom, sdk.NewIntFromUint64(55).Sub(myErc20.Amount))},
		tv.input.BankKeeper.GetAllBalances(tv.ctx, gravityAddr),
	)
}

func addIbcDenomToERC20Relation(tv *testingVars) {

	tokenContract := "0xE486cC1a00aA806C3e40224EDAd5FdCA93dDdA62"
	ibcDenom := "ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED/grav"
	metadata := banktypes.Metadata{
		Description: "Atom",
		Name:        "Atom",
		Base:        ibcDenom,
		Display:     "Atom",
		Symbol:      "ATOM",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    ibcDenom,
				Exponent: 0,
			},
			{
				Denom:    "Atom",
				Exponent: 6,
			},
		},
	}
	tv.input.BankKeeper.SetDenomMetaData(tv.ctx, metadata)

	var (
		myNonce = uint64(2)
	)

	// have all five validators observe this event
	for _, v := range keeper.OrchAddrs {
		ethClaim := types.MsgERC20DeployedClaim{
			EventNonce:     myNonce,
			EthBlockHeight: 0,
			CosmosDenom:    ibcDenom,
			TokenContract:  tokenContract,
			Name:           "Atom",
			Symbol:         "ATOM",
			Decimals:       6,
			Orchestrator:   v.String(),
		}
		_, err := tv.h(tv.ctx, &ethClaim)
		require.NoError(tv.t, err)

		// check if attestations persisted
		hash, err := ethClaim.ClaimHash()
		require.NoError(tv.t, err)
		a := tv.input.GravityKeeper.GetAttestation(tv.ctx, myNonce, hash)
		require.NotNil(tv.t, a)
	}

	EndBlocker(tv.ctx, tv.input.GravityKeeper)

	// check if erc20<>denom relation added to db
	isCosmosOriginated, gotERC20, err := tv.input.GravityKeeper.DenomToERC20Lookup(tv.ctx, tv.denom)
	require.NoError(tv.t, err)
	assert.True(tv.t, isCosmosOriginated)

	ethAddr, err := types.NewEthAddress(tv.erc20)
	require.NoError(tv.t, err)
	isCosmosOriginated, gotDenom := tv.input.GravityKeeper.ERC20ToDenomLookup(tv.ctx, *ethAddr)
	assert.True(tv.t, isCosmosOriginated)

	assert.Equal(tv.t, tv.denom, gotDenom)
	assert.Equal(tv.t, tv.erc20, gotERC20.GetAddress().Hex())
}
