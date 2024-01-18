package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

// ignore
func TestPrefixKeysSameLength(t *testing.T) {
	allKeys := getAllKeys(t)
	prefixKeys := allKeys[0:19]
	length := len(HashString("All keys should be same length when hashed"))

	for _, key := range prefixKeys {
		require.Equal(t, length, len(key), "key %v does not have the correct length %d", key, length)
	}
}

func TestNoDuplicateKeys(t *testing.T) {
	keys := getAllKeys(t)

	for i, key := range keys {
		keys[i] = nil
		require.NotContains(t, keys, key, "key %v should not be in keys!", key)
	}
}

func getAllKeys(t *testing.T) [][]byte {
	i := 0
	inc := func(i *int) *int { *i += 1; return i }

	keys := make([][]byte, 33)

	keys[i] = EthAddressByValidatorKey
	keys[*inc(&i)] = ValidatorByEthAddressKey
	keys[*inc(&i)] = LEGACYOracleClaimKey
	keys[*inc(&i)] = OracleAttestationKey
	keys[*inc(&i)] = OutgoingTXPoolKey
	keys[*inc(&i)] = OutgoingTXBatchKey
	keys[*inc(&i)] = BatchConfirmKey
	keys[*inc(&i)] = LastEventNonceByValidatorKey
	keys[*inc(&i)] = LastObservedEventNonceKey
	keys[*inc(&i)] = LEGACYSequenceKeyPrefix
	keys[*inc(&i)] = KeyLastTXPoolID
	keys[*inc(&i)] = KeyLastOutgoingBatchID
	keys[*inc(&i)] = KeyOrchestratorAddress
	keys[*inc(&i)] = LastObservedEthereumBlockHeightKey
	keys[*inc(&i)] = DenomToERC20Key
	keys[*inc(&i)] = ERC20ToDenomKey
	keys[*inc(&i)] = LastSlashedBatchBlock
	keys[*inc(&i)] = LastUnBondingBlockHeight
	keys[*inc(&i)] = PastEthSignatureCheckpointKey

	// sdk.AccAddress, sdk.ValAddress
	dummyAddr := []byte("paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk")
	// EthAddress
	ethAddr, err := NewEthAddress("0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B")
	require.NoError(t, err)
	dummyEthAddr := *ethAddr
	// Nonce
	dummyNonce := uint64(1)
	// Claim, InvalidationId
	dummyBytes := []byte("0xc783df8a850f42e7F7e57013759C285caa701eB6")
	// InternationalERC20Token
	dummyErc := InternalERC20Token{Amount: math.OneInt(), Contract: dummyEthAddr}

	// Denom
	dummyChainReferenceID := "test-chain"
	dummyDenom := "footoken"

	orchestratorAddressKey, err := GetOrchestratorAddressKey(dummyAddr)
	require.NoError(t, err)
	keys[*inc(&i)] = orchestratorAddressKey

	ethAddressByValidatorKey, err := GetEthAddressByValidatorKey(dummyAddr)
	require.NoError(t, err)
	keys[*inc(&i)] = ethAddressByValidatorKey

	keys[*inc(&i)] = GetValidatorByEthAddressKey(dummyEthAddr)
	keys[*inc(&i)] = GetAttestationKey(dummyNonce, dummyBytes)
	keys[*inc(&i)] = GetOutgoingTxPoolContractPrefix(dummyEthAddr)
	keys[*inc(&i)] = GetOutgoingTxPoolKey(dummyErc, dummyNonce)
	keys[*inc(&i)] = GetOutgoingTxBatchContractPrefix(dummyEthAddr)
	keys[*inc(&i)] = GetOutgoingTxBatchKey(dummyEthAddr, dummyNonce)
	keys[*inc(&i)] = GetBatchConfirmNonceContractPrefix(dummyEthAddr, dummyNonce)

	batchConfirmKey, err := GetBatchConfirmKey(dummyEthAddr, dummyNonce, dummyAddr)
	keys[*inc(&i)] = batchConfirmKey

	lastEventNonceByValidatorKey, err := GetLastEventNonceByValidatorKey(dummyAddr)
	keys[*inc(&i)] = lastEventNonceByValidatorKey

	keys[*inc(&i)] = GetDenomToERC20Key(dummyChainReferenceID, dummyDenom)
	keys[*inc(&i)] = GetERC20ToDenomKey(dummyChainReferenceID, dummyEthAddr)
	keys[*inc(&i)] = GetPastEthSignatureCheckpointKey(dummyBytes)

	return keys
}
