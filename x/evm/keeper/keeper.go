package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/slice"
	wasmutil "github.com/palomachain/paloma/util/wasm"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
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
	ConsensusTurnstoneMessage = "evm-turnstone-message"
	SignaturePrefix           = "\x19Ethereum Signed Message:\n32"
)

type supportedChainInfo struct {
	batch       bool
	msgType     any
	processFunc func(Keeper) func(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) error
}

var SupportedConsensusQueues = map[string]supportedChainInfo{
	ConsensusTurnstoneMessage: {
		batch:   false,
		msgType: &types.Message{},
		processFunc: func(k Keeper) func(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) error {
			return k.attestRouter(ctx, q, msg)
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

var zero32Byte [32]byte

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
	return k.ConsensusKeeper.PutMessageForSigning(
		ctx,
		consensustypes.Queue(
			ConsensusTurnstoneMessage,
			consensustypes.ChainTypeEVM,
			chainReferenceID,
		),
		&types.Message{
			ChainReferenceID: chainReferenceID,
			TurnstoneID:      turnstoneID,
			Action: &types.Message_SubmitLogicCall{
				SubmitLogicCall: logicCall,
			},
		},
	)
}

func (k Keeper) deploySmartContractToChain(ctx sdk.Context, chainReferenceID string, smartContract *types.SmartContract) error {
	contractABI, err := abi.JSON(strings.NewReader(smartContract.GetAbiJSON()))
	if err != nil {
		return err
	}

	chainInfo, err := k.GetChainInfo(ctx, chainReferenceID)
	if err != nil {
		return err
	}

	if chainInfo.GetStatus() == types.ChainInfo_IN_PROPOSAL {
		k.Logger(ctx).Info("skipping chain as it's in proposal", "chain-id", chainReferenceID)
		return nil
	}

	snapshot, err := k.Valset.GetCurrentSnapshot(ctx)
	if err != nil {
		return err
	}
	smartContractID := generateSmartContractID(ctx)
	valset := transformSnapshotToCompass(snapshot, chainReferenceID)

	if !isEnoughToReachConsensus(valset) {
		k.Logger(ctx).Info("skipping as there are not enough validators", "chain-id", chainInfo.GetChainReferenceID())
		return nil
	}

	// set the smart contract constructor arguments
	input, err := contractABI.Pack("", smartContractID, transformValsetToABIValset(valset))
	if err != nil {
		return err

	}

	return k.ConsensusKeeper.PutMessageForSigning(
		ctx,
		consensustypes.Queue(
			ConsensusTurnstoneMessage,
			consensustypes.ChainTypeEVM,
			chainReferenceID,
		),
		&types.Message{
			ChainReferenceID: chainReferenceID,
			Action: &types.Message_UploadSmartContract{
				UploadSmartContract: &types.UploadSmartContract{
					Id:               smartContract.GetId(),
					Bytecode:         smartContract.GetBytecode(),
					Abi:              smartContract.GetAbiJSON(),
					ConstructorInput: input,
				},
			},
		},
	)
}

func (k Keeper) UpdateWithSmartContract(ctx sdk.Context, abiJSON string, bytecode []byte) error {
	ctx, write := ctx.CacheContext()

	smartContract := &types.SmartContract{
		Id:       k.ider.IncrementNextID(ctx, "smart-contract"),
		AbiJSON:  abiJSON,
		Bytecode: bytecode,
	}

	err = keeperutil.Save(k.smartContractsStore(ctx), k.cdc, keeperutil.Uint64ToByte(smartContract.GetId()), smartContract)
	if err != nil {
		return err
	}

	err = k.tryDeployingSmartContract(ctx, smartContract)
	if err != nil {
		return err
	}

	write()

	return nil
}

func (k Keeper) tryDeployingSmartContract(ctx sdk.Context, smartContract *types.SmartContract) error {
	var g whoops.Group
	chainInfos, err := k.getAllChainInfos(ctx)

	if err != nil {
		return err
	}

	for _, chainInfo := range chainInfos {
		if chainInfo.SmartContractVersion != smartContract.GetId() {
			g.Add(k.deploySmartContractToChain(ctx, chainInfo.GetChainReferenceID(), smartContract))
		}

	}

	if g.Err() {
		return g
	}

	return nil
}

// {"target_contract_info":{"method":"foo","chain_id":"abc","compass_id":"abc","contract_address":"0xabc","smart_contract_abi":"abc"},"paloma_address":"paloma1sp6yeu2cdemlh0jpterpe3as9mvx36ck6ys0ce","eth_address":[0,0,0,0,0,0,0,0,0,0,0,0,22,248,182,92,183,148,210,0,134,193,229,48,158,88,192,76,57,198,237,233]}
type executeEVMFromCosmWasm struct {
	TargetContractInfo struct {
		Method               string `json:"method"`
		ChainReferenceID     string `json:"chain_id"`
		SmartContractAddress string `json:"contract_address"`
		SmartContractABI     string `json:"smart_contract_abi"`

		CompassID string `json:"compass_id"`
	} `json:"target_contract_info"`

	// TODO: we need to have this as a payload
	Payload string `json:"payload"`
}

func (e executeEVMFromCosmWasm) valid() bool {
	zero := executeEVMFromCosmWasm{}
	if e == zero {
		return false
	}
	// todo: add more in the future
	return true
}

func (k Keeper) WasmMessengerHandler() wasmutil.MessengerFnc {
	return func(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
		var executeMsg executeEVMFromCosmWasm
		err := json.Unmarshal(msg.Custom, &executeMsg)
		if err != nil {
			return nil, nil, err
		}
		if !executeMsg.valid() {
			return nil, nil, wasmtypes.ErrUnknownMsg
		}

		err = k.AddSmartContractExecutionToConsensus(ctx, executeMsg.TargetContractInfo.ChainReferenceID, executeMsg.TargetContractInfo.CompassID, &types.SubmitLogicCall{
			HexContractAddress: executeMsg.TargetContractInfo.SmartContractAddress,
			Payload:            []byte(executeMsg.Payload),
			Deadline:           ctx.BlockTime().UTC().Add(5 * time.Minute).Unix(),
			Abi:                []byte(executeMsg.TargetContractInfo.SmartContractABI),
		})

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
	chains, err := k.getAllChainInfos(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]consensus.QueueOptions)

	for _, chainInfo := range chains {
		// if !chainInfo.IsActive() {
		// 	continue
		// }
		for subQueue, queueInfo := range SupportedConsensusQueues {

			queue := fmt.Sprintf("EVM/%s/%s", chainInfo.ChainReferenceID, subQueue)
			opts := *consensus.ApplyOpts(nil,
				consensus.WithChainInfo(consensustypes.ChainTypeEVM, chainInfo.ChainReferenceID),
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
				QueueOptions: opts,
				Process:      queueInfo.processFunc(k),
			}
		}
	}

	return res, nil
}

func (k Keeper) getAllChainInfos(ctx sdk.Context) ([]*types.ChainInfo, error) {
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

func (k Keeper) AddSupportForNewChain(ctx sdk.Context, addChain *types.AddChainProposal) error {
	_, err := k.GetChainInfo(ctx, addChain.GetChainReferenceID())
	if !errors.Is(err, ErrChainNotFound) {
		// we want chain not to exist when adding a new one!
		if err != nil {
			err = whoops.String("expected chain not to exist")
		}
		return whoops.Wrap(ErrUnexpectedError, err)
	}
	chainInfo := &types.ChainInfo{
		ChainID:              addChain.GetChainID(),
		ChainReferenceID:     addChain.GetChainReferenceID(),
		ReferenceBlockHeight: addChain.GetBlockHeight(),
		ReferenceBlockHash:   addChain.GetBlockHashAtHeight(),
		Status:               types.ChainInfo_WAITING_FOR_EVIDENCE,
	}
	return k.updateChainInfo(ctx, chainInfo)
}

func (k Keeper) ActivateChainReferenceID(ctx sdk.Context, chainReferenceID, smartContractAddr, smartContractID string) error {
	chainInfo, err := k.GetChainInfo(ctx, chainReferenceID)
	if err != nil {
		return err
	}
	chainInfo.Status = types.ChainInfo_ACTIVE
	chainInfo.SmartContractAddr = smartContractAddr
	chainInfo.SmartContractID = smartContractID
	return k.updateChainInfo(ctx, chainInfo)
}

func (k Keeper) RemoveSupportForChain(ctx sdk.Context, proposal *types.RemoveChainProposal) error {
	_, err := k.GetChainInfo(ctx, proposal.GetChainReferenceID())
	if err != nil {
		return err
	}

	k.chainInfoStore(ctx).Delete([]byte(proposal.GetChainReferenceID()))

	for subQueue := range SupportedConsensusQueues {
		queue := fmt.Sprintf("EVM/%s/%s", proposal.GetChainReferenceID(), subQueue)
		k.ConsensusKeeper.RemoveConsensusQueue(ctx, queue)
	}

	return nil
}

func (k Keeper) chainInfoStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("chain-info"))
}

func (k Keeper) smartContractsStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("smart-contracts"))
}

