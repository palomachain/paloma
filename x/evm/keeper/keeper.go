package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	wasmutil "github.com/palomachain/paloma/util/wasm"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	ptypes "github.com/palomachain/paloma/x/paloma/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/vizualni/whoops"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/x/evm/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

const (
	maxPower                     = 1 << 32
	thresholdForConsensus uint64 = 2_863_311_530
)
const (
	ConsensusTurnstoneMessage     = "evm-turnstone-message"
	ConsensusGetValidatorBalances = "validators-balances"
	SignaturePrefix               = "\x19Ethereum Signed Message:\n32"
)

var _ ptypes.ExternalChainSupporterKeeper = Keeper{}

type supportedChainInfo struct {
	batch                 bool
	msgType               any
	processAttesationFunc func(Keeper) func(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) error
}

var SupportedConsensusQueues = map[string]supportedChainInfo{
	ConsensusTurnstoneMessage: {
		batch:   false,
		msgType: &types.Message{},
		processAttesationFunc: func(k Keeper) func(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) error {
			return k.attestRouter
		},
	},
	ConsensusGetValidatorBalances: {
		batch:   false,
		msgType: &types.ValidatorBalancesAttestation{},
		processAttesationFunc: func(k Keeper) func(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) error {
			return k.attestValidatorBalances
		},
	},
}

type evmChainTemp struct {
	chainReferenceID string
	turnstoneID      string
}

func (e evmChainTemp) ChainReferenceID() string {
	return e.chainReferenceID
}

var _ valsettypes.OnSnapshotBuiltListener = Keeper{}

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	memKey     sdk.StoreKey
	paramstore paramtypes.Subspace

	ConsensusKeeper types.ConsensusKeeper
	Valset          types.ValsetKeeper
	ider            keeperutil.IDGenerator
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	k := &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
	}

	k.ider = keeperutil.NewIDGenerator(keeperutil.StoreGetterFn(k.smartContractsStore), []byte("id-key"))

	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) AddSmartContractExecutionToConsensus(
	ctx sdk.Context,
	chainReferenceID,
	turnstoneID string,
	logicCall *types.SubmitLogicCall,
) error {
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
	contractABI, err := abi.JSON(strings.NewReader(smartContract.GetAbiJSON()))
	if err != nil {
		return err
	}

	snapshot, err := k.Valset.GetCurrentSnapshot(ctx)
	switch {
	case err == nil:
		// does nothing
	case errors.Is(err, keeperutil.ErrNotFound):
		// can't deploy as there is no consensus
		return nil
	default:
		return err
	}
	valset := transformSnapshotToCompass(snapshot, chainInfo.GetChainReferenceID())

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

	k.setSmartContractAsDeploying(ctx, smartContract, chainInfo, uniqueID[:])

	// set the smart contract constructor arguments
	input, err := contractABI.Pack("", uniqueID, types.TransformValsetToABIValset(valset))
	if err != nil {
		return err
	}

	return k.ConsensusKeeper.PutMessageInQueue(
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
		}, nil)
}

func (k Keeper) ChangeMinOnChainBalance(ctx sdk.Context, chainReferenceID string, balance *big.Int) error {
	ci, err := k.GetChainInfo(ctx, chainReferenceID)
	if err != nil {
		return err
	}
	ci.MinOnChainBalance = balance.Text(10)
	return k.updateChainInfo(ctx, ci)
}

func (k Keeper) SaveNewSmartContract(ctx sdk.Context, abiJSON string, bytecode []byte) (*types.SmartContract, error) {
	smartContract := &types.SmartContract{
		Id:       k.ider.IncrementNextID(ctx, "smart-contract"),
		AbiJSON:  abiJSON,
		Bytecode: bytecode,
	}

	err := k.saveSmartContract(ctx, smartContract)
	if err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("saving new smart contract", "smart-contract-id", smartContract.GetId())
	err = k.setAsLastSmartContract(ctx, smartContract)
	if err != nil {
		return nil, err
	}
	k.Logger(ctx).Info("setting smart contract as the latest one", "smart-contract-id", smartContract.GetId())

	err = k.tryDeployingSmartContractToAllChains(ctx, smartContract)
	if err != nil {
		// that's ok. it will try to deploy it on every end blocker
		if !errors.Is(err, ErrConsensusNotAchieved) {
			return nil, err
		}
	}

	return smartContract, nil
}

