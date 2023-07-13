package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// for the moment this is only testing delegate keys being set, but it would be good to make
// this test more expansive
func TestExportAndImport(t *testing.T) {
	env := CreateTestEnv(t)
	ctx := env.Context
	keeper := env.GravityKeeper

	valAddr, _ := sdk.ValAddressFromBech32("cosmosvaloper13yfm8as7y0mzsxqkfmk5jvgm45aez0u24jk95z")
	orchAddr, _ := sdk.AccAddressFromBech32("cosmos1h706wwrghfpydyh735aet8aluhf95dqj0psgyf")
	ethAddr := common.BytesToAddress([]byte("0xFDb0aaBD40774BBF3068Bf29E8b0a6C88BE26F83"))

	keeper.setValidatorEthereumAddress(ctx, valAddr, ethAddr)
	keeper.setEthereumOrchestratorAddress(ctx, ethAddr, orchAddr)
	keeper.SetOrchestratorValidatorAddress(ctx, valAddr, orchAddr)

	exportedGenesis := ExportGenesis(ctx, keeper)
	newEnv := CreateTestEnv(t)
	newCtx := newEnv.Context
	newKeeper := newEnv.GravityKeeper

	InitGenesis(newCtx, newKeeper, exportedGenesis)

	assert.Equal(t, newKeeper.GetValidatorEthereumAddress(newCtx, valAddr), ethAddr)
	assert.Equal(t, newKeeper.GetEthereumOrchestratorAddress(newCtx, ethAddr), orchAddr)
	assert.Equal(t, newKeeper.GetOrchestratorValidatorAddress(newCtx, orchAddr), valAddr)
}
