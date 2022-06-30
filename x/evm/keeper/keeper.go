package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	wasmutil "github.com/palomachain/paloma/util/wasm"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/tendermint/tendermint/libs/log"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/x/evm/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

const (
	maxPower = 1 << 32
)
const (
	ConsensusTurnstoneMessage = "evm-turnstone-message"
	SignaturePrefix           = "\x19Ethereum Signed Message:\n32"
)

type supportedChainInfo struct {
	batch   bool
	msgType any
}

var SupportedConsensusQueues = map[string]supportedChainInfo{
	ConsensusTurnstoneMessage: {
		batch:   false,
		msgType: &types.Message{},
	},
}

type evmChainTemp struct {
	chainID     string
	turnstoneID string
}

func (e evmChainTemp) ChainID() string {
	return e.chainID
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

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) AddSmartContractExecutionToConsensus(
	ctx sdk.Context,
	chainID,
	turnstoneID string,
	logicCall *types.SubmitLogicCall,
) error {
	return k.ConsensusKeeper.PutMessageForSigning(
		ctx,
		consensustypes.Queue(
			ConsensusTurnstoneMessage,
			consensustypes.ChainTypeEVM,
			chainID,
		),
		&types.Message{
			ChainID:     chainID,
			TurnstoneID: turnstoneID,
			Action: &types.Message_SubmitLogicCall{
				SubmitLogicCall: logicCall,
			},
		},
	)
}

func (k Keeper) DeployNewSmartContractFromProposal(
	ctx sdk.Context,
	proposal *types.DeployNewSmartContractProposal,
) error {
	chainInfo, err := k.getChainInfo(ctx, proposal.GetChainID())
	if err != nil {
		return err
	}

	if chainInfo.GetStatus() != types.ChainInfo_ACTIVE {
		return ErrChainNotActive.Format(proposal.GetChainID())
	}

	snapshot, err := k.Valset.GetCurrentSnapshot(ctx)

	if err != nil {
		return err
	}

	contractABI, err := abi.JSON(strings.NewReader(proposal.GetAbiJSON()))
	if err != nil {
		return err
	}

	smartContractID := generateSmartContractID(ctx)
	valset := transformSnapshotToTurnstoneValset(snapshot, proposal.GetChainID())

	if len(valset.Validators) == 0 {
		return ErrNotEnoughValidatorsForGivenChainID
	}

	// set the smart contract constructor arguments
	input, err := contractABI.Pack("", smartContractID, valset)
	if err != nil {
		return err

	}

	// TODO: do this for every chainInfo
	return k.ConsensusKeeper.PutMessageForSigning(
		ctx,
		consensustypes.Queue(
			ConsensusTurnstoneMessage,
			consensustypes.ChainTypeEVM,
			proposal.GetChainID(),
		),
		&types.Message{
			ChainID: proposal.GetChainID(),
			Action: &types.Message_UploadSmartContract{
				UploadSmartContract: &types.UploadSmartContract{
					Bytecode:         common.FromHex(proposal.GetBytecodeHex()),
					Abi:              proposal.GetAbiJSON(),
					ConstructorInput: input,
				},
			},
		},
	)

}

