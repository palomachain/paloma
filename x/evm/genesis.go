package evm

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/VolumeFi/whoops"
	"github.com/ethereum/go-ethereum/common"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/evm/keeper"
	"github.com/palomachain/paloma/x/evm/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, chainInfo := range genState.GetChains() {
		if chainInfo.GetMinOnChainBalance() == "" {
			panic("minimum on-chain balance is not a valid number")
		}
		balance, ok := new(big.Int).SetString(chainInfo.GetMinOnChainBalance(), 10)
		if !ok {
			panic("cannot parse balance " + chainInfo.GetMinOnChainBalance())
		}
		err := k.AddSupportForNewChain(
			ctx,
			chainInfo.GetChainReferenceID(),
			chainInfo.GetChainID(),
			chainInfo.GetBlockHeight(),
			chainInfo.GetBlockHashAtHeight(),
			balance,
		)
		if err != nil {
			panic(err)
		}

		err = k.SetRelayWeights(
			ctx,
			chainInfo.GetChainReferenceID(),
			&types.RelayWeights{
				Fee:           "1.0",
				Uptime:        "1.0",
				SuccessRate:   "1.0",
				ExecutionTime: "1.0",
			},
		)
		if err != nil {
			panic(err)
		}
	}

	sc := genState.GetSmartContract()
	if sc != nil {
		b := common.FromHex(sc.GetBytecodeHex())
		nsc, err := k.SaveNewSmartContract(ctx, sc.GetAbiJson(), b)
		if err != nil {
			panic(fmt.Errorf("failed to save new compass contract: %w", err))
		}
		if err := k.SetAsCompassContract(ctx, nsc); err != nil {
			panic(fmt.Errorf("failed to set as compass contract: %w", err))
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	var genesisChainInfos []*types.GenesisChainInfo

	for _, chainInfo := range whoops.Must(k.GetAllChainInfos(ctx)) {
		if !chainInfo.IsActive() {
			continue
		}
		genesisChainInfos = append(genesisChainInfos, &types.GenesisChainInfo{
			ChainReferenceID:  chainInfo.GetChainReferenceID(),
			ChainID:           chainInfo.GetChainID(),
			BlockHeight:       chainInfo.GetReferenceBlockHeight(),
			BlockHashAtHeight: chainInfo.GetReferenceBlockHash(),
			MinOnChainBalance: whoops.Must(chainInfo.GetMinOnChainBalanceBigInt()).Text(10),
			RelayWeights:      whoops.Must(k.GetRelayWeights(ctx, chainInfo.GetChainReferenceID())),
		})
	}
	genesis.Chains = genesisChainInfos

	sc, err := k.GetLastCompassContract(ctx)
	switch {
	case err == nil:
		genesis.SmartContract = &types.GenesisSmartContract{
			AbiJson:     sc.GetAbiJSON(),
			BytecodeHex: "0x" + common.Bytes2Hex(sc.GetBytecode()),
		}
	case errors.Is(err, keeperutil.ErrNotFound):
		// do nothing
	default:
		panic(err)

	}

	return genesis
}
