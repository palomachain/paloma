package rand

import (
	"crypto/rand"
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

func Bytes(length uint) []byte {
	if length == 0 {
		panic("length can't be zero")
	}
	res := make([]byte, length)
	n, err := rand.Read(res)
	if err != nil {
		panic(err)
	}
	if uint(n) != length {
		panic("didn't read all bytes")
	}
	return res
}

func ValAddress() sdk.ValAddress {
	return Bytes(32)
}

func ETHAddress() common.Address {
	return common.BytesToAddress(Bytes(20))
}

func CryptoRandReader() io.Reader {
	return rand.Reader
}
