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
