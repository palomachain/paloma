package keeper

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
)

var lastSmartContractKey = []byte{0x1}

func (k Keeper) AllSmartContractsDeployments(ctx sdk.Context) ([]*types.SmartContractDeployment, error) {
	_, res, err := keeperutil.IterAll[*types.SmartContractDeployment](
		k.provideSmartContractDeploymentStore(ctx),
		k.cdc,
	)
	return res, err
}

func (k Keeper) HasAnySmartContractDeployment(ctx sdk.Context, chainReferenceID string) (found bool) {
	if err := keeperutil.IterAllFnc(
		k.provideSmartContractDeploymentStore(ctx),
		k.cdc,
		func(keyArg []byte, item *types.SmartContractDeployment) bool {
			if item.ChainReferenceID == chainReferenceID {
				found = true
				return false
			}
			return true
		},
	); err != nil {
		k.Logger(ctx).Error(
			"error getting smart contract from deployment store by chain Ref",
			"err", err,
			"chainReferenceID", chainReferenceID,
		)
	}
	return
}

func (k Keeper) DeleteSmartContractDeploymentByContractID(ctx sdk.Context, smartContractID uint64, chainReferenceID string) {
	_, key := k.getSmartContractDeploymentByContractID(ctx, smartContractID, chainReferenceID)
	if key == nil {
		return
	}
	k.Logger(ctx).Info("removing a smart contract deployment", "smart-contract-id", smartContractID, "chain-reference-id", chainReferenceID)
	k.provideSmartContractDeploymentStore(ctx).Delete(key)
}

func (k Keeper) SetSmartContractDeploymentStatusByContractID(ctx sdk.Context, smartContractID uint64, chainReferenceID string, status types.SmartContractDeployment_Status) error {
	logger := k.Logger(ctx).WithFields("smart-contract-id", smartContractID, "chain-reference-id", chainReferenceID, "new-smart-contract-status", status)
	v, key := k.getSmartContractDeploymentByContractID(ctx, smartContractID, chainReferenceID)
	if key == nil {
		return keeperutil.ErrNotFound
	}

	v.Status = status
	if err := keeperutil.Save(k.provideSmartContractDeploymentStore(ctx), k.cdc, key, v); err != nil {
		logger.WithError(err).Error("failed to update smart contract deployment record")
		return err
	}

	logger.Debug("updated contract deployment state")
	return nil
}

func (k Keeper) GetLastCompassContract(ctx sdk.Context) (*types.SmartContract, error) {
	kv := k.provideLastCompassContractStore(ctx)
	id := kv.Get(lastSmartContractKey)
	return keeperutil.Load[*types.SmartContract](k.provideSmartContractStore(ctx), k.cdc, id)
}

func (k Keeper) SetAsCompassContract(ctx sdk.Context, smartContract *types.SmartContract) error {
	k.Logger(ctx).Info("setting smart contract as the latest one", "smart-contract-id", smartContract.GetId())
	err := k.setAsLastCompassContract(ctx, smartContract)
	if err != nil {
		return fmt.Errorf("failed to set contract as last smart contract: %w", err)
	}

	err = k.tryDeployingSmartContractToAllChains(ctx, smartContract)
	if err != nil {
		// that's ok. it will try to deploy it on every end blocker
		if !errors.Is(err, ErrConsensusNotAchieved) {
			return fmt.Errorf("failed to deploy smart contract to all chains: %w", err)
		}
	}

	return nil
}

func (k Keeper) SaveNewSmartContract(ctx sdk.Context, abiJSON string, bytecode []byte) (*types.SmartContract, error) {
	smartContract := &types.SmartContract{
		Id:       k.ider.IncrementNextID(ctx, "smart-contract"),
		AbiJSON:  abiJSON,
		Bytecode: bytecode,
	}

	k.Logger(ctx).Info("saving new smart contract", "smart-contract-id", smartContract.GetId())
	err := k.createSmartContract(ctx, smartContract)
	if err != nil {
		return nil, err
	}

	return smartContract, nil
}

func (k Keeper) TryDeployingLastCompassContractToAllChains(ctx sdk.Context) {
	smartContract, err := k.GetLastCompassContract(ctx)
	if err != nil {
		k.Logger(ctx).Error("error while getting latest smart contract", "err", err)
		return
	}
	err = k.tryDeployingSmartContractToAllChains(ctx, smartContract)
	if err != nil {
		k.Logger(ctx).Error("error while trying to deploy smart contract to all chains",
			"err", err,
			"smart-contract-id", smartContract.GetId(),
		)
		return
	}
	k.Logger(ctx).Info("trying to deploy smart contract to all chains",
		"smart-contract-id", smartContract.GetId(),
	)
}

