package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppendBytes(t *testing.T) {
	// Prefix
	prefix := EthAddressByValidatorKey
	// EthAddress
	ethAddrBytes := []byte("0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B")
	// Nonce
	nonce := uint64(1)
	// Data
	bytes := []byte("0xc783df8a850f42e7F7e57013759C285caa701eB6")

	appended := AppendBytes(prefix, ethAddrBytes, UInt64Bytes(nonce), bytes)

	lenPrefix := len(prefix)
	lenEthAddr := len(ethAddrBytes)
	lenNonce := len(UInt64Bytes(nonce))

	// Appended bytes should be same length as sum of all lengths
	require.Equal(t, lenPrefix+lenEthAddr+lenNonce+len(bytes), len(appended))

	// Appended bytes should be in correct order and be same as source
	require.Equal(t, prefix, appended[:lenPrefix])
	require.Equal(t, ethAddrBytes, appended[lenPrefix:lenPrefix+lenEthAddr])
	require.Equal(t, UInt64Bytes(nonce), appended[lenPrefix+lenEthAddr:lenPrefix+lenEthAddr+lenNonce])
	require.Equal(t, bytes, appended[lenPrefix+lenEthAddr+lenNonce:])
}
