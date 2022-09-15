package keeper

import "github.com/vizualni/whoops"

const (
	ErrChainNotFound                               = whoops.Errorf("chain with chainReferenceID '%s' was not found")
	ErrChainNotActive                              = whoops.Errorf("chain with chainReferenceID '%s' is not active")
	ErrNotEnoughValidatorsForGivenChainReferenceID = whoops.String("not enough validators in the current snapshot to form a proper valset")
	ErrUnexpectedError                             = whoops.String("unexpected error")
	ErrConsensusNotAchieved                        = whoops.String("evm: consensus not achieved")
	ErrCannotAddSupportForChainThatExists          = whoops.Errorf("chain info already exists: %s")
	ErrCannotActiveSmartContractThatIsNotDeploying = whoops.String("trying to activate a smart contract that is not currently deploying")

	ErrWasmExecuteMessageNotValid = whoops.String("message is not valid")
)
