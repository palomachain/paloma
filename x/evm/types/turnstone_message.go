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
