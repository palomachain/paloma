package keeper

import (
	"github.com/palomachain/paloma/util/liberr"
)

const (
	ErrChainNotFound                               = liberr.Error("chain with chainReferenceID '%s' was not found")
	ErrChainNotActive                              = liberr.Error("chain with chainReferenceID '%s' is not active")
	ErrNotEnoughValidatorsForGivenChainReferenceID = liberr.Error("not enough validators in the current snapshot to form a proper valset")
	ErrUnexpectedError                             = liberr.Error("unexpected error")
	ErrConsensusNotAchieved                        = liberr.Error("evm: consensus not achieved")
	ErrCannotAddSupportForChainThatExists          = liberr.Error("chain info already exists: %s")
	ErrCannotActiveSmartContractThatIsNotDeploying = liberr.Error("trying to activate a smart contract that is not currently deploying")
)
