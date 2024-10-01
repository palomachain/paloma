package types

import (
	keeperutil "github.com/palomachain/paloma/v2/util/keeper"
)

const (
	ItemRemovedEventKey                                   = "ConsensusQueueItemRemoved"
	ItemRemovedEventID          keeperutil.EventAttribute = "ItemID"
	ItemRemovedEventQueueName   keeperutil.EventAttribute = "ConsensusQueueName"
	ItemRemovedChainReferenceID keeperutil.EventAttribute = "ChainReferenceID"
)
