package types

import (
	"crypto/md5" // nolint:gosec
	"encoding/binary"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// UInt64FromBytes create uint from binary big endian representation
func UInt64FromBytes(s []byte) (uint64, error) {
	if len(s) > 8 {
		return 0, fmt.Errorf("Invalid uint64 bytes passed to UInt64FromBytes!")
	}
	return binary.BigEndian.Uint64(s), nil
}

// UInt64Bytes uses the SDK byte marshaling to encode a uint64
func UInt64Bytes(n uint64) []byte {
	return sdk.Uint64ToBigEndian(n)
}

// UInt64FromString to parse out a uint64 for a nonce
func UInt64FromString(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

// IBCAddressFromBech32 decodes an IBC-compatible Address from a Bech32
// encoded string
func IBCAddressFromBech32(bech32str string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, ErrEmpty
	}

	_, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	err = sdk.VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// Hashing string using cryptographic MD5 function
// returns 128bit(16byte) value
func HashString(input string) []byte {
	md5 := md5.New() // nolint:gosec
	md5.Write([]byte(input))
	return md5.Sum(nil)
}

func AppendBytes(args ...[]byte) []byte {
	length := 0
	for _, v := range args {
		length += len(v)
	}

	res := make([]byte, length)

	length = 0
	for _, v := range args {
		copy(res[length:length+len(v)], v)
		length += len(v)
	}

	return res
}