var lastSmartContractKey = []byte{0x1}

func (k Keeper) lastSmartContractStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("latest-smart-contract"))
}

func (k Keeper) getSmartContract(ctx sdk.Context, id uint64) (*types.SmartContract, error) {
	return keeperutil.Load[*types.SmartContract](k.smartContractsStore(ctx), k.cdc, keeperutil.Uint64ToByte(id))
}

func (k Keeper) getLastSmartContract(ctx sdk.Context) (*types.SmartContract, error) {
	return keeperutil.Load[*types.SmartContract](k.lastSmartContractStore(ctx), k.cdc, lastSmartContractKey)
}

func (k Keeper) OnSnapshotBuilt(ctx sdk.Context, snapshot *valsettypes.Snapshot) {
	chainInfos, err := k.getAllChainInfos(ctx)
	if err != nil {
		panic(err)
	}
	for _, chain := range chainInfos {
		if !chain.IsActive() {
			continue
		}
		valset := transformSnapshotToCompass(snapshot, chain.GetChainReferenceID())

		if !isEnoughToReachConsensus(valset) {
			continue
		}

		k.ConsensusKeeper.PutMessageForSigning(
			ctx,
			consensustypes.Queue(ConsensusTurnstoneMessage, consensustypes.ChainTypeEVM, chain.GetChainReferenceID()),
			&types.Message{
				TurnstoneID:      chain.GetSmartContractID(),
				ChainReferenceID: chain.GetChainReferenceID(),
				Action: &types.Message_UpdateValset{
					UpdateValset: &types.UpdateValset{
						Valset: &valset,
					},
				},
			},
		)
	}

	// given that valset was changes, there still might be a chainReferenceID that had
	// zero validators in the valset. This tries to update the state for those
	// smart contracts to get them up online.
	k.tryDeployingSmartContract(ctx)
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
			if ext.GetChainReferenceID() == chainReferenceID {
				power := maxPower * (float64(val.ShareCount.Int64()) / float64(totalPower))

				valset.Validators = append(valset.Validators, ext.Address)
				valset.Powers = append(valset.Powers, uint64(power))
			}
		}
	}

	return valset
}

func transformValsetToABIValset(val types.Valset) any {
	return struct {
		Validators []common.Address
		Powers     []*big.Int
		ValsetId   *big.Int
	}{
		Validators: slice.Map(val.GetValidators(), func(s string) common.Address {
			return common.HexToAddress(s)
		}),
		Powers: slice.Map(val.GetPowers(), func(p uint64) *big.Int {
			return big.NewInt(int64(p))
		}),
		ValsetId: big.NewInt(int64(val.GetValsetID())),
	}

}

func generateSmartContractID(ctx sdk.Context) (res [32]byte) {
	heightstr := strconv.Itoa(int(ctx.BlockHeight()))
	copy(res[:], []byte(heightstr))
	return
}
