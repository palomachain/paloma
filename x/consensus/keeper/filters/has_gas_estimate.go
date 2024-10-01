package filters

import "github.com/palomachain/paloma/v2/x/consensus/types"

// HasGasEstimate checks if the message has a gas estimate
func HasGasEstimate(msg types.QueuedSignedMessageI) bool {
	if !msg.GetRequireGasEstimation() {
		// Message does not require gas estimation
		return true
	}

	return msg.GetGasEstimate() > 0
}