func (k Keeper) TryDeployingLastSmartContractToAllChains(ctx sdk.Context) {
	smartContract, err := k.GetLastSmartContract(ctx)
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

func (k Keeper) tryDeployingSmartContractToAllChains(ctx sdk.Context, smartContract *types.SmartContract) error {
	var g whoops.Group
	chainInfos, err := k.GetAllChainInfos(ctx)

	if err != nil {
		return err
	}

	for _, chainInfo := range chainInfos {
		k.Logger(ctx).Info("trying to deploy smart contract to EVM chain", "smart-contract-id", smartContract.GetId(), "chain-reference-id", chainInfo.GetChainReferenceID())
		if k.HasAnySmartContractDeployment(ctx, chainInfo.GetChainReferenceID()) {
			// we are already deploying to this chain. Lets wait it out.
			continue
		}
		if chainInfo.GetActiveSmartContractID() >= smartContract.GetId() {
			// the chain has the newer version of the chain, so skipping the "old" smart contract upgrade
			continue
		}
		g.Add(k.deploySmartContractToChain(ctx, chainInfo, smartContract))
	}

	if g.Err() {
		return g
	}

	return nil
}

type ExecuteEVMFromCosmWasm struct {
	TargetContractInfo struct {
		ChainReferenceID     string `json:"chain_id"`
		SmartContractAddress string `json:"contract_address"`
		SmartContractABI     string `json:"smart_contract_abi"`

		CompassID string `json:"compass_id"`
	} `json:"target_contract_info"`

	Payload []byte `json:"payload"`
}

func (e ExecuteEVMFromCosmWasm) valid() error {
	zero := ExecuteEVMFromCosmWasm{}
	if e.TargetContractInfo == zero.TargetContractInfo {
		return whoops.String("target contract info is empty")
	}
	if len(e.Payload) == 0 {
		return whoops.String("payload bytes is empty")
	}
	// todo: add more in the future
	return nil
}

func (k Keeper) WasmMessengerHandler() wasmutil.MessengerFnc {
	return func(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
		var executeMsg ExecuteEVMFromCosmWasm
		err := json.Unmarshal(msg.Custom, &executeMsg)
		if err != nil {
			return nil, nil, err
		}
		if err = executeMsg.valid(); err != nil {
			return nil, nil, whoops.Wrap(err, ErrWasmExecuteMessageNotValid)
		}

		ci, err := k.GetChainInfo(ctx, executeMsg.TargetContractInfo.ChainReferenceID)
		if err != nil {
			return nil, nil, err
		}

		if !ci.IsActive() {
			return nil, nil, ErrChainNotActive.Format(ci.GetChainReferenceID())
		}

		err = k.AddSmartContractExecutionToConsensus(
			ctx,
			executeMsg.TargetContractInfo.ChainReferenceID,
			executeMsg.TargetContractInfo.CompassID,
			&types.SubmitLogicCall{
				HexContractAddress: executeMsg.TargetContractInfo.SmartContractAddress,
				Payload:            executeMsg.Payload,
				Deadline:           ctx.BlockTime().UTC().Add(10 * time.Minute).Unix(),
				Abi:                []byte(executeMsg.TargetContractInfo.SmartContractABI),
			},
		)

		if err != nil {
			return nil, nil, err
		}

		return nil, nil, nil
	}

}

// func (k Keeper) OnSchedulerMessageProcess(ctx sdk.Context, rawMsg any) (processed bool, err error) {
// 	// when scheduler ticks then this gets executed

// 	processed = true
// 	switch msg := rawMsg.(type) {
// 	case *types.ArbitrarySmartContractCall:
// 		err = k.AddSmartContractExecutionToConsensus(
// 			ctx,
// 			msg,
// 		)
// 	default:
// 		processed = false
// 	}

// 	return
// }

func (k Keeper) SupportedQueues(ctx sdk.Context) (map[string]consensus.SupportsConsensusQueueAction, error) {
	chains, err := k.GetAllChainInfos(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]consensus.SupportsConsensusQueueAction)

	for _, chainInfo := range chains {
		// if !chainInfo.IsActive() {
		// 	continue
		// }
		for subQueue, queueInfo := range SupportedConsensusQueues {
			queue := consensustypes.Queue(subQueue, xchainType, xchain.ReferenceID(chainInfo.ChainReferenceID))
			opts := *consensus.ApplyOpts(nil,
				consensus.WithChainInfo(xchainType, chainInfo.ChainReferenceID),
				consensus.WithQueueTypeName(queue),
				consensus.WithStaticTypeCheck(queueInfo.msgType),
				consensus.WithBytesToSignCalc(
					consensustypes.BytesToSignFunc(func(msg consensustypes.ConsensusMsg, salt consensustypes.Salt) []byte {
						k := msg.(interface {
							Keccak256(uint64) []byte
						})
						return k.Keccak256(salt.Nonce)
					}),
				),
				consensus.WithVerifySignature(func(bz []byte, sig []byte, address []byte) bool {
					receivedAddr := common.BytesToAddress(address)

					bytesToVerify := crypto.Keccak256(append(
						[]byte(SignaturePrefix),
						bz...,
					))
					recoveredPk, err := crypto.Ecrecover(bytesToVerify, sig)
					if err != nil {
						return false
					}
					pk, err := crypto.UnmarshalPubkey(recoveredPk)
					if err != nil {
						return false
					}
					recoveredAddr := crypto.PubkeyToAddress(*pk)
					return receivedAddr.Hex() == recoveredAddr.Hex()
				}),
			)

			res[queue] = consensus.SupportsConsensusQueueAction{
				QueueOptions:                 opts,
				ProcessMessageForAttestation: queueInfo.processAttesationFunc(k),
			}
		}
	}

	return res, nil
}

