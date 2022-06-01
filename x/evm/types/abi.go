package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vizualni/whoops"
)

func (m *ArbitrarySmartContractCall) Keccak256(nonce uint64) []byte {
	arguments := abi.Arguments{
		abi.Argument{
			Type: whoops.Must(abi.NewType("tuple", "", []abi.ArgumentMarshaling{
				{Name: "addr", Type: "address"},
				{Name: "bytes", Type: "bytes"},
			})),
		},
		abi.Argument{
			Type: whoops.Must(abi.NewType("uint256", "", nil)),
		},
		abi.Argument{
			Type: whoops.Must(abi.NewType("uint256", "", nil)),
		},
	}
	bytes, err := arguments.Pack(
		struct {
			Addr  common.Address
			Bytes []byte
		}{
			common.HexToAddress(m.HexAddress),
			m.Payload,
		},
		big.NewInt(int64(nonce)),
		big.NewInt(123),
	)
	if err != nil {
		panic(err)
	}
	return crypto.Keccak256(bytes)
}
