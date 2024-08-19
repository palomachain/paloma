package types

import "github.com/cosmos/gogoproto/proto"

type TurnstoneMsg interface {
	proto.Message
	GetAssignee() string
}

var (
	_ TurnstoneMsg = &Message{}
	_ TurnstoneMsg = &ValidatorBalancesAttestation{}
	_ TurnstoneMsg = &CollectFunds{}
)

// SmartContractUploader defines an interface to unify upload operation on both
// compass and user-defined smart contract uploads
type SmartContractUploader interface {
	GetConstructorInput() []byte
	GetAbi() string
	GetBytecode() []byte
}

type FeePayer interface {
	SetFees(fees *Fees)
}

var (
	_ FeePayer = &Message_SubmitLogicCall{}
	_ FeePayer = &Message_UploadUserSmartContract{}
)

func (m *Message_SubmitLogicCall) SetFees(fees *Fees) {
	m.SubmitLogicCall.Fees = fees
}

func (m *Message_UploadUserSmartContract) SetFees(fees *Fees) {
	m.UploadUserSmartContract.Fees = fees
}
