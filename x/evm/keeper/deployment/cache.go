package deployment

import (
	"context"
)

// Cache is meant as an asynchronous KV memcache to reduce
// constant keeper queries. It is NOT threadsafe, it requires
// manual bootstrapping by loading data from the keeper once
// during initialisation and needs to be kept in sync
// with added and removed keeper data manually.
//
// If this works well, I'd like to build this into a v2
// that functions basically like a keeper wrapper, agnostic
// of the underlaying data and keeping in sync with added
// and removed items automatically.
type Cache struct {
	data        map[string]map[uint64]uint64
	bootstrap   func(context.Context, *Cache)
	initialised bool
}

func NewCache(bootstrap func(context.Context, *Cache)) *Cache {
	return &Cache{
		data:      make(map[string]map[uint64]uint64),
		bootstrap: bootstrap,
	}
}

func (c *Cache) Add(ctx context.Context, chainReferenceID string, smartContractID uint64, msgIDs ...uint64) {
	c.assertExists(ctx, chainReferenceID)

	for _, v := range msgIDs {
		c.data[chainReferenceID][v] = smartContractID
	}
}

func (c *Cache) Delete(ctx context.Context, chainReferenceID string, msgIDs ...uint64) {
	c.assertExists(ctx, chainReferenceID)

	for _, v := range msgIDs {
		delete(c.data[chainReferenceID], v)
	}
}

func (c *Cache) Has(ctx context.Context, chainReferenceID string, msgID uint64) bool {
	c.assertExists(ctx, chainReferenceID)

	_, found := c.data[chainReferenceID][msgID]
	return found
}

func (c *Cache) Get(ctx context.Context, chainReferenceID string, msgID uint64) uint64 {
	c.assertExists(ctx, chainReferenceID)

	return c.data[chainReferenceID][msgID]
}

func (c *Cache) assertExists(ctx context.Context, chainReferenceID string) {
	c.assertInitialised(ctx)

	if _, found := c.data[chainReferenceID]; !found {
		c.data[chainReferenceID] = make(map[uint64]uint64)
	}
}

func (c *Cache) assertInitialised(ctx context.Context) {
	if c.initialised {
		return
	}

	c.initialised = true
	c.bootstrap(ctx, c)
}
