package keeper

import (
	"encoding/json"

	xchain "github.com/palomachain/paloma/internal/x-chain"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
)

var (
	xchainType = xchain.Type("EVM")
)

func (k Keeper) XChainType() xchain.Type {
	return xchainType
}

func (k Keeper) UnmarshalJob(raw string, chainReferenceID xchain.ReferenceID) (xchain.JobInfo, error) {
	ji := xchain.JobInfo{}

	// for now we only support execution type of messages

	msg := types.SubmitLogicCall{}
	err := json.Unmarshal([]byte(raw), &msg)
	if err != nil {
		return ji, err
	}
	ji.Definition = &msg
	ji.Queue = consensustypes.Queue(
		ConsensusTurnstoneMessage,
		xchainType,
		chainReferenceID,
	)

	return ji, nil
}

func (k Keeper) JobWithPayload(job any, payload any) (any, error) {
	slc, ok := job.(*types.SubmitLogicCall)
	if !ok {
		return nil, xchain.ErrIncorrectJobType.Format(job, &types.SubmitLogicCall{})
	}

	// verify payload
	slc.Payload = payload.([]byte)
	return slc, nil
}
