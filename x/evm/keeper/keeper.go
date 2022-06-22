package keeper

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	wasmutil "github.com/palomachain/paloma/util/wasm"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/tendermint/tendermint/libs/log"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/x/evm/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

const (
	maxPower = 1 << 32
)
const (
	ConsensusArbitraryContractCall = consensustypes.ConsensusQueueType("evm-arbitrary-smart-contract-call")
	ConsensusTurnstoneMessage      = consensustypes.ConsensusQueueType("evm-turnstone-message")
	signaturePrefix                = "\x19Ethereum Signed Message:\n32"
)

type evmChainTemp struct {
	chainID     string
	turnstoneID string
}

func (e evmChainTemp) ChainID() string {
	return e.chainID
}

var zero32Byte [32]byte

var SupportedChainIDs = []evmChainTemp{
	{"eth-main", string(zero32Byte[:])},
	{"ropsten", string(zero32Byte[:])},
}

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

func (k Keeper) RegisterConsensusQueues(adder consensus.RegistryAdder) {
	ethVerifySig := func(bz []byte, sig []byte, address []byte) bool {
		receivedAddr := common.BytesToAddress(address)

		bytesToVerify := crypto.Keccak256(append(
			[]byte(signaturePrefix),
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
	}

	for _, chain := range SupportedChainIDs {
		adder.AddConcencusQueueType(
			false,
			consensus.WithChainInfo(consensustypes.ChainTypeEVM, chain.chainID),
			consensus.WithQueueTypeName(ConsensusArbitraryContractCall),
			consensus.WithStaticTypeCheck(&types.ArbitrarySmartContractCall{}),
			consensus.WithBytesToSignCalc(
				consensustypes.TypedBytesToSign(func(msg *types.ArbitrarySmartContractCall, salt consensustypes.Salt) []byte {
					return msg.Keccak256(salt.Nonce)
				}),
			),
			consensus.WithVerifySignature(ethVerifySig),
		)

		adder.AddConcencusQueueType(
			false,
			consensus.WithChainInfo(consensustypes.ChainTypeEVM, chain.chainID),
			consensus.WithQueueTypeName(ConsensusTurnstoneMessage),
			consensus.WithStaticTypeCheck(&types.Message{}),
			consensus.WithBytesToSignCalc(
				consensustypes.TypedBytesToSign(func(msg *types.Message, salt consensustypes.Salt) []byte {
					return msg.Keccak256(salt.Nonce)
				}),
			),
			consensus.WithVerifySignature(ethVerifySig),
		)
	}

}

func (k Keeper) OnSnapshotBuilt(ctx sdk.Context, snapshot *valsettypes.Snapshot) {
	for _, chain := range SupportedChainIDs {
		valset := transformSnapshotToTurnstoneValset(snapshot, chain.chainID)

		k.ConsensusKeeper.PutMessageForSigning(
			ctx,
			consensustypes.Queue(ConsensusTurnstoneMessage, consensustypes.ChainTypeEVM, chain.chainID),
			&types.Message{
				TurnstoneID: chain.turnstoneID,
				ChainID:     chain.chainID,
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
