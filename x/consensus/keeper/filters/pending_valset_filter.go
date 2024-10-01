package filters

import "github.com/palomachain/paloma/v2/x/consensus/types"

func IsNotBlockedByValset(pendingValsetUpdates []types.QueuedSignedMessageI, msg types.QueuedSignedMessageI) bool {
	if pendingValsetUpdates == nil || len(pendingValsetUpdates) < 1 {
		return true
	}

	// Looks like there is a valset update for the target chain,
	// only return true if this message is younger than the valset update
	return msg.GetId() <= pendingValsetUpdates[0].GetId()
}
