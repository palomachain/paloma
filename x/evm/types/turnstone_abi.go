package types

import (
	"bytes"
	"encoding/binary"
	"errors"
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

var (
	_ consensustypes.MessageHasher = (*Message)(nil)
	_ consensustypes.MessageHasher = (*ValidatorBalancesAttestation)(nil)
	_ consensustypes.MessageHasher = (*ReferenceBlockAttestation)(nil)
)

func (_m *Message_UpdateValset) keccak256(
	orig *Message,
	_, gasEstimate uint64,
) ([]byte, error) {
	m := _m.UpdateValset
	// checkpoint(address[],uint256[],uint256,bytes32)
	checkpointArgs := abi.Arguments{
		// addresses
		{Type: whoops.Must(abi.NewType("address[]", "", nil))},
		// powers
		{Type: whoops.Must(abi.NewType("uint256[]", "", nil))},
		// valset id
		{Type: whoops.Must(abi.NewType("uint256", "", nil))},
		// turnstone id
		{Type: whoops.Must(abi.NewType("bytes32", "", nil))},
	}
	checkpointMethod := abi.NewMethod("checkpoint", "checkpoint", abi.Function, "", false, false, checkpointArgs, abi.Arguments{})

	var bytes32 [32]byte

	copy(bytes32[:], orig.GetTurnstoneID())

	checkpointBytes, err := checkpointArgs.Pack(
		slice.Map(m.GetValset().GetValidators(), func(s string) common.Address {
			return common.HexToAddress(s)
		}),
		slice.Map(m.GetValset().GetPowers(), func(a uint64) *big.Int {
			return big.NewInt(int64(a))
		}),
		big.NewInt(int64(m.GetValset().GetValsetID())),
		bytes32,
	)
	if err != nil {
		return nil, err
	}

	checkpointBytes = append(checkpointMethod.ID[:], checkpointBytes...)

	checkpointHash := crypto.Keccak256(checkpointBytes)
	var hash32 [32]byte
	copy(hash32[:], checkpointHash)

	// update_valset(bytes32,address,uint256)
	arguments := abi.Arguments{
		// checkpoint
		{Type: whoops.Must(abi.NewType("bytes32", "", nil))},
		// relayer
		{Type: whoops.Must(abi.NewType("address", "", nil))},
		// gas estimate
		{Type: whoops.Must(abi.NewType("uint256", "", nil))},
	}

	method := abi.NewMethod("update_valset", "update_valset", abi.Function, "", false, false, arguments, abi.Arguments{})

	estimate := gasEstimate
	if estimate == 0 {
		// If there's no estimate, we use the same default as pigeon
		estimate = 300_000
	}

	bytes, err := arguments.Pack(
		hash32,
		common.HexToAddress(orig.AssigneeRemoteAddress),
		new(big.Int).SetUint64(gasEstimate),
	)
	if err != nil {
		return nil, err
	}

	bytes = append(method.ID[:], bytes...)

	return crypto.Keccak256(bytes), nil
}

func uint64ToByte(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

func (_m *Message_UploadSmartContract) keccak256(
	_ *Message,
	nonce, _ uint64,
) ([]byte, error) {
	m := _m.UploadSmartContract

	return crypto.Keccak256(append(m.GetBytecode()[:], uint64ToByte(nonce)...)), nil
}

func (_m *Message_SubmitLogicCall) keccak256(
	orig *Message,
	nonce, _ uint64,
) ([]byte, error) {
	m := _m.SubmitLogicCall
	// logic_call((address,bytes),uint256,uint256)
	arguments := abi.Arguments{
		// arguments
		{Type: whoops.Must(abi.NewType("tuple", "", []abi.ArgumentMarshaling{
			{Name: "address", Type: "address"},
			{Name: "payload", Type: "bytes"},
		}))},
		// fee arguments
		{Type: whoops.Must(abi.NewType("tuple", "", []abi.ArgumentMarshaling{
			{Name: "relayer_fee", Type: "uint256"},
			{Name: "community_fee", Type: "uint256"},
			{Name: "security_fee", Type: "uint256"},
			{Name: "fee_payer_paloma_address", Type: "bytes32"},
		}))},
		// message id
		{Type: whoops.Must(abi.NewType("uint256", "", nil))},
		// turnstone id
		{Type: whoops.Must(abi.NewType("bytes32", "", nil))},
		// deadline
		{Type: whoops.Must(abi.NewType("uint256", "", nil))},
		// relayer
		{Type: whoops.Must(abi.NewType("address", "", nil))},
	}

	method := abi.NewMethod("logic_call", "logic_call", abi.Function, "", false, false, arguments, abi.Arguments{})
	var bytes32 [32]byte

	copy(bytes32[:], orig.GetTurnstoneID())

	var fees *SubmitLogicCall_Fees
	if m.Fees == nil {
		// If we have no fees set in the message, we use the same defaults as
		// pigeon
		fees = &SubmitLogicCall_Fees{
			RelayerFee:   100_000,
			CommunityFee: 100_000,
			SecurityFee:  100_000,
		}
	} else {
		fees = m.Fees
	}

	// Left-pad the address with zeroes
	padding := bytes.Repeat([]byte{0}, 32-len(m.SenderAddress))
	senderAddress := [32]byte(append(padding, m.SenderAddress...))

	bytes, err := arguments.Pack(
		struct {
			Address common.Address
			Payload []byte
		}{
			common.HexToAddress(m.GetHexContractAddress()),
			m.GetPayload(),
		},
		struct {
			RelayerFee            *big.Int
			CommunityFee          *big.Int
			SecurityFee           *big.Int
			FeePayerPalomaAddress [32]byte
		}{
			new(big.Int).SetUint64(fees.RelayerFee),
			new(big.Int).SetUint64(fees.CommunityFee),
			new(big.Int).SetUint64(fees.SecurityFee),
			senderAddress,
		},
		new(big.Int).SetInt64(int64(nonce)),
		bytes32,
		big.NewInt(m.GetDeadline()),
		common.HexToAddress(orig.AssigneeRemoteAddress),
	)
	if err != nil {
		return nil, err
	}

	bytes = append(method.ID[:], bytes...)

	return crypto.Keccak256(bytes), nil
}

func (m *Message) SetAssignee(ctx sdk.Context, val, remoteAddr string) {
	m.Assignee = val
	m.AssigneeRemoteAddress = remoteAddr
	m.AssignedAtBlockHeight = math.NewInt(ctx.BlockHeight())
}

func (m *Message) Keccak256WithSignedMessage(q *consensustypes.QueuedSignedMessage) ([]byte, error) {
	type keccak256able interface {
		keccak256(*Message, uint64, uint64) ([]byte, error)
	}
	k, ok := m.GetAction().(keccak256able)
	if !ok {
		return nil, errors.New("message's action is not hashable")
	}

	return k.keccak256(m, q.GetId(), q.GasEstimate)
}

func (m *ValidatorBalancesAttestation) Keccak256WithSignedMessage(_ *consensustypes.QueuedSignedMessage) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString(m.FromBlockTime.String())
	sb.WriteRune('\n')
	for i := range m.ValAddresses {
		sb.WriteString(m.ValAddresses[i].String())
		sb.WriteRune('\t')
		sb.WriteString(m.HexAddresses[i])
		sb.WriteRune('\n')
	}

	return crypto.Keccak256([]byte(sb.String())), nil
}

func (m *ReferenceBlockAttestation) Keccak256WithSignedMessage(_ *consensustypes.QueuedSignedMessage) ([]byte, error) {
	return crypto.Keccak256([]byte(m.FromBlockTime.String())), nil
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
			// This shouldn't matter, but in case it does, we guarantee we do it
			// like Pigeon, so we add a zero signature
			con.originalSignatures = append(con.originalSignatures, nil)
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
