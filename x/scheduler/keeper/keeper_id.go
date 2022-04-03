package keeper

// A helper methods to help with incrementing and getting the last IDs for a given name.

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ider struct {
	sg    storeGetter
	idKey []byte
}

// getLastID returns the last id that was inserted for the given name.
// If one does not exist, then it returns 0,
func (i ider) getLastID(ctx sdk.Context, name string) uint64 {
	store := i.sg.Store(ctx)
	prefixKey := i.nextKeyPrefix([]byte(name))
	if !store.Has(prefixKey) {
		return 0
	}

	valB := store.Get(prefixKey)
	return binary.BigEndian.Uint64(valB)
}

// incrementNextID returns new ID which can now be used for referencing data
// and increments the counter internally. It returns the newly inserted ID.
func (i ider) incrementNextID(ctx sdk.Context, name string) uint64 {
	store := i.sg.Store(ctx)

	nextID := i.getLastID(ctx, name) + 1

	prefixKey := i.nextKeyPrefix([]byte(name))
	store.Set(prefixKey, uint64ToByte(nextID))

	return nextID
}

func uint64ToByte(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

func (i ider) nextKeyPrefix(key []byte) []byte {
	return append(i.idKey[:], key...)
}
