package types

import (
	"encoding/binary"
	"math/big"
	"strings"

	"cosmossdk.io/math"
	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palomachain/paloma/util/slice"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
)

type Signature struct {
	V *big.Int
	R *big.Int
	S *big.Int
}

type CompassValset struct {
	ValsetId   *big.Int
	Validators []common.Address
	Powers     []*big.Int
}

type CompassConsensus struct {
	Valset     CompassValset
	Signatures []Signature

	originalSignatures [][]byte
}

type CompassLogicCallArgs struct {
	Payload              []byte
	LogicContractAddress common.Address
}

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
			{Name: "address", Type: "address"},
			{Name: "payload", Type: "bytes"},
		}))},
		// message id
		{Type: whoops.Must(abi.NewType("uint256", "", nil))},
		// turnstone id
		{Type: whoops.Must(abi.NewType("bytes32", "", nil))},
		// deadline
		{Type: whoops.Must(abi.NewType("uint256", "", nil))},
	}

	method := abi.NewMethod("logic_call", "logic_call", abi.Function, "", false, false, arguments, abi.Arguments{})
	var bytes32 [32]byte

	copy(bytes32[:], orig.GetTurnstoneID())

	bytes, err := arguments.Pack(
		struct {
			Address common.Address
			Payload []byte
		}{
			common.HexToAddress(m.GetHexContractAddress()),
			m.GetPayload(),
		},
		new(big.Int).SetInt64(int64(nonce)),
		bytes32,
		big.NewInt(m.GetDeadline()),
	)
	if err != nil {
		panic(err)
	}

	bytes = append(method.ID[:], bytes...)

	return crypto.Keccak256(bytes)
}

func (m *Message) SetAssignee(ctx sdk.Context, val string) {
	m.Assignee = val
	m.AssignedAtBlockHeight = math.NewInt(ctx.BlockHeight())
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

func (m *ValidatorBalancesAttestation) Keccak256(nonce uint64) []byte {
	var sb strings.Builder
	sb.WriteString(m.FromBlockTime.String())
	sb.WriteRune('\n')
	for i := range m.ValAddresses {
		sb.WriteString(m.ValAddresses[i].String())
		sb.WriteRune('\t')
		sb.WriteString(m.HexAddresses[i])
		sb.WriteRune('\n')
	}

	return crypto.Keccak256([]byte(sb.String()))
}

func (m *ReferenceBlockAttestation) Keccak256(nonce uint64) []byte {
	return crypto.Keccak256([]byte(m.FromBlockTime.String()))
}

func BuildCompassConsensus(
	v *Valset,
	signatures []*consensustypes.SignData,
) CompassConsensus {
	signatureMap := slice.MakeMapKeys(
		signatures,
		func(sig *consensustypes.SignData) string {
			return sig.ExternalAccountAddress
		},
	)
	con := CompassConsensus{
		Valset: TransformValsetToCompassValset(v),
	}

	for i := range v.GetValidators() {
		sig, ok := signatureMap[v.GetValidators()[i]]
		if !ok {
			con.Signatures = append(con.Signatures,
				Signature{
					V: big.NewInt(0),
					R: big.NewInt(0),
					S: big.NewInt(0),
				})
		} else {
			con.Signatures = append(con.Signatures,
				Signature{
					V: new(big.Int).SetInt64(int64(sig.Signature[64]) + 27),
					R: new(big.Int).SetBytes(sig.Signature[:32]),
					S: new(big.Int).SetBytes(sig.Signature[32:64]),
				},
			)

			con.originalSignatures = append(con.originalSignatures, sig.Signature)
		}
	}

	return con
}

func TransformValsetToCompassValset(val *Valset) CompassValset {
	return CompassValset{
		Validators: slice.Map(val.GetValidators(), func(s string) common.Address {
			return common.HexToAddress(s)
		}),
		Powers: slice.Map(val.GetPowers(), func(p uint64) *big.Int {
			return big.NewInt(int64(p))
		}),
		ValsetId: big.NewInt(int64(val.GetValsetID())),
	}
}
