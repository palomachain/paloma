package types

import (
	"bytes"
	"encoding/hex"
	mrand "math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValsetConfirmHash(t *testing.T) {
	powers := []uint64{3333, 3333, 3333}
	ethAddresses := []string{
		"0xc783df8a850f42e7F7e57013759C285caa701eB6",
		"0xeAD9C93b79Ae7C1591b1FB5323BD777E86e150d4",
		"0xE5904695748fe4A84b40b3fc79De2277660BD1D3",
	}
	members := make(InternalBridgeValidators, len(powers))
	for i := range powers {
		bv := BridgeValidator{
			Power:           powers[i],
			EthereumAddress: ethAddresses[i],
		}
		ibv, err := NewInternalBridgeValidator(bv)
		require.NoError(t, err)
		members[i] = ibv
	}

	v, err := NewValset(0, 0, members, sdk.NewInt(0), ZeroAddress())
	require.NoError(t, err)

	// normally we would load the GravityID from the store, but for this test we use
	// the same hardcoded value in the solidity tests
	hash := v.GetCheckpoint("foo")
	hexHash := hex.EncodeToString(hash)
	correctHash := "0xaca2f283f21a03ba182dc7d34a55c04771b25087401d680011df7dcba453f798"[2:]
	assert.Equal(t, correctHash, hexHash)
}

func TestValsetCheckpointGold1(t *testing.T) {
	bridgeValidators, err := BridgeValidators{{
		Power:           6667,
		EthereumAddress: "0xc783df8a850f42e7F7e57013759C285caa701eB6",
	}}.ToInternal()
	require.NoError(t, err)
	src, err := NewValset(0, 0, *bridgeValidators, sdk.NewInt(0), ZeroAddress())
	require.NoError(t, err)

	// normally we would load the GravityID from the store, but for this test we use
	// the same hardcoded value in the solidity tests
	ourHash := src.GetCheckpoint("foo")

	// hash from bridge contract
	goldHash := "0x89731c26bab12cf0cb5363ef9abab6f9bd5496cf758a2309311c7946d54bca85"[2:]
	assert.Equal(t, goldHash, hex.EncodeToString(ourHash))
}

func TestValsetPowerDiff(t *testing.T) {
	specs := map[string]struct {
		start BridgeValidators
		diff  BridgeValidators
		exp   float64
	}{
		"no diff": {
			start: BridgeValidators{
				{Power: 1, EthereumAddress: "0x479FFc856Cdfa0f5D1AE6Fa61915b01351A7773D"},
				{Power: 2, EthereumAddress: "0x8E91960d704Df3fF24ECAb78AB9df1B5D9144140"},
				{Power: 3, EthereumAddress: "0xF14879a175A2F1cEFC7c616f35b6d9c2b0Fd8326"},
			},
			diff: BridgeValidators{
				{Power: 1, EthereumAddress: "0x479FFc856Cdfa0f5D1AE6Fa61915b01351A7773D"},
				{Power: 2, EthereumAddress: "0x8E91960d704Df3fF24ECAb78AB9df1B5D9144140"},
				{Power: 3, EthereumAddress: "0xF14879a175A2F1cEFC7c616f35b6d9c2b0Fd8326"},
			},
			exp: 0.0,
		},
		"one": {
			start: BridgeValidators{
				{Power: 1073741823, EthereumAddress: "0x479FFc856Cdfa0f5D1AE6Fa61915b01351A7773D"},
				{Power: 1073741823, EthereumAddress: "0x8E91960d704Df3fF24ECAb78AB9df1B5D9144140"},
				{Power: 2147483646, EthereumAddress: "0xF14879a175A2F1cEFC7c616f35b6d9c2b0Fd8326"},
			},
			diff: BridgeValidators{
				{Power: 858993459, EthereumAddress: "0x479FFc856Cdfa0f5D1AE6Fa61915b01351A7773D"},
				{Power: 858993459, EthereumAddress: "0x8E91960d704Df3fF24ECAb78AB9df1B5D9144140"},
				{Power: 2576980377, EthereumAddress: "0xF14879a175A2F1cEFC7c616f35b6d9c2b0Fd8326"},
			},
			exp: 0.2,
		},
		"real world": {
			start: BridgeValidators{
				{Power: 678509841, EthereumAddress: "0x6db48cBBCeD754bDc760720e38E456144e83269b"},
				{Power: 671724742, EthereumAddress: "0x8E91960d704Df3fF24ECAb78AB9df1B5D9144140"},
				{Power: 685294939, EthereumAddress: "0x479FFc856Cdfa0f5D1AE6Fa61915b01351A7773D"},
				{Power: 671724742, EthereumAddress: "0x0A7254b318dd742A3086882321C27779B4B642a6"},
				{Power: 671724742, EthereumAddress: "0x454330deAaB759468065d08F2b3B0562caBe1dD1"},
				{Power: 617443955, EthereumAddress: "0x3511A211A6759d48d107898302042d1301187BA9"},
				{Power: 6785098, EthereumAddress: "0x37A0603dA2ff6377E5C7f75698dabA8EE4Ba97B8"},
				{Power: 291759231, EthereumAddress: "0xF14879a175A2F1cEFC7c616f35b6d9c2b0Fd8326"},
			},
			diff: BridgeValidators{
				{Power: 642345266, EthereumAddress: "0x479FFc856Cdfa0f5D1AE6Fa61915b01351A7773D"},
				{Power: 678509841, EthereumAddress: "0x6db48cBBCeD754bDc760720e38E456144e83269b"},
				{Power: 671724742, EthereumAddress: "0x0A7254b318dd742A3086882321C27779B4B642a6"},
				{Power: 671724742, EthereumAddress: "0x454330deAaB759468065d08F2b3B0562caBe1dD1"},
				{Power: 671724742, EthereumAddress: "0x8E91960d704Df3fF24ECAb78AB9df1B5D9144140"},
				{Power: 617443955, EthereumAddress: "0x3511A211A6759d48d107898302042d1301187BA9"},
				{Power: 291759231, EthereumAddress: "0xF14879a175A2F1cEFC7c616f35b6d9c2b0Fd8326"},
				{Power: 6785098, EthereumAddress: "0x37A0603dA2ff6377E5C7f75698dabA8EE4Ba97B8"},
			},
			exp: 0.010000000011641532,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			startInternal, err := spec.start.ToInternal()
			require.NoError(t, err)
			diffInternal, err := spec.diff.ToInternal()
			require.NoError(t, err)
			assert.Equal(t, spec.exp, startInternal.PowerDiff(*diffInternal))
		})
	}
}

