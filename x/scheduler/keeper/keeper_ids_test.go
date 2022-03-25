package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStoreIDs tests relevant methods that are used to increment counter IDs
// and fetching the last incremented values for a given name.
func TestStoreIDs(t *testing.T) {
	t.Parallel()
	const keyName = "random-key"

	t.Run("generating next increases the ID by one", func(t *testing.T) {
		k, ctx := newSchedulerKeeper(t)
		assert.Equal(t, uint64(1), k.incrementNextID(ctx, keyName))
		assert.Equal(t, uint64(2), k.incrementNextID(ctx, keyName))
		assert.Equal(t, uint64(3), k.incrementNextID(ctx, keyName))
	})

	t.Run("calling lastID with empty storage returns zero", func(t *testing.T) {
		k, ctx := newSchedulerKeeper(t)
		assert.Equal(t, uint64(0), k.getLastID(ctx, keyName))
	})

	t.Run("calling lastID after incrementing returns the same value that was inserted", func(t *testing.T) {
		k, ctx := newSchedulerKeeper(t)
		assert.Equal(t, uint64(1), k.incrementNextID(ctx, keyName))
		assert.Equal(t, uint64(1), k.getLastID(ctx, keyName))
	})

	t.Run("test that ID names are not coupled", func(t *testing.T) {
		k, ctx := newSchedulerKeeper(t)
		const name1, name2 = "name-1", "name-2"
		// increse it 3 times for name1
		k.incrementNextID(ctx, name1)
		k.incrementNextID(ctx, name1)
		k.incrementNextID(ctx, name1)

		// and increase it once for name2
		k.incrementNextID(ctx, name2)

		assert.Equal(t, uint64(3), k.getLastID(ctx, name1))
		assert.Equal(t, uint64(1), k.getLastID(ctx, name2))
	})

}
