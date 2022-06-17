package types

import (
	"encoding/binary"
	fmt "fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palomachain/paloma/util/slice"
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
	method := abi.NewMethod("checkpoint", "checkpoint", abi.Function, "", false, false, arguments, abi.Arguments{})

	var bytes32 [32]byte

	copy(bytes32[:], orig.GetTurnstoneID())

	bytes, err := arguments.Pack(
		slice.Map(m.GetValset().GetValidators(), func(s string) common.Address {
			return common.HexToAddress(s)
		}),
		slice.Map(m.GetValset().GetPowers(), func(a uint64) *big.Int {
			return big.NewInt(int64(a))
		}),
		big.NewInt(int64(m.GetValset().GetValsetID())),
		bytes32,
	)
	bytes = append(method.ID[:], bytes...)

	bytes1, _ := arguments.Pack(
		[]common.Address{
			common.HexToAddress("0xe4ab6f4d62ba7e0bbc4cf6c5e8153e105108fba9"),
		},
		[]*big.Int{
			big.NewInt(4294967296),
		},
		big.NewInt(int64(7)),
		bytes32,
	)
	bytes1 = append(method.ID[:], bytes1...)
	fmt.Println(common.Bytes2Hex(bytes1), common.Bytes2Hex(crypto.Keccak256(bytes1)))

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
	method := abi.NewMethod("logic_call", "logic_call", abi.Function, "", false, false, arguments, abi.Arguments{})

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
	bytes = append(method.ID[:], bytes...)

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
