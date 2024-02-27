package v1_test

import (
	"testing"

	st "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/exported"
	"github.com/palomachain/paloma/x/gravity/keeper"
	v1 "github.com/palomachain/paloma/x/gravity/migrations/v1"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/require"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps exported.ParamSet) {
	*ps.(*types.Params) = ms.ps
}

func TestMigration(t *testing.T) {
	encCfg := keeper.MakeTestEncodingConfig()
	cdc := encCfg.Codec
	storeKey := st.NewKVStoreKey(types.ModuleName)
	tkey := st.NewTransientStoreKey("test")
	ctx := testutil.DefaultContext(storeKey, tkey)

	b := types.DefaultParams()
	store := ctx.KVStore(storeKey)
	legacySubspace := newMockSubspace(*b)
	require.NoError(t, v1.MigrateParams(ctx, store, legacySubspace, cdc))

	var res types.Params
	bz := store.Get(types.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &res))
	require.Equal(t, legacySubspace.ps, res)
}
