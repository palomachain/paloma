package keeper

import "github.com/vizualni/whoops"

const (
	ErrChainNotFound                      = whoops.Errorf("chain with chainID '%s' was not found")
	ErrChainNotActive                     = whoops.Errorf("chain with chainID '%s' is not active")
	ErrNotEnoughValidatorsForGivenChainID = whoops.String("not enough validators in the current snapshot to form a proper valset")
)
