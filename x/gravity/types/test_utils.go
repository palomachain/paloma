package types

import (
	"encoding/hex"
	"math/big"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Creates a random nonzero uint64 test value
func NonzeroUint64() (ret uint64) {
	for ret == 0 {
		ret = rand.Uint64()
	}
	return
}

// Creates a random nonempty 20 byte sdk.AccAddress string test value
func NonemptySdkAccAddress() (ret sdk.AccAddress) {
	for ret.Empty() {
		addr := make([]byte, 20)
		rand.Read(addr)
		ret = sdk.AccAddress(addr)
	}
	return
}

// Creates a random nonempty 20 byte address hex string test value
func NonemptyEthAddress() (ret string) {
	for ret == "" {
		addr := make([]byte, 20)
		rand.Read(addr)
		ret = hex.EncodeToString(addr)
	}
	ret = "0x" + ret
	return
}

// Creates a random nonzero sdk.Int test value
func NonzeroSdkInt() (ret sdk.Int) {
	amount := big.NewInt(0)
	for amount.Cmp(big.NewInt(0)) == 0 {
		amountBz := make([]byte, 32)
		rand.Read(amountBz)
		amount = big.NewInt(0).SetBytes(amountBz)
	}
	ret = sdk.NewIntFromBigInt(amount)
	return
}
