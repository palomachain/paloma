package keeper

// A helper methods to help with incrementing and getting the last IDs for a given name.

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// getLastID returns the last id that was inserted for the given name.
// If one does not exist, then it returns 0,
func (k Keeper) getLastID(ctx sdk.Context, name string) uint64 {
	store := k.store(ctx)
	prefixKey := nextKeyPrefix([]byte(name))
	if !store.Has(prefixKey) {
		return 0
	}

	valB := store.Get(prefixKey)
	return binary.BigEndian.Uint64(valB)
}

// incrementNextID returns new ID which can now be used for referencing data
// and increments the counter internally. It returns the newly inserted ID.
func (k Keeper) incrementNextID(ctx sdk.Context, name string) uint64 {
	store := k.store(ctx)

	nextID := k.getLastID(ctx, name) + 1

	prefixKey := nextKeyPrefix([]byte(name))
	store.Set(prefixKey, uint64ToByte(nextID))

	return nextID
}

func uint64ToByte(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

func nextKeyPrefix(key []byte) []byte {
	nextIDKey := []byte("id-counter-")
	return append(nextIDKey, key...)
}
