package keeper

import (
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
)

type registry struct {
	slice []consensus.SupportsConsensusQueue
}

func (r *registry) Add(ss ...consensus.SupportsConsensusQueue) {
	r.slice = append(r.slice, ss...)
}

func NewRegistry() *registry {
	return &registry{
		slice: []consensus.SupportsConsensusQueue{},
	}
}
