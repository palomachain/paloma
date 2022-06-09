package keeper

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/x/evm/types"
)

const (
	ConsensusArbitraryContractCall = consensustypes.ConsensusQueueType("evm-arbitrary-smart-contract-call")
)

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        sdk.StoreKey
	memKey          sdk.StoreKey
	paramstore      paramtypes.Subspace
	consensusKeeper types.ConsensusKeeper
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
	adder.AddConcencusQueueType(
		false,
		consensus.WithChainInfo(consensustypes.ChainTypeEVM, "eth-main"),
		consensus.WithQueueTypeName(ConsensusArbitraryContractCall),
		consensus.WithStaticTypeCheck(&types.ArbitrarySmartContractCall{}),
		consensus.WithBytesToSignCalc(
			consensustypes.TypedBytesToSign(func(msg *types.ArbitrarySmartContractCall, salt consensustypes.Salt) []byte {
				return msg.Keccak256(salt.Nonce)
			}),
		),
		consensus.WithVerifySignature(func(bz []byte, sig []byte, address []byte) bool {
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
		}),
	)
}
