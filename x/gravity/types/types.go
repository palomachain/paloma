package types

import (
	"crypto/md5" // nolint:gosec
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bech32ibckeeper "github.com/palomachain/paloma/x/bech32ibc/keeper"
)

// UInt64FromBytesUnsafe create uint from binary big endian representation
// Note: This is unsafe because the function will panic if provided over 8 bytes
func UInt64FromBytesUnsafe(s []byte) uint64 {
	if len(s) > 8 {
		panic("Invalid uint64 bytes passed to UInt64FromBytes!")
	}
	return binary.BigEndian.Uint64(s)
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

// GetPrefixFromBech32 returns the human readable part of a bech32 string (excluding the 1 byte)
// Returns an error on too short input or when the 1 byte cannot be found
// Note: This is an excerpt from the Decode function for bech32 strings
func GetPrefixFromBech32(bech32str string) (string, error) {
	if len(bech32str) < 8 {
		return "", fmt.Errorf("invalid bech32 string length %d",
			len(bech32str))
	}
	one := strings.LastIndexByte(bech32str, '1')
	if one < 1 || one+7 > len(bech32str) {
		return "", fmt.Errorf("invalid index of 1")
	}

	return bech32str[:one], nil
}

// GetNativePrefixedAccAddressString treats the input as an AccAddress and re-prefixes the string
// with this chain's configured Bech32AccountAddrPrefix
// Returns an error when input is not a bech32 string
func GetNativePrefixedAccAddressString(ctx sdk.Context, bech32IbcKeeper bech32ibckeeper.Keeper, foreignStr string) (string, error) {
	prefix, err := GetPrefixFromBech32(foreignStr)
	if err != nil {
		return "", sdkerrors.Wrap(err, "invalid bech32 string")
	}
	nativePrefix, err := bech32IbcKeeper.GetNativeHrp(ctx)
	if err != nil {
		panic(sdkerrors.Wrap(err, "bech32ibc NativePrefix has not been registered!"))
	}
	if prefix == nativePrefix {
		return foreignStr, nil
	}

	return nativePrefix + foreignStr[len(prefix):], nil
}

// GetNativePrefixedAccAddress re-prefixes the input AccAddress with the registered bech32ibc NativeHrp
func GetNativePrefixedAccAddress(ctx sdk.Context, bech32IbcKeeper bech32ibckeeper.Keeper, foreignAddr sdk.AccAddress) (sdk.AccAddress, error) {
	nativeStr, err := GetNativePrefixedAccAddressString(ctx, bech32IbcKeeper, foreignAddr.String())
	if err != nil {
		return nil, err
	}
	return sdk.AccAddressFromBech32(nativeStr)
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
