package filters

import evmtypes "github.com/palomachain/paloma/v2/x/evm/types"

func IsAssignedTo(msg *evmtypes.Message, assignee string) bool {
	return msg.GetAssignee() == assignee
}