func (k Keeper) AddSmartContractExecutionToConsensus(
	ctx sdk.Context,
	chainReferenceID,
	turnstoneID string,
	logicCall *types.SubmitLogicCall,
) (uint64, error) {
	requirements := &xchain.JobRequirements{
		EnforceMEVRelay: logicCall.ExecutionRequirements.EnforceMEVRelay,
	}
	assignee, err := k.PickValidatorForMessage(ctx, chainReferenceID, requirements)
	if err != nil {
		return 0, err
	}

	return k.ConsensusKeeper.PutMessageInQueue(
		ctx,
		consensustypes.Queue(
			ConsensusTurnstoneMessage,
			xchainType,
			chainReferenceID,
		),
		&types.Message{
			ChainReferenceID: chainReferenceID,
			TurnstoneID:      turnstoneID,
			Action: &types.Message_SubmitLogicCall{
				SubmitLogicCall: logicCall,
			},
			Assignee: assignee,
		}, nil)
}

func (k Keeper) deploySmartContractToChain(ctx sdk.Context, chainInfo *types.ChainInfo, smartContract *types.SmartContract) (retErr error) {
	defer func() {
		args := []any{
			"chain-reference-id", chainInfo.GetChainReferenceID(),
			"smart-contract-id", smartContract.GetId(),
		}
		if retErr != nil {
			args = append(args, "err", retErr)
		}

		if retErr != nil {
			k.Logger(ctx).Error("error adding a message to deploy smart contract to chain", args...)
		} else {
			k.Logger(ctx).Info("added a new smart contract deployment to queue", args...)
		}
	}()
	logger := k.Logger(ctx)
	contractABI, err := abi.JSON(strings.NewReader(smartContract.GetAbiJSON()))
	if err != nil {
		return err
	}

	snapshot, err := k.Valset.GetCurrentSnapshot(ctx)
	var totalShares sdkmath.Int
	if snapshot != nil {
		totalShares = snapshot.TotalShares
	}
	logger.Info(
		"get current snapshot",
		"snapshot-id", snapshot.GetId(),
		"validators-size", len(snapshot.GetValidators()),
		"total-shares", totalShares,
	)

	if err != nil {
		if errors.Is(err, keeperutil.ErrNotFound) {
			logger.WithFields("error", err).Info("cannot deploy due to no consensus")
			return nil
		}

		return err
	}

	valset := transformSnapshotToCompass(snapshot, chainInfo.GetChainReferenceID(), logger)
	logger.Info("returning valset info for deploy smart contract to chain",
		"valset-id", valset.ValsetID,
		"valset-validator-size", len(valset.Validators),
		"valset-power-size", len(valset.Powers),
	)
	if !isEnoughToReachConsensus(valset) {
		k.Logger(ctx).Info(
			"skipping deployment as there are not enough validators to form a consensus",
			"chain-id", chainInfo.GetChainReferenceID(),
			"smart-contract-id", smartContract.GetId(),
			"valset-id", valset.GetValsetID(),
		)
		return whoops.WrapS(
			ErrConsensusNotAchieved,
			"cannot build a valset. valset-id: %d, chain-reference-id: %s, smart-contract-id: %d",
			valset.GetValsetID(), chainInfo.GetChainReferenceID(), smartContract.GetId(),
		)
	}
	uniqueID := generateSmartContractID(ctx)

	k.createSmartContractDeployment(ctx, smartContract, chainInfo, uniqueID[:])

	// set the smart contract constructor arguments
	input, err := contractABI.Pack("", uniqueID, types.TransformValsetToABIValset(valset))

	logger.Info(
		"transform valset to abi valset",
		"valset-id", valset.GetValsetID(),
		"validators-size", len(valset.GetValidators()),
		"power-size", len(valset.GetPowers()),
	)
	if err != nil {
		return err
	}

	vals, err := contractABI.Constructor.Inputs.Unpack(input)
	logger.Debug("[deploySmartContractToChain] UNPACK",
		"ERR", err,
		"ARGS", vals,
	)
	if err != nil {
		return err
	}

	logger.Info(
		"smart contract deployment constructor input",
		"x-chain-type", xchainType,
		"chain-reference-id", chainInfo.GetChainReferenceID(),
		"constructor-input", input,
	)

	assignee, err := k.PickValidatorForMessage(ctx, chainInfo.GetChainReferenceID(), nil)
	if err != nil {
		return err
	}

	_, err = k.ConsensusKeeper.PutMessageInQueue(
		ctx,
		consensustypes.Queue(
			ConsensusTurnstoneMessage,
			xchainType,
			chainInfo.GetChainReferenceID(),
		),
		&types.Message{
			ChainReferenceID: chainInfo.GetChainReferenceID(),
			Action: &types.Message_UploadSmartContract{
				UploadSmartContract: &types.UploadSmartContract{
					Id:               smartContract.GetId(),
					Bytecode:         smartContract.GetBytecode(),
					Abi:              smartContract.GetAbiJSON(),
					ConstructorInput: input,
				},
			},
			Assignee: assignee,
		}, nil)
	return err
}

func (k Keeper) getSmartContract(ctx sdk.Context, id uint64) (*types.SmartContract, error) {
	return keeperutil.Load[*types.SmartContract](k.provideSmartContractStore(ctx), k.cdc, keeperutil.Uint64ToByte(id))
}

