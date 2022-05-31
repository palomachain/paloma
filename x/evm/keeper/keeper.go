package keeper

import (
	"fmt"

	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/palomachain/paloma/x/evm/types"
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

func (k Keeper) AddSmartContractExecutionToConsensus(ctx sdk.Context, msg *types.ArbitrarySmartContractCall) error {
	return k.consensusKeeper.PutMessageForSigning(ctx, "eth-sc-execution", msg)
}

func (k Keeper) OnSchedulerMessageProcess(ctx sdk.Context, rawMsg any) (processed bool, err error) {
	// when scheduler ticks then this gets executed

	processed = true
	switch msg := rawMsg.(type) {
	case *types.ArbitrarySmartContractCall:
		err = k.AddSmartContractExecutionToConsensus(ctx, msg)
	default:
		processed = false
	}

	return
}

func (k Keeper) SupportedConsensusQueues(consensusKeeper types.ConsensusKeeper) {
	consensusKeeper.AddConcencusQueueType(
		consensustypes.ConsensusQueueType("bla"),
		&types.ArbitrarySmartContractCall{},
		consensustypes.TypedBytesToSign(func(msg *types.ArbitrarySmartContractCall, nonce uint64) []byte {
			return msg.Keccak256(nonce)
		}),
	)
}
