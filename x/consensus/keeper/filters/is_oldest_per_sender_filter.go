package filters

import evmtypes "github.com/palomachain/paloma/x/evm/types"

// Biased: Filter expects messages to be ordered by message ID (default)
// Filter requires injection of a look-up table to keep track of senders
// Filter out multiple messages from same sender
// This is done to avoid possible relays of more messages than
// user is able to pay for.
func IsOldestMsgPerSender(lut map[string]struct{}, msg *evmtypes.Message) bool {
	slc := msg.GetSubmitLogicCall()
	if slc == nil {
		return true
	}
	if len(slc.GetSenderAddress()) < 1 {
		return true
	}
	sender := string(slc.GetSenderAddress())
	if _, fnd := lut[sender]; fnd {
		// One message for this sender was already included in the
		// result set.
		return false
	}

	lut[sender] = struct{}{}
	return true
}
