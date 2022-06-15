package turnstone

import (
	"encoding/binary"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vizualni/whoops"
)

func (_m *Message_UpdateValset) keccak256(orig *Message, _ uint64) []byte {
	m := _m.UpdateValset
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
		orig.GetTurnstoneID(),
	)
	if err != nil {
		panic(err)
	}

	return crypto.Keccak256(bytes)
}
func uint64ToByte(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

func (_m *Message_UploadSmartContract) keccak256(orig *Message, nonce uint64) []byte {
	m := _m.UploadSmartContract

	return crypto.Keccak256(append(m.GetBytecode()[:], uint64ToByte(nonce)...))
}

func (_m *Message_SubmitLogicCall) keccak256(orig *Message, nonce uint64) []byte {
	m := _m.SubmitLogicCall
	// logic_call((address,bytes),uint256,uint256)
	arguments := abi.Arguments{
		// arguments
		{Type: whoops.Must(abi.NewType("tuple", "", []abi.ArgumentMarshaling{
			{Type: "address"},
			{Type: "bytes"},
		}))},
		// message id
		{Type: whoops.Must(abi.NewType("uint256", "", nil))},
		// turnstone id
		{Type: whoops.Must(abi.NewType("bytes", "", nil))},
		// deadline
		{Type: whoops.Must(abi.NewType("uint256", "", nil))},
	}

	bytes, err := arguments.Pack(
		[]any{
			common.HexToAddress(m.GetHexContractAddress()),
			m.GetPayload(),
		},
		new(big.Int).SetInt64(int64(nonce)).Bytes(),
		orig.GetTurnstoneID(),
		m.GetDeadline(),
	)

	if err != nil {
		panic(err)
	}

	return crypto.Keccak256(bytes)
}

func (m *Message) Keccak256(nonce uint64) []byte {
	type keccak256able interface {
		keccak256(*Message, uint64) []byte
	}
	k, ok := m.GetAction().(keccak256able)
	if !ok {
		panic("message's action is not hashable")
	}
	return k.keccak256(m, nonce)
}
