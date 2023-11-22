package keeper

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func newCtx(store storetypes.MultiStore) sdk.Context {
	return sdk.NewContext(store, tmproto.Header{}, false, nil)
}

// TestStoreIDs tests relevant methods that are used to increment counter IDs
// and fetching the last incremented values for a given name.
func TestStoreIDs(t *testing.T) {
	const keyName = "random-key"

	prepareTest := func() (IDGenerator, sdk.Context) {
		store, kv, _ := SampleStore("a", "b")
		sg := SimpleStoreGetter(store.GetKVStore(kv))
		ctx := newCtx(store)
		return NewIDGenerator(sg, nil), ctx
	}

	t.Run("generating next increases the ID by one", func(t *testing.T) {
		ider, ctx := prepareTest()
		assert.Equal(t, uint64(1), ider.IncrementNextID(ctx, keyName))
		assert.Equal(t, uint64(2), ider.IncrementNextID(ctx, keyName))
		assert.Equal(t, uint64(3), ider.IncrementNextID(ctx, keyName))
	})

	t.Run("calling lastID with empty storage returns zero", func(t *testing.T) {
		ider, ctx := prepareTest()
		assert.Equal(t, uint64(0), ider.GetLastID(ctx, keyName))
	})

	t.Run("calling lastID after incrementing returns the same value that was inserted", func(t *testing.T) {
		ider, ctx := prepareTest()
		assert.Equal(t, uint64(1), ider.IncrementNextID(ctx, keyName))
		assert.Equal(t, uint64(1), ider.GetLastID(ctx, keyName))
	})

	t.Run("test that ID names are not coupled", func(t *testing.T) {
		ider, ctx := prepareTest()
		const name1, name2 = "name-1", "name-2"
		// increse it 3 times for name1
		ider.IncrementNextID(ctx, name1)
		ider.IncrementNextID(ctx, name1)
		ider.IncrementNextID(ctx, name1)

		// and increase it once for name2
		ider.IncrementNextID(ctx, name2)

		assert.Equal(t, uint64(3), ider.GetLastID(ctx, name1))
		assert.Equal(t, uint64(1), ider.GetLastID(ctx, name2))
	})
}