func TestValsetSort(t *testing.T) {
	address1 := gethcommon.BytesToAddress(bytes.Repeat([]byte{byte(1)}, 20)).String()
	address2 := gethcommon.BytesToAddress(bytes.Repeat([]byte{byte(2)}, 20)).String()
	address3 := gethcommon.BytesToAddress(bytes.Repeat([]byte{byte(3)}, 20)).String()

	specs := map[string]struct {
		src BridgeValidators
		exp BridgeValidators
	}{

		"by power desc": {
			src: BridgeValidators{
				{Power: 1, EthereumAddress: address3},
				{Power: 2, EthereumAddress: address1},
				{Power: 3, EthereumAddress: address2},
			},
			exp: BridgeValidators{
				{Power: 3, EthereumAddress: address2},
				{Power: 2, EthereumAddress: address1},
				{Power: 1, EthereumAddress: address3},
			},
		},
		"by eth addr on same power": {
			src: BridgeValidators{
				{Power: 1, EthereumAddress: address2},
				{Power: 1, EthereumAddress: address1},
				{Power: 1, EthereumAddress: address3},
			},
			exp: BridgeValidators{
				{Power: 1, EthereumAddress: address1},
				{Power: 1, EthereumAddress: address2},
				{Power: 1, EthereumAddress: address3},
			},
		},
		// if you're thinking about changing this due to a change in the sorting algorithm
		// you MUST go change this in gravity_utils/types.rs as well. You will also break all
		// bridges in production when they try to migrate so use extreme caution!
		"real world": {
			src: BridgeValidators{
				{Power: 678509841, EthereumAddress: "0x6db48cBBCeD754bDc760720e38E456144e83269b"},
				{Power: 671724742, EthereumAddress: "0x8E91960d704Df3fF24ECAb78AB9df1B5D9144140"},
				{Power: 685294939, EthereumAddress: "0x479FFc856Cdfa0f5D1AE6Fa61915b01351A7773D"},
				{Power: 671724742, EthereumAddress: "0x0A7254b318dd742A3086882321C27779B4B642a6"},
				{Power: 671724742, EthereumAddress: "0x454330deAaB759468065d08F2b3B0562caBe1dD1"},
				{Power: 617443955, EthereumAddress: "0x3511A211A6759d48d107898302042d1301187BA9"},
				{Power: 6785098, EthereumAddress: "0x37A0603dA2ff6377E5C7f75698dabA8EE4Ba97B8"},
				{Power: 291759231, EthereumAddress: "0xF14879a175A2F1cEFC7c616f35b6d9c2b0Fd8326"},
			},
			exp: BridgeValidators{
				{Power: 685294939, EthereumAddress: "0x479FFc856Cdfa0f5D1AE6Fa61915b01351A7773D"},
				{Power: 678509841, EthereumAddress: "0x6db48cBBCeD754bDc760720e38E456144e83269b"},
				{Power: 671724742, EthereumAddress: "0x0A7254b318dd742A3086882321C27779B4B642a6"},
				{Power: 671724742, EthereumAddress: "0x454330deAaB759468065d08F2b3B0562caBe1dD1"},
				{Power: 671724742, EthereumAddress: "0x8E91960d704Df3fF24ECAb78AB9df1B5D9144140"},
				{Power: 617443955, EthereumAddress: "0x3511A211A6759d48d107898302042d1301187BA9"},
				{Power: 291759231, EthereumAddress: "0xF14879a175A2F1cEFC7c616f35b6d9c2b0Fd8326"},
				{Power: 6785098, EthereumAddress: "0x37A0603dA2ff6377E5C7f75698dabA8EE4Ba97B8"},
			},
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			srcInternal, err := spec.src.ToInternal()
			require.NoError(t, err)
			expInternal, err := spec.exp.ToInternal()
			require.NoError(t, err)
			srcInternal.Sort()
			assert.Equal(t, srcInternal, expInternal)
			shuffled := shuffled(*srcInternal)
			shuffled.Sort()
			assert.Equal(t, shuffled, *expInternal)
		})
	}
}

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

func shuffled(v InternalBridgeValidators) InternalBridgeValidators {
	mrand.Shuffle(len(v), func(i, j int) {
		v[i], v[j] = v[j], v[i]
	})
	return v
}
