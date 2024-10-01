package types

import (
	keeperutil "github.com/palomachain/paloma/v2/util/keeper"
)

const (
	SmartContractExecutionFailedKey              string                    = "SmartContractExecutionFailed"
	SmartContractExecutionFailedMessageID        keeperutil.EventAttribute = "MessageID"
	SmartContractExecutionFailedChainReferenceID keeperutil.EventAttribute = "ChainReferenceID"
	SmartContractExecutionFailedError            keeperutil.EventAttribute = "ErrorMessage"
	SmartContractExecutionMessageType            keeperutil.EventAttribute = "MessageType"
)

const (
	AttestingUpdateValsetRemoveOldMessagesKey string = "AttestingUpdateValsetRemoveOldMessages"
)