func (k Keeper) GetAllChainInfos(ctx sdk.Context) ([]*types.ChainInfo, error) {
	_, all, err := keeperutil.IterAll[*types.ChainInfo](k.chainInfoStore(ctx), k.cdc)
	return all, err
}

func (k Keeper) GetChainInfo(ctx sdk.Context, targetChainReferenceID string) (*types.ChainInfo, error) {
	res, err := keeperutil.Load[*types.ChainInfo](k.chainInfoStore(ctx), k.cdc, []byte(targetChainReferenceID))
	if errors.Is(err, keeperutil.ErrNotFound) {
		return nil, ErrChainNotFound.Format(targetChainReferenceID)
	}
	return res, nil
}

func (k Keeper) updateChainInfo(ctx sdk.Context, chainInfo *types.ChainInfo) error {
	return keeperutil.Save(k.chainInfoStore(ctx), k.cdc, []byte(chainInfo.GetChainReferenceID()), chainInfo)
}

func (k Keeper) AddSupportForNewChain(
	ctx sdk.Context,
	chainReferenceID string,
	chainID uint64,
	blockHeight uint64,
	blockHashAtHeight string,
	minimumOnChainBalance *big.Int,
) error {
	_, err := k.GetChainInfo(ctx, chainReferenceID)
	switch {
	case err == nil:
		return ErrCannotAddSupportForChainThatExists.Format(chainReferenceID)
	case errors.Is(err, ErrChainNotFound):
		// we want chain not to exist when adding a new one!
	default:
		return whoops.Wrap(ErrUnexpectedError, err)
	}
	all, err := k.GetAllChainInfos(ctx)
	if err != nil {
		return err
	}
	for _, existing := range all {
		if existing.GetChainID() == chainID {
			return ErrCannotAddSupportForChainThatExists.Format(chainReferenceID).
				WrapS("chain with chainID %d already exists", chainID)
		}
	}

	chainInfo := &types.ChainInfo{
		ChainID:              chainID,
		ChainReferenceID:     chainReferenceID,
		ReferenceBlockHeight: blockHeight,
		ReferenceBlockHash:   blockHashAtHeight,
		MinOnChainBalance:    minimumOnChainBalance.Text(10),
	}

	err = k.updateChainInfo(ctx, chainInfo)
	if err != nil {
		return err
	}

	k.TryDeployingLastSmartContractToAllChains(ctx)
	return nil
}

