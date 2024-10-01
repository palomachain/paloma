package keeper

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/skyway/types"
	vtypes "github.com/palomachain/paloma/v2/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint: exhaustruct
func TestPrefixRange(t *testing.T) {
	cases := map[string]struct {
		src      []byte
		expStart []byte
		expEnd   []byte
		expError bool
	}{
		"normal":              {src: []byte{1, 3, 4}, expStart: []byte{1, 3, 4}, expEnd: []byte{1, 3, 5}},
		"normal short":        {src: []byte{79}, expStart: []byte{79}, expEnd: []byte{80}},
		"empty case":          {src: []byte{}},
		"roll-over example 1": {src: []byte{17, 28, 255}, expStart: []byte{17, 28, 255}, expEnd: []byte{17, 29, 0}},
		"roll-over example 2": {
			src:      []byte{15, 42, 255, 255},
			expStart: []byte{15, 42, 255, 255}, expEnd: []byte{15, 43, 0, 0},
		},
		"pathological roll-over": {src: []byte{255, 255, 255, 255}, expStart: []byte{255, 255, 255, 255}},
		"nil prohibited":         {expError: true},
	}

	for testName, tc := range cases {
		tc := tc
		t.Run(testName, func(t *testing.T) {
			start, end, err := prefixRange(tc.src)
			if tc.expError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expStart, start)
				assert.Equal(t, tc.expEnd, end)
			}
		})
	}
}

