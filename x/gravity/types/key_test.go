package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// ignore
func TestPrefixKeysSameLength(t *testing.T) {
	allKeys := getAllKeys()
	prefixKeys := allKeys[0:27]
	length := len(HashString("All keys should be same length when hashed"))

	for _, key := range prefixKeys {
		require.Equal(t, length, len(key), "key %v does not have the correct length %d", key, length)
	}
}

func TestNoDuplicateKeys(t *testing.T) {
	keys := getAllKeys()

	for i, key := range keys {
		keys[i] = nil
		require.NotContains(t, keys, key, "key %v should not be in keys!", key)
	}
}

func getAllKeys() [][]byte {
	i := 0
	inc := func(i *int) *int { *i += 1; return i }

	keys := make([][]byte, 47)

	keys[i] = EthAddressByValidatorKey
	keys[*inc(&i)] = ValidatorByEthAddressKey
	keys[*inc(&i)] = ValsetRequestKey
	keys[*inc(&i)] = ValsetConfirmKey
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
	keys[*inc(&i)] = KeyOutgoingLogicCall
	keys[*inc(&i)] = KeyOutgoingLogicConfirm
	keys[*inc(&i)] = LastObservedEthereumBlockHeightKey
	keys[*inc(&i)] = DenomToERC20Key
	keys[*inc(&i)] = ERC20ToDenomKey
	keys[*inc(&i)] = LastSlashedValsetNonce
	keys[*inc(&i)] = LatestValsetNonce
	keys[*inc(&i)] = LastSlashedBatchBlock
	keys[*inc(&i)] = LastSlashedLogicCallBlock
	keys[*inc(&i)] = LastUnBondingBlockHeight
	keys[*inc(&i)] = LastObservedValsetKey
	keys[*inc(&i)] = PastEthSignatureCheckpointKey

	// sdk.AccAddress, sdk.ValAddress
	dummyAddr := []byte("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
	// EthAddress
	ethAddr, err := NewEthAddress("0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B")
	if err != nil {
		panic(err)
	}
	dummyEthAddr := *ethAddr
	// Nonce
	dummyNonce := uint64(1)
	// Claim, InvalidationId
	dummyBytes := []byte("0xc783df8a850f42e7F7e57013759C285caa701eB6")
	// InternationalERC20Token
	dummyErc := InternalERC20Token{Amount: sdk.OneInt(), Contract: dummyEthAddr}
	// Denom
	dummyDenom := "footoken"

	keys[*inc(&i)] = GetOrchestratorAddressKey(dummyAddr)
	keys[*inc(&i)] = GetEthAddressByValidatorKey(dummyAddr)
	keys[*inc(&i)] = GetValidatorByEthAddressKey(dummyEthAddr)
	keys[*inc(&i)] = GetValsetKey(dummyNonce)
	keys[*inc(&i)] = GetValsetConfirmNoncePrefix(dummyNonce)
	keys[*inc(&i)] = GetValsetConfirmKey(dummyNonce, dummyAddr)
	keys[*inc(&i)] = GetAttestationKey(dummyNonce, dummyBytes)
	keys[*inc(&i)] = GetOutgoingTxPoolContractPrefix(dummyEthAddr)
	keys[*inc(&i)] = GetOutgoingTxPoolKey(dummyErc, dummyNonce)
	keys[*inc(&i)] = GetOutgoingTxBatchContractPrefix(dummyEthAddr)
	keys[*inc(&i)] = GetOutgoingTxBatchKey(dummyEthAddr, dummyNonce)
	keys[*inc(&i)] = GetBatchConfirmNonceContractPrefix(dummyEthAddr, dummyNonce)
	keys[*inc(&i)] = GetBatchConfirmKey(dummyEthAddr, dummyNonce, dummyAddr)
	keys[*inc(&i)] = GetLastEventNonceByValidatorKey(dummyAddr)
	keys[*inc(&i)] = GetDenomToERC20Key(dummyDenom)
	keys[*inc(&i)] = GetERC20ToDenomKey(dummyEthAddr)
	keys[*inc(&i)] = GetOutgoingLogicCallKey(dummyBytes, dummyNonce)
	keys[*inc(&i)] = GetLogicConfirmNonceInvalidationIdPrefix(dummyBytes, dummyNonce)
	keys[*inc(&i)] = GetLogicConfirmKey(dummyBytes, dummyNonce, dummyAddr)
	keys[*inc(&i)] = GetPastEthSignatureCheckpointKey(dummyBytes)

	return keys
}