func (k Keeper) ActivateChainReferenceID(
	ctx sdk.Context,
	chainReferenceID string,
	smartContract *types.SmartContract,
	smartContractAddr string,
	smartContractUniqueID []byte,
) (retErr error) {
	defer func() {
		args := []any{
			"chain-reference-id", chainReferenceID,
			"smart-contract-id", smartContract.GetId(),
			"smart-contract-addr", smartContractAddr,
			"smart-contract-unique-id", smartContractUniqueID,
		}
		if retErr != nil {
			args = append(args, "err", retErr)
		}

		if retErr != nil {
			k.Logger(ctx).Error("error while activating chain with a new smart contract", args...)
		} else {
			k.Logger(ctx).Info("activated chain with a new smart contract", args...)
		}
	}()
	chainInfo, err := k.GetChainInfo(ctx, chainReferenceID)
	if err != nil {
		return err
	}
	// if this is called with version lower than the current one, then do nothing
	if chainInfo.GetActiveSmartContractID() >= smartContract.GetId() {
		return nil
	}
	chainInfo.Status = types.ChainInfo_ACTIVE
	chainInfo.Abi = smartContract.GetAbiJSON()
	chainInfo.Bytecode = smartContract.GetBytecode()
	chainInfo.ActiveSmartContractID = smartContract.GetId()

	chainInfo.SmartContractAddr = smartContractAddr
	chainInfo.SmartContractUniqueID = smartContractUniqueID

	k.RemoveSmartContractDeployment(ctx, smartContract.GetId(), chainInfo.GetChainReferenceID())

	return k.updateChainInfo(ctx, chainInfo)
}

func (k Keeper) RemoveSupportForChain(ctx sdk.Context, proposal *types.RemoveChainProposal) error {
	_, err := k.GetChainInfo(ctx, proposal.GetChainReferenceID())
	if err != nil {
		return err
	}

	k.chainInfoStore(ctx).Delete([]byte(proposal.GetChainReferenceID()))

	for subQueue := range SupportedConsensusQueues {
		queue := consensustypes.Queue(subQueue, xchainType, xchain.ReferenceID(proposal.GetChainReferenceID()))
		k.ConsensusKeeper.RemoveConsensusQueue(ctx, queue)
	}

	return nil
}

func (k Keeper) smartContractDeploymentStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("smart-contract-deployment"))
}

func (k Keeper) chainInfoStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("chain-info"))
}

func (k Keeper) smartContractsStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("smart-contracts"))
}

func (k Keeper) setSmartContractAsDeploying(
	ctx sdk.Context,
	smartContract *types.SmartContract,
	chainInfo *types.ChainInfo,
	uniqueID []byte,
) *types.SmartContractDeployment {

	if foundItem, _ := k.getSmartContractDeploying(ctx, smartContract.GetId(), chainInfo.GetChainReferenceID()); foundItem != nil {
		k.Logger(ctx).Error(
			"smart contract is already deploying",
			"smart-contract-id", smartContract.GetId(),
			"chain-reference-id", chainInfo.GetChainReferenceID(),
		)
		return foundItem
	}

	item := &types.SmartContractDeployment{
		SmartContractID:  smartContract.GetId(),
		ChainReferenceID: chainInfo.GetChainReferenceID(),
		UniqueID:         uniqueID,
	}

	id := k.ider.IncrementNextID(ctx, "smart-contract-deploying")

	keeperutil.Save(
		k.smartContractDeploymentStore(ctx),
		k.cdc,
		keeperutil.Uint64ToByte(id),
		item,
	)

	k.Logger(ctx).Info("setting smart contract in deployment state", "smart-contract-id", smartContract.GetId(), "chain-reference-id", chainInfo.GetChainReferenceID())

	return item
}

func (k Keeper) getSmartContractDeploying(ctx sdk.Context, smartContractID uint64, chainReferenceID string) (res *types.SmartContractDeployment, key []byte) {
	keeperutil.IterAllFnc(
		k.smartContractDeploymentStore(ctx),
		k.cdc,
		func(keyArg []byte, item *types.SmartContractDeployment) bool {
			if item.ChainReferenceID == chainReferenceID && item.SmartContractID == smartContractID {
				res = item
				key = keyArg
				return false
			}
			return true
		})
	return
}

