package keeper

import (
	"math/big"
	"testing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestDetectMaliciousSupply(t *testing.T) {
	input := CreateTestEnv(t)

	// set supply to maximum value
	var testBigInt big.Int
	testBigInt.SetBit(new(big.Int), 256, 1).Sub(&testBigInt, big.NewInt(1))
	bigCoinAmount := sdktypes.NewIntFromBigInt(&testBigInt)

	err := input.GravityKeeper.DetectMaliciousSupply(input.Context, "ugrain", bigCoinAmount)
	require.Error(t, err, "didn't error out on too much added supply")
}