func (k Keeper) createSmartContract(ctx sdk.Context, smartContract *types.SmartContract) error {
	return keeperutil.Save(k.provideSmartContractStore(ctx), k.cdc, keeperutil.Uint64ToByte(smartContract.GetId()), smartContract)
}

func (k Keeper) setAsLastCompassContract(ctx sdk.Context, smartContract *types.SmartContract) error {
	kv := k.provideLastCompassContractStore(ctx)
	kv.Set(lastSmartContractKey, keeperutil.Uint64ToByte(smartContract.GetId()))
	return nil
}

func (k Keeper) tryDeployingSmartContractToAllChains(ctx sdk.Context, smartContract *types.SmartContract) error {
	var g whoops.Group
	chainInfos, err := k.GetAllChainInfos(ctx)
	if err != nil {
		return err
	}

	for _, chainInfo := range chainInfos {
		k.Logger(ctx).Info("trying to deploy smart contract to EVM chain", "smart-contract-id", smartContract.GetId(), "chain-reference-id", chainInfo.GetChainReferenceID())
		if k.HasAnySmartContractDeployment(ctx, chainInfo.GetChainReferenceID()) {
			// TODO: Only wait if the status is IN_FLIGHT
			// TODO: We probably want to still delete the deployment in case of error AS LONG as we haven't sent a move ownership message
			// we are already deploying to this chain. Lets wait it out.
			continue
		}
		if chainInfo.GetActiveSmartContractID() >= smartContract.GetId() {
			// the chain has the newer version of the chain, so skipping the "old" smart contract upgrade
			continue
		}
		k.Logger(ctx).Info("deploying smart contracts actually",
			"smart-contract-id", smartContract.GetId(),
			"chain-reference-id", chainInfo.GetChainReferenceID())
		g.Add(k.deploySmartContractToChain(ctx, chainInfo, smartContract))
	}

	if g.Err() {
		return g
	}

	return nil
}

func (k Keeper) createSmartContractDeployment(
	ctx sdk.Context,
	smartContract *types.SmartContract,
	chainInfo *types.ChainInfo,
	uniqueID []byte,
) *types.SmartContractDeployment {
	if foundItem, _ := k.getSmartContractDeploymentByContractID(ctx, smartContract.GetId(), chainInfo.GetChainReferenceID()); foundItem != nil {
		k.Logger(ctx).Error(
			"smart contract is already deploying",
			"smart-contract-id", smartContract.GetId(),
			"smart-contract-status", foundItem.GetStatus(),
			"chain-reference-id", chainInfo.GetChainReferenceID(),
		)
		return foundItem
	}

	item := &types.SmartContractDeployment{
		SmartContractID:  smartContract.GetId(),
		ChainReferenceID: chainInfo.GetChainReferenceID(),
		Status:           types.SmartContractDeployment_IN_FLIGHT,
		UniqueID:         uniqueID,
	}

	id := k.ider.IncrementNextID(ctx, "smart-contract-deploying")

	if err := keeperutil.Save(
		k.provideSmartContractDeploymentStore(ctx),
		k.cdc,
		keeperutil.Uint64ToByte(id),
		item,
	); err != nil {
		k.Logger(ctx).Error("error setting smart contract in deployment store", "err", err)
	}

	k.Logger(ctx).Info("setting smart contract in deployment state", "smart-contract-id", smartContract.GetId(), "chain-reference-id", chainInfo.GetChainReferenceID())

	return item
}

func (k Keeper) getSmartContractDeploymentByContractID(ctx sdk.Context, smartContractID uint64, chainReferenceID string) (res *types.SmartContractDeployment, key []byte) {
	if err := keeperutil.IterAllFnc(
		k.provideSmartContractDeploymentStore(ctx),
		k.cdc,
		func(keyArg []byte, item *types.SmartContractDeployment) bool {
			if item.ChainReferenceID == chainReferenceID && item.SmartContractID == smartContractID {
				res = item
				key = keyArg
				return false
			}
			return true
		},
	); err != nil {
		k.Logger(ctx).Error(
			"error getting smart contract from deployment store by contractID, chainRef",
			"err", err,
			"smartContractID", smartContractID,
			"chainReferenceID", chainReferenceID,
		)
	}
	return
}

func (k Keeper) provideSmartContractDeploymentStore(ctx sdk.Context) storetypes.KVStore {
	return prefix.NewStore(ctx.KVStore(k.memKey), []byte("smart-contract-deployment"))
}

func (k Keeper) provideSmartContractStore(ctx sdk.Context) storetypes.KVStore {
	return prefix.NewStore(ctx.KVStore(k.memKey), []byte("smart-contracts"))
}

func (k Keeper) provideLastCompassContractStore(ctx sdk.Context) storetypes.KVStore {
	return prefix.NewStore(ctx.KVStore(k.memKey), []byte("latest-smart-contract"))
}