// {"target_contract_info":{"method":"foo","chain_id":"abc","compass_id":"abc","contract_address":"0xabc","smart_contract_abi":"abc"},"paloma_address":"paloma1sp6yeu2cdemlh0jpterpe3as9mvx36ck6ys0ce","eth_address":[0,0,0,0,0,0,0,0,0,0,0,0,22,248,182,92,183,148,210,0,134,193,229,48,158,88,192,76,57,198,237,233]}
type executeEVMFromCosmWasm struct {
	TargetContractInfo struct {
		Method               string `json:"method"`
		ChainID              string `json:"chain_id"`
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

		err = k.AddSmartContractExecutionToConsensus(ctx, executeMsg.TargetContractInfo.ChainID, executeMsg.TargetContractInfo.CompassID, &types.SubmitLogicCall{
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

func (k Keeper) SupportedQueues(ctx sdk.Context) (map[string]consensus.QueueOptions, error) {
	chains, err := k.getAllChainInfos(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]consensus.QueueOptions)

	for _, chainInfo := range chains {
		if !chainInfo.IsActive() {
			continue
		}
		for subQueue, queueInfo := range SupportedConsensusQueues {

			queue := fmt.Sprintf("EVM/%s/%s", chainInfo.ChainID, subQueue)
			res[queue] = *consensus.ApplyOpts(nil,
				consensus.WithChainInfo(consensustypes.ChainTypeEVM, chainInfo.ChainID),
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
		}
	}

	return res, nil
}

func (k Keeper) getAllChainInfos(ctx sdk.Context) ([]*types.ChainInfo, error) {
	_, all, err := keeperutil.IterAll[*types.ChainInfo](k.chainInfoStore(ctx), k.cdc)
	return all, err
}

func (k Keeper) getChainInfo(ctx sdk.Context, targetChainID string) (*types.ChainInfo, error) {
	res, err := keeperutil.Load[*types.ChainInfo](k.chainInfoStore(ctx), k.cdc, []byte(targetChainID))
	if errors.Is(err, keeperutil.ErrNotFound) {
		return nil, ErrChainNotFound.Format(targetChainID)
	}
	return res, nil
}

func (k Keeper) updateChainInfo(ctx sdk.Context, chainInfo *types.ChainInfo) error {
	return keeperutil.Save(k.chainInfoStore(ctx), k.cdc, []byte(chainInfo.GetChainID()), chainInfo)
}

func (k Keeper) AddSupportForNewChain(ctx sdk.Context, addChain *types.AddChainProposal) error {
	_, err := k.getChainInfo(ctx, addChain.GetChainID())
	if err != nil {
		if !errors.Is(err, ErrChainNotFound) {
			return err
		}
	}
	chainInfo := &types.ChainInfo{
		ChainID:              addChain.GetChainID(),
		ReferenceBlockHeight: addChain.GetBlockHeight(),
		ReferenceBlockHash:   addChain.GetBlockHashAtHeight(),
		Status:               types.ChainInfo_WAITING_FOR_EVIDENCE,
	}
	return k.updateChainInfo(ctx, chainInfo)
}

func (k Keeper) ActivateChainID(ctx sdk.Context, chainID, smartContractAddr, smartContractID string) error {
	chainInfo, err := k.getChainInfo(ctx, chainID)
	if err != nil {
		return err
	}
	chainInfo.Status = types.ChainInfo_ACTIVE
	chainInfo.SmartContractAddr = smartContractAddr
	chainInfo.SmartContractID = smartContractID
	return k.updateChainInfo(ctx, chainInfo)
}

func (k Keeper) RemoveSupportForChain(ctx sdk.Context, proposal *types.RemoveChainProposal) error {
	_, err := k.getChainInfo(ctx, proposal.GetChainID())
	if err != nil {
		return err
	}

	k.chainInfoStore(ctx).Delete([]byte(proposal.GetChainID()))

	for subQueue := range SupportedConsensusQueues {
		queue := fmt.Sprintf("EVM/%s/%s", proposal.GetChainID(), subQueue)
		k.ConsensusKeeper.RemoveConsensusQueue(ctx, queue)
	}

	return nil
}

func (k Keeper) chainInfoStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("chain-info"))
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
		valset := transformSnapshotToTurnstoneValset(snapshot, chain.GetChainID())

		k.ConsensusKeeper.PutMessageForSigning(
			ctx,
			consensustypes.Queue(ConsensusTurnstoneMessage, consensustypes.ChainTypeEVM, chain.GetChainID()),
			&types.Message{
				TurnstoneID: chain.GetSmartContractID(),
				ChainID:     chain.GetChainID(),
				Action: &types.Message_UpdateValset{
					UpdateValset: &types.UpdateValset{
						Valset: &valset,
					},
				},
			},
		)
	}
}

func transformSnapshotToTurnstoneValset(snapshot *valsettypes.Snapshot, chainID string) types.Valset {
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
			if ext.GetChainID() == chainID {
				power := maxPower * (float64(val.ShareCount.Int64()) / float64(totalPower))

				valset.Validators = append(valset.Validators, ext.Address)
				valset.Powers = append(valset.Powers, uint64(power))
			}
		}
	}

	return valset
}

func generateSmartContractID(ctx sdk.Context) (res [32]byte) {
	hs := ctx.HeaderHash()[:32].String()
	copy(res[:], []byte(hs))
	return
}
