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
