package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/evm/types"
)

var (
	xchainType = xchain.Type("EVM")
)

func (k Keeper) XChainType() xchain.Type {
	return xchainType
}

func (k Keeper) XChainReferenceIDs(ctx sdk.Context) []xchain.ReferenceID {
	chainInfos, err := k.GetAllChainInfos(ctx)
	if err != nil {
		panic(err)
	}

	return slice.Map(chainInfos, func(ci *types.ChainInfo) xchain.ReferenceID {
		return xchain.ReferenceID(ci.GetChainReferenceID())
	})
}

func (k Keeper) UnmarshalJob(definition, payload []byte, chainReferenceID xchain.ReferenceID) (xchain.JobInfo, error) {
	ji := xchain.JobInfo{}

	var jobDefinition types.JobDefinition
	var jobPayload types.JobPayload

	err := json.Unmarshal(definition, &jobDefinition)
	if err != nil {
		return ji, err
	}
	err = json.Unmarshal(payload, &jobPayload)
	if err != nil {
		return ji, err
	}
	// does nothing anymore

	// TODO: this should only verify
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