// nolint: exhaustruct
func TestAttestationIterator(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() {
		sdk.UnwrapSDKContext(input.Context).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	ctx := input.Context
	// add some attestations to the store

	claim1 := &types.MsgSendToPalomaClaim{
		SkywayNonce:    1,
		EventNonce:     1,
		TokenContract:  TokenContractAddrs[0],
		Amount:         math.NewInt(100),
		EthereumSender: EthAddrs[0].String(),
		PalomaReceiver: AccAddrs[0].String(),
		Orchestrator:   AccAddrs[0].String(),
		Metadata: vtypes.MsgMetadata{
			Creator: AccAddrs[0].String(),
			Signers: []string{AccAddrs[0].String()},
		},
	}
	ne1, err := codecTypes.NewAnyWithValue(claim1)
	require.NoError(t, err)
	att1 := &types.Attestation{
		Claim:    ne1,
		Observed: true,
		Votes:    []string{ValAddrs[0].String()},
	}

	claim2 := &types.MsgSendToPalomaClaim{
		SkywayNonce:    2,
		EventNonce:     5,
		TokenContract:  TokenContractAddrs[0],
		Amount:         math.NewInt(100),
		EthereumSender: EthAddrs[0].String(),
		PalomaReceiver: AccAddrs[0].String(),
		Orchestrator:   AccAddrs[0].String(),
		Metadata: vtypes.MsgMetadata{
			Creator: AccAddrs[0].String(),
			Signers: []string{AccAddrs[0].String()},
		},
	}
	ne2, err := codecTypes.NewAnyWithValue(claim2)
	require.NoError(t, err)
	att2 := &types.Attestation{
		Claim:    ne2,
		Observed: true,
		Votes:    []string{ValAddrs[0].String()},
	}

	hash1, err := claim1.ClaimHash()
	require.NoError(t, err)
	hash2, err := claim2.ClaimHash()
	require.NoError(t, err)

	input.SkywayKeeper.SetAttestation(ctx, "test-chain", claim1.SkywayNonce, hash1, att1)
	err = input.SkywayKeeper.setLastObservedSkywayNonce(ctx, "test-chain", claim1.SkywayNonce)
	require.NoError(t, err)
	input.SkywayKeeper.SetAttestation(ctx, "test-chain", claim2.SkywayNonce, hash2, att2)
	err = input.SkywayKeeper.setLastObservedSkywayNonce(ctx, "test-chain", claim2.SkywayNonce)
	require.NoError(t, err)

	atts := []types.Attestation{}
	err = input.SkywayKeeper.IterateAttestations(ctx, "test-chain", false, func(_ []byte, att types.Attestation) bool {
		atts = append(atts, att)
		return false
	})
	require.NoError(t, err)

	require.Len(t, atts, 2)
}

func TestSetBridgeTax(t *testing.T) {
	input := CreateTestEnv(t)
	ctx := input.Context
	k := input.SkywayKeeper

	accAddresses := []string{
		"paloma1dg55rtevlfxh46w88yjpdd08sqhh5cc37jmmth",
		"paloma164knshrzuuurf05qxf3q5ewpfnwzl4gjd7cwmp",
	}

	addresses := make([]sdk.AccAddress, len(accAddresses))
	for i := range accAddresses {
		addr, err := sdk.AccAddressFromBech32(accAddresses[i])
		require.NoError(t, err)

		addresses[i] = addr
	}

	t.Run("Return new bridge tax after setting it", func(t *testing.T) {
		expected := types.BridgeTax{
			Token:           "test",
			Rate:            "0.02",
			ExemptAddresses: addresses,
		}

		err := k.SetBridgeTax(ctx, &expected)
		require.NoError(t, err)

		actual, err := k.BridgeTax(ctx, "test")
		require.NoError(t, err)
		require.Equal(t, *actual, expected)

		all, err := k.AllBridgeTaxes(ctx)
		require.NoError(t, err)
		require.Equal(t, all, []*types.BridgeTax{&expected})
	})

	t.Run("Return error when trying to set a negative tax", func(t *testing.T) {
		err := k.SetBridgeTax(ctx, &types.BridgeTax{
			Token: "test",
			Rate:  "-0.2",
		})
		require.Error(t, err)
	})

	t.Run("Return OK with a tax higher than 100%", func(t *testing.T) {
		err := k.SetBridgeTax(ctx, &types.BridgeTax{
			Token: "test",
			Rate:  "1.1",
		})
		require.NoError(t, err)
	})
}

func TestBridgeTaxAmount(t *testing.T) {
	var ctx context.Context
	var k Keeper
	setup := func() {
		input := CreateTestEnv(t)
		ctx = input.Context
		k = input.SkywayKeeper
	}

	accAddresses := []string{
		"paloma1dg55rtevlfxh46w88yjpdd08sqhh5cc37jmmth",
		"paloma164knshrzuuurf05qxf3q5ewpfnwzl4gjd7cwmp",
	}

	addresses := make([]sdk.AccAddress, len(accAddresses))
	for i := range accAddresses {
		addr, err := sdk.AccAddressFromBech32(accAddresses[i])
		require.NoError(t, err)

		addresses[i] = addr
	}

	coin := sdk.Coin{
		Denom:  "test",
		Amount: math.NewInt(1_000_000),
	}

	t.Run("Return zero when bridge tax is not set yet", func(t *testing.T) {
		setup()

		actual, err := k.bridgeTaxAmount(ctx, addresses[0], coin)
		require.NoError(t, err)
		require.Equal(t, actual, math.ZeroInt())
	})

	t.Run("Return zero when bridge tax is set to zero", func(t *testing.T) {
		setup()

		err := k.SetBridgeTax(ctx, &types.BridgeTax{
			Token: coin.Denom,
			Rate:  "0",
		})
		require.NoError(t, err)

		actual, err := k.bridgeTaxAmount(ctx, addresses[0], coin)
		require.NoError(t, err)
		require.Equal(t, actual, math.ZeroInt())
	})

	t.Run("Return all when bridge tax is set to one", func(t *testing.T) {
		setup()

		err := k.SetBridgeTax(ctx, &types.BridgeTax{
			Token: coin.Denom,
			Rate:  "1",
		})
		require.NoError(t, err)

		actual, err := k.bridgeTaxAmount(ctx, addresses[0], coin)
		require.NoError(t, err)
		require.Equal(t, actual, coin.Amount)
	})

	t.Run("Return zero when token does not have a tax set", func(t *testing.T) {
		setup()

		err := k.SetBridgeTax(ctx, &types.BridgeTax{
			Token: "anothertoken",
			Rate:  "0.01",
		})
		require.NoError(t, err)

		actual, err := k.bridgeTaxAmount(ctx, addresses[0], coin)
		require.NoError(t, err)
		require.Equal(t, actual, math.ZeroInt())
	})

	t.Run("Return zero when sender is exempt", func(t *testing.T) {
		setup()

		err := k.SetBridgeTax(ctx, &types.BridgeTax{
			Rate:            "0.01",
			Token:           coin.Denom,
			ExemptAddresses: addresses,
		})
		require.NoError(t, err)

		actual, err := k.bridgeTaxAmount(ctx, addresses[0], coin)
		require.NoError(t, err)
		require.Equal(t, actual, math.ZeroInt())
	})

	t.Run("Return the correct tax amount", func(t *testing.T) {
		setup()

		err := k.SetBridgeTax(ctx, &types.BridgeTax{
			Token: coin.Denom,
			Rate:  "0.00123",
		})
		require.NoError(t, err)

		actual, err := k.bridgeTaxAmount(ctx, addresses[0], coin)
		require.NoError(t, err)
		require.Equal(t, actual, math.NewInt(1230))
	})

	t.Run("Return the correct tax amount on huge transfer", func(t *testing.T) {
		setup()

		amount, _ := math.NewIntFromString("1500000000000000000000")
		hugeCoin := sdk.Coin{
			Denom:  "test",
			Amount: amount,
		}

		err := k.SetBridgeTax(ctx, &types.BridgeTax{
			Token: coin.Denom,
			Rate:  "2",
		})
		require.NoError(t, err)

		actual, err := k.bridgeTaxAmount(ctx, addresses[0], hugeCoin)
		require.NoError(t, err)
		expected, _ := math.NewIntFromString("3000000000000000000000")
		require.Equal(t, actual, expected)
	})
}

func TestBridgeTransferLimits(t *testing.T) {
	var ctx context.Context
	var k Keeper

	sender, err := sdk.AccAddressFromBech32("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
	require.NoError(t, err)

	accAddresses := []string{
		"paloma1dg55rtevlfxh46w88yjpdd08sqhh5cc37jmmth",
		"paloma164knshrzuuurf05qxf3q5ewpfnwzl4gjd7cwmp",
	}

	addresses := make([]sdk.AccAddress, len(accAddresses))
	for i := range accAddresses {
		addr, err := sdk.AccAddressFromBech32(accAddresses[i])
		require.NoError(t, err)

		addresses[i] = addr
	}

	coin := sdk.Coin{
		Denom:  "test",
		Amount: math.NewInt(1_000_000),
	}

	transferLimit := &types.BridgeTransferLimit{
		Token:           "test",
		Limit:           math.NewInt(1000),
		LimitPeriod:     types.LimitPeriod_DAILY,
		ExemptAddresses: addresses,
	}

	setup := func(setLimit bool) {
		input := CreateTestEnv(t)
		ctx = input.Context
		k = input.SkywayKeeper

		if setLimit {
			err := k.SetBridgeTransferLimit(ctx, transferLimit)
			require.NoError(t, err)
		}
	}

	t.Run("Not return error when no limit is set", func(t *testing.T) {
		setup(false)

		err = k.UpdateBridgeTransferUsageWithLimit(ctx, sender, coin)
		require.NoError(t, err)
	})

	t.Run("Not return error when limit is set to NONE", func(t *testing.T) {
		setup(false)

		err := k.SetBridgeTransferLimit(ctx, &types.BridgeTransferLimit{
			Token:       "test",
			Limit:       math.NewInt(1000),
			LimitPeriod: types.LimitPeriod_NONE,
		})
		require.NoError(t, err)

		err = k.UpdateBridgeTransferUsageWithLimit(ctx, sender, coin)
		require.NoError(t, err)
	})

	t.Run("Not return error when sender is exempt", func(t *testing.T) {
		setup(true)

		err = k.UpdateBridgeTransferUsageWithLimit(ctx, addresses[0], coin)
		require.NoError(t, err)
	})

	t.Run("Return error when transfer is above limit", func(t *testing.T) {
		setup(true)

		err = k.UpdateBridgeTransferUsageWithLimit(ctx, sender, coin)
		require.Error(t, err)
	})

	t.Run("Update usage on successful transfer", func(t *testing.T) {
		setup(true)

		coin := sdk.Coin{
			Denom:  "test",
			Amount: math.NewInt(600),
		}

		err = k.UpdateBridgeTransferUsageWithLimit(ctx, sender, coin)
		require.NoError(t, err)

		actual, err := k.BridgeTransferUsage(ctx, coin.Denom)
		require.NoError(t, err)

		expected := &types.BridgeTransferUsage{
			Total:            coin.Amount,
			StartBlockHeight: sdk.UnwrapSDKContext(ctx).BlockHeight(),
		}

		require.Equal(t, expected, actual)
	})

	t.Run("Update usage but keep block height on multiple transfers", func(t *testing.T) {
		setup(true)

		coin := sdk.Coin{
			Denom:  "test",
			Amount: math.NewInt(300),
		}

		// Send the first transfer
		err = k.UpdateBridgeTransferUsageWithLimit(ctx, sender, coin)
		require.NoError(t, err)

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		newCtx := sdkCtx.WithBlockHeight(sdkCtx.BlockHeight() + 1000)

		// Send a second transfer with a new block height
		err = k.UpdateBridgeTransferUsageWithLimit(newCtx, sender, coin)
		require.NoError(t, err)

		actual, err := k.BridgeTransferUsage(newCtx, coin.Denom)
		require.NoError(t, err)

		// Should get usage of sum of both transfers, but keep the first block
		// height
		expected := &types.BridgeTransferUsage{
			Total:            math.NewInt(600),
			StartBlockHeight: sdkCtx.BlockHeight(),
		}

		require.Equal(t, expected, actual)
	})

	t.Run("Update usage after block limit", func(t *testing.T) {
		setup(true)

		coin := sdk.Coin{
			Denom:  "test",
			Amount: math.NewInt(300),
		}

		// Send the first transfer
		err = k.UpdateBridgeTransferUsageWithLimit(ctx, sender, coin)
		require.NoError(t, err)

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		newCtx := sdkCtx.WithBlockHeight(sdkCtx.BlockHeight() + transferLimit.BlockLimit())

		newCoin := sdk.Coin{
			Denom:  "test",
			Amount: math.NewInt(450),
		}

		// Send a second transfer with a new block height
		err = k.UpdateBridgeTransferUsageWithLimit(newCtx, sender, newCoin)
		require.NoError(t, err)

		actual, err := k.BridgeTransferUsage(newCtx, coin.Denom)
		require.NoError(t, err)

		// Should get usage and block height of last transfer
		expected := &types.BridgeTransferUsage{
			Total:            newCoin.Amount,
			StartBlockHeight: newCtx.BlockHeight(),
		}

		require.Equal(t, expected, actual)
	})

	t.Run("Keep usage after passing limit", func(t *testing.T) {
		setup(true)

		coin := sdk.Coin{
			Denom:  "test",
			Amount: math.NewInt(300),
		}

		// Send the first transfer
		err = k.UpdateBridgeTransferUsageWithLimit(ctx, sender, coin)
		require.NoError(t, err)

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		newCtx := sdkCtx.WithBlockHeight(sdkCtx.BlockHeight() + 1000)

		newCoin := sdk.Coin{
			Denom:  "test",
			Amount: math.NewInt(1000),
		}

		// Send a second transfer with a new block height
		// Will not go through as we're over the limit
		err = k.UpdateBridgeTransferUsageWithLimit(newCtx, sender, newCoin)
		require.Error(t, err)

		actual, err := k.BridgeTransferUsage(newCtx, coin.Denom)
		require.NoError(t, err)

		// Should get usage and block height of first transfer
		expected := &types.BridgeTransferUsage{
			Total:            coin.Amount,
			StartBlockHeight: sdkCtx.BlockHeight(),
		}

		require.Equal(t, expected, actual)
	})
}
