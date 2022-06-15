package keeper

import (
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/palomachain/paloma/x/evm/types/turnstone"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

const (
	ConsensusArbitraryContractCall = consensustypes.ConsensusQueueType("evm-arbitrary-smart-contract-call")
	ConsensusTurnstoneMessage      = consensustypes.ConsensusQueueType("evm-turnstone-message")
)

type evmChainTemp struct {
	chainID     string
	turnstoneID uint64
}

var supportedChainIDs = []evmChainTemp{
	{"eth-main", 0xabab},
	{"ropsten", 0xaaaa},
}

var _ valsettypes.OnSnapshotBuiltListener = Keeper{}

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        sdk.StoreKey
	memKey          sdk.StoreKey
	paramstore      paramtypes.Subspace
	consensusKeeper types.ConsensusKeeper

	Valset types.ValsetKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	consensusKeeper types.ConsensusKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		memKey:          memKey,
		paramstore:      ps,
		consensusKeeper: consensusKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) AddSmartContractExecutionToConsensus(
	ctx sdk.Context,
	chainType string,
	chainID string,
	msg *types.ArbitrarySmartContractCall,
) error {
	return k.consensusKeeper.PutMessageForSigning(
		ctx,
		consensustypes.Queue(ConsensusArbitraryContractCall, chainType, chainID),
		msg,
	)
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
		recoveredPk, err := crypto.Ecrecover(bz, sig)
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

	for _, chain := range supportedChainIDs {
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
			consensus.WithStaticTypeCheck(&turnstone.Message{}),
			consensus.WithBytesToSignCalc(
				consensustypes.TypedBytesToSign(func(msg *turnstone.Message, salt consensustypes.Salt) []byte {
					return msg.Keccak256(salt.Nonce)
				}),
			),
			consensus.WithVerifySignature(ethVerifySig),
		)
	}

}

const (
	maxPower = 1 << 32
)

func (k Keeper) OnSnapshotBuilt(ctx sdk.Context, snapshot *valsettypes.Snapshot) {
	for _, chain := range supportedChainIDs {
		valset := transformSnapshotToTurnstoneValset(snapshot, chain.chainID)

		k.consensusKeeper.PutMessageForSigning(
			ctx,
			consensustypes.Queue(ConsensusTurnstoneMessage, consensustypes.ChainTypeEVM, chain.chainID),
			&turnstone.Message{
				TurnstoneID: chain.turnstoneID,
				ChainID:     chain.chainID,
				Action: &turnstone.Message_UpdateValset{
					UpdateValset: &turnstone.UpdateValset{
						Valset: &valset,
					},
				},
			},
		)
	}
}

func transformSnapshotToTurnstoneValset(snapshot *valsettypes.Snapshot, chainID string) turnstone.Valset {
	validators := make([]valsettypes.Validator, len(snapshot.GetValidators()))
	copy(validators, snapshot.GetValidators())

	sort.SliceStable(validators, func(i, j int) bool {
		// doing GTE because we want a reverse sort
		return validators[i].ShareCount.GTE(validators[j].ShareCount)
	})

	var totalPowerInt sdk.Int
	for _, val := range validators {
		totalPowerInt = totalPowerInt.Add(val.ShareCount)
	}

	totalPower := totalPowerInt.Int64()

	valset := turnstone.Valset{}

	for _, val := range snapshot.GetValidators() {
		for _, ext := range val.GetExternalChainInfos() {
			if ext.GetChainID() == chainID {
				power := uint32(maxPower * (val.ShareCount.Int64() / totalPower))

				valset.HexAddress = append(valset.HexAddress, ext.Address)
				valset.Powers = append(valset.Powers, power)
			}
		}
	}

	return valset
}
