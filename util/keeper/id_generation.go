package keeper

// A helper methods to help with incrementing and getting the last IDs for a given name.

import (
	"context"
	"encoding/binary"
)

const (
	defaultIDKey = "generated-ids-"
)

type IDGenerator struct {
	sg    StoreGetter
	idKey []byte
}

// NewIDGenerator creates a new ID generator. IT uses idKey as a main
func NewIDGenerator(sg StoreGetter, idkey []byte) IDGenerator {
	if idkey == nil {
		idkey = []byte(defaultIDKey)
	}
	return IDGenerator{
		sg:    sg,
		idKey: idkey,
	}
}

// getLastID returns the last id that was inserted for the given name.
// If one does not exist, then it returns 0,
func (i IDGenerator) GetLastID(ctx context.Context, name string) uint64 {
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
func (i IDGenerator) IncrementNextID(ctx context.Context, name string) uint64 {
	store := i.sg.Store(ctx)

	nextID := i.GetLastID(ctx, name) + 1

	prefixKey := i.nextKeyPrefix([]byte(name))
	store.Set(prefixKey, Uint64ToByte(nextID))

	return nextID
}

func Uint64ToByte(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

func (i IDGenerator) nextKeyPrefix(key []byte) []byte {
	return append(i.idKey[:], key...)
}

func (i IDGenerator) Zero() bool { return i.idKey == nil && i.sg == nil }
