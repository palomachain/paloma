package keeper

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/evm/types"
)

var xchainType = xchain.Type("evm")

var _ xchain.Bridge = Keeper{}

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

func (k Keeper) unmarshalJob(definition, payload []byte, chainReferenceID xchain.ReferenceID) (resdef types.JobDefinition, respay types.JobPayload, reserr error) {
	var jobDefinition types.JobDefinition
	var jobPayload types.JobPayload

	err := json.Unmarshal(definition, &jobDefinition)
	if err != nil {
		reserr = err
		return
	}
	err = json.Unmarshal(payload, &jobPayload)
	if err != nil {
		reserr = err
		return
	}

	return jobDefinition, jobPayload, nil
}

func (k Keeper) VerifyJob(ctx sdk.Context, definition, payload []byte, chainReferenceID xchain.ReferenceID) error {
	_, _, err := k.unmarshalJob(definition, payload, chainReferenceID)
	return err
}

// ExecuteJob schedules the definition and payload for execution via consensus queue
func (k Keeper) ExecuteJob(ctx sdk.Context, definition, payload []byte, senderPubKey []byte, contractAddress []byte, chainReferenceID xchain.ReferenceID) error {
	def, load, err := k.unmarshalJob(definition, payload, chainReferenceID)
	if err != nil {
		return err
	}
	ci, err := k.GetChainInfo(ctx, chainReferenceID)
	if err != nil {
		return err
	}

	modifiedPayload := common.FromHex(load.GetHexPayload())

	switch {
	case senderPubKey != nil:
		modifiedPayload = append(modifiedPayload, senderPubKey...)
	case contractAddress != nil:
		modifiedPayload = append(modifiedPayload, contractAddress...)
	}

	return k.AddSmartContractExecutionToConsensus(
		ctx,
		chainReferenceID,
		string(ci.GetSmartContractUniqueID()),
		&types.SubmitLogicCall{
			HexContractAddress: def.GetAddress(),
			Abi:                common.FromHex(def.GetABI()),
			Payload:            modifiedPayload,
			Deadline:           ctx.BlockTime().Add(10 * time.Minute).Unix(),
			SenderPubKey:       senderPubKey,
			ContractAddress:    contractAddress,
		},
	)
}