func (k Keeper) AllSmartContractsDeployments(ctx sdk.Context) ([]*types.SmartContractDeployment, error) {
	_, res, err := keeperutil.IterAll[*types.SmartContractDeployment](
		k.smartContractDeploymentStore(ctx),
		k.cdc,
	)
	return res, err
}

func (k Keeper) HasAnySmartContractDeployment(ctx sdk.Context, chainReferenceID string) (found bool) {
	keeperutil.IterAllFnc(
		k.smartContractDeploymentStore(ctx),
		k.cdc,
		func(keyArg []byte, item *types.SmartContractDeployment) bool {
			if item.ChainReferenceID == chainReferenceID {
				found = true
				return false
			}
			return true
		})
	return
}

func (k Keeper) RemoveSmartContractDeployment(ctx sdk.Context, smartContractID uint64, chainReferenceID string) {
	_, key := k.getSmartContractDeploying(ctx, smartContractID, chainReferenceID)
	if key == nil {
		return
	}
	k.Logger(ctx).Info("removing a smart contract deployment", "smart-contract-id", smartContractID, "chain-reference-id", chainReferenceID)
	k.smartContractDeploymentStore(ctx).Delete(key)
}

var lastSmartContractKey = []byte{0x1}

func (k Keeper) lastSmartContractStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("latest-smart-contract"))
}

func (k Keeper) getSmartContract(ctx sdk.Context, id uint64) (*types.SmartContract, error) {
	return keeperutil.Load[*types.SmartContract](k.smartContractsStore(ctx), k.cdc, keeperutil.Uint64ToByte(id))
}

func (k Keeper) saveSmartContract(ctx sdk.Context, smartContract *types.SmartContract) error {
	return keeperutil.Save(k.smartContractsStore(ctx), k.cdc, keeperutil.Uint64ToByte(smartContract.GetId()), smartContract)
}

func (k Keeper) setAsLastSmartContract(ctx sdk.Context, smartContract *types.SmartContract) error {
	kv := k.lastSmartContractStore(ctx)
	kv.Set(lastSmartContractKey, keeperutil.Uint64ToByte(smartContract.GetId()))
	return nil
}

func (k Keeper) GetLastSmartContract(ctx sdk.Context) (*types.SmartContract, error) {
	kv := k.lastSmartContractStore(ctx)
	id := kv.Get(lastSmartContractKey)
	return keeperutil.Load[*types.SmartContract](k.smartContractsStore(ctx), k.cdc, id)
}

func (k Keeper) OnSnapshotBuilt(ctx sdk.Context, snapshot *valsettypes.Snapshot) {
	chainInfos, err := k.GetAllChainInfos(ctx)
	if err != nil {
		panic(err)
	}
	for _, chain := range chainInfos {
		valset := transformSnapshotToCompass(snapshot, chain.GetChainReferenceID())
		if !chain.IsActive() {
			k.Logger(ctx).Info("ignoring valset for chain as the chain is not yet active",
				"chain-reference-id", chain.GetChainReferenceID(),
				"valset-id", valset.GetValsetID(),
			)
			continue
		}

		if !isEnoughToReachConsensus(valset) {
			k.Logger(ctx).Info("ignoring valset for chain as there isn't enough validators to form a consensus for this chain",
				"chain-reference-id", chain.GetChainReferenceID(),
				"valset-id", valset.GetValsetID(),
			)
			continue
		}

		k.Logger(ctx).Info("snapshot was built and a new update valset message is being sent over",
			"chain-reference-id", chain.GetChainReferenceID(),
			"valset-id", valset.GetValsetID(),
		)

		// clear all previous instances of the update valset from the queue
		k.Logger(ctx).Debug("clearing previous instances of the update valset from the queue")
		queueName := consensustypes.Queue(ConsensusTurnstoneMessage, xchainType, xchain.ReferenceID(chain.GetChainReferenceID()))
		messages, err := k.ConsensusKeeper.GetMessagesFromQueue(ctx, queueName, 999)
		if err != nil {
			k.Logger(ctx).Error("unable to get messages from queue", "err", err)
			continue
		}

		for _, msg := range messages {
			cmsg, err := msg.ConsensusMsg(k.cdc)
			if err != nil {
				k.Logger(ctx).Error("unable to unpack message", "err", err)
				continue
			}

			mmsg := cmsg.(*types.Message)
			act := mmsg.GetAction()
			if mmsg.GetTurnstoneID() != string(chain.GetSmartContractUniqueID()) {
				continue
			}
			if _, ok := act.(*types.Message_UpdateValset); ok {
				err := k.ConsensusKeeper.DeleteJob(ctx, queueName, msg.GetId())
				if err != nil {
					k.Logger(ctx).Error("unable to delete message", "err", err)
					continue
				}
			}
		}

		// put update valset message into the queue
		k.ConsensusKeeper.PutMessageInQueue(
			ctx,
			consensustypes.Queue(ConsensusTurnstoneMessage, xchainType, xchain.ReferenceID(chain.GetChainReferenceID())),
			&types.Message{
				TurnstoneID:      string(chain.GetSmartContractUniqueID()),
				ChainReferenceID: chain.GetChainReferenceID(),
				Action: &types.Message_UpdateValset{
					UpdateValset: &types.UpdateValset{
						Valset: &valset,
					},
				},
			}, nil)
	}

	k.TryDeployingLastSmartContractToAllChains(ctx)
}

