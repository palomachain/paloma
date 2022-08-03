package types

import (
	keeperutil "github.com/palomachain/paloma/util/keeper"
)

const (
	ItemRemovedEventKey                                   = "ConsensusQueueItemRemoved"
	ItemRemovedEventID          keeperutil.EventAttribute = "ItemID"
	ItemRemovedEventQueueName   keeperutil.EventAttribute = "ConsensusQueueName"
	ItemRemovedChainReferenceID keeperutil.EventAttribute = "ChainReferenceID"
)
