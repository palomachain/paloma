package filters

import "github.com/palomachain/paloma/x/consensus/types"

func IsUnprocessed(msg types.QueuedSignedMessageI) bool {
	return msg.GetPublicAccessData() == nil && msg.GetErrorData() == nil
}