func (k Keeper) CheckExternalBalancesForChain(ctx sdk.Context, chainReferenceID string) error {
	snapshot, err := k.Valset.GetCurrentSnapshot(ctx)
	if err != nil {
		return err
	}

	var msg types.ValidatorBalancesAttestation
	msg.FromBlockTime = ctx.BlockTime().UTC()

	for _, val := range snapshot.GetValidators() {
		for _, ext := range val.GetExternalChainInfos() {
			if ext.GetChainReferenceID() == chainReferenceID && ext.GetChainType() == "evm" {
				msg.ValAddresses = append(msg.ValAddresses, val.GetAddress())
				msg.HexAddresses = append(msg.HexAddresses, ext.GetAddress())
			}
		}
	}

	if len(msg.ValAddresses) == 0 {
		return nil
	}
	return k.ConsensusKeeper.PutMessageInQueue(
		ctx,
		consensustypes.Queue(ConsensusGetValidatorBalances, xchainType, chainReferenceID),
		&msg,
		&consensus.PutOptions{
			RequireSignatures: false,
			PublicAccessData:  []byte{1}, // anything because pigeon cares if public access data exists to be able to provide evidence
		},
	)
}

func isEnoughToReachConsensus(val types.Valset) bool {
	var sum uint64
	for _, power := range val.Powers {
		sum += power
	}

	return sum >= thresholdForConsensus
}

func transformSnapshotToCompass(snapshot *valsettypes.Snapshot, chainReferenceID string) types.Valset {
	validators := make([]valsettypes.Validator, len(snapshot.GetValidators()))
	copy(validators, snapshot.GetValidators())

	sort.SliceStable(validators, func(i, j int) bool {
		// doing GTE because we want a reverse sort
		return validators[i].ShareCount.GTE(validators[j].ShareCount)
	})

	totalPowerInt := sdk.NewInt(0)
	for _, val := range validators {
		totalPowerInt = totalPowerInt.Add(val.ShareCount)
	}

	totalPower := totalPowerInt.Int64()

	valset := types.Valset{
		ValsetID: snapshot.GetId(),
	}

	for _, val := range validators {
		for _, ext := range val.GetExternalChainInfos() {
			if strings.ToLower(ext.GetChainType()) == xchainType && ext.GetChainReferenceID() == chainReferenceID {
				power := maxPower * (float64(val.ShareCount.Int64()) / float64(totalPower))

				valset.Validators = append(valset.Validators, ext.Address)
				valset.Powers = append(valset.Powers, uint64(power))
			}
		}
	}

	return valset
}

func (k Keeper) ModuleName() string { return types.ModuleName }

func queueName(ref xchain.ReferenceID, subqueue string) string {
	return fmt.Sprintf("%s/%s/%s", xchainType, ref, subqueue)
}

func generateSmartContractID(ctx sdk.Context) (res [32]byte) {
	b := []byte(fmt.Sprintf("%d", ctx.BlockHeight()))
	copy(res[:], b)
	return
}
