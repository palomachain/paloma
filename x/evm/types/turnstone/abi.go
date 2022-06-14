package turnstone

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vizualni/whoops"
)

func (m *UpdateValset) Keccak256() []byte {
	// checkpoint(address[],uint256[],uint256,bytes32)
	arguments := abi.Arguments{
		// addresses
		{Type: whoops.Must(abi.NewType("address[]", "", nil))},
		// powers
		{Type: whoops.Must(abi.NewType("uint256[]", "", nil))},
		// valset id
		{Type: whoops.Must(abi.NewType("uint256", "", nil))},
		// turnstone id
		{Type: whoops.Must(abi.NewType("bytes32", "", nil))},
	}

	addrs := make([]common.Address, len(m.GetValset().HexAddress))

	for i := range m.GetValset().GetHexAddress() {
		addrs[i] = common.HexToAddress(m.GetValset().HexAddress[i])
	}

	bytes, err := arguments.Pack(
		addrs,
		m.GetValset().GetPowers(),
		new(big.Int).SetInt64(int64(m.GetValset().GetValsetID())).Bytes(),
		m.GetValset().GetTurnstoneID(),
	)
	if err != nil {
		panic(err)
	}

	return crypto.Keccak256(bytes)
}
