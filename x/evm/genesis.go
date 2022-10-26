package evm

import (
	"errors"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/evm/keeper"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/vizualni/whoops"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {

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
	}

	sc := genState.GetSmartContract()
	if sc != nil {
		b := common.FromHex(sc.GetBytecodeHex())
		_, err := k.SaveNewSmartContract(ctx, sc.GetAbiJson(), b)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
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
		})
	}
	genesis.Chains = genesisChainInfos

	sc, err := k.GetLastSmartContract(ctx)
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
