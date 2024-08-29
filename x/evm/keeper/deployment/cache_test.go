package deployment_test

import (
	"context"
	"testing"

	"github.com/palomachain/paloma/x/evm/keeper/deployment"
	"github.com/stretchr/testify/assert"
)

var (
	dummyBootstrapper = func(_ context.Context, _ *deployment.Cache) {}
	ctx               = context.Background()
)

func TestDeploymentCache_Add(t *testing.T) {
	cache := deployment.NewCache(dummyBootstrapper)

	// Add a message ID to a chain reference ID
	cache.Add(ctx, "chain1", 1, 5)

	// Check if the message ID exists in the chain reference ID
	assert.True(t, cache.Has(ctx, "chain1", 5))
}

func TestDeploymentCache_Delete(t *testing.T) {
	cache := deployment.NewCache(dummyBootstrapper)

	// Add a message ID to a chain reference ID
	cache.Add(ctx, "chain1", 1, 5)

	// Delete the message ID from the chain reference ID
	cache.Delete(ctx, "chain1", 5)

	// Check if the message ID no longer exists in the chain reference ID
	assert.False(t, cache.Has(ctx, "chain1", 5))
}

func TestDeploymentCache_Has(t *testing.T) {
	cache := deployment.NewCache(dummyBootstrapper)

	// Add a message ID to a chain reference ID
	cache.Add(ctx, "chain1", 1, 5)

	// Check if the message ID exists in the chain reference ID
	assert.True(t, cache.Has(ctx, "chain1", 5))

	// Check if a non-existent message ID exists in the chain reference ID
	assert.False(t, cache.Has(ctx, "chain1", 1))
}

func TestCache_Get(t *testing.T) {
	// Create a new cache
	cache := deployment.NewCache(func(ctx context.Context, c *deployment.Cache) {
		c.Add(ctx, "chain1", 100, 1)
		c.Add(ctx, "chain1", 200, 2)

		c.Add(ctx, "chain2", 300, 1)
		c.Add(ctx, "chain2", 400, 2)
	})

	// Test case 1: Existing chainReferenceID and msgID
	expectedResult := uint64(200)
	result := cache.Get(ctx, "chain1", 2)
	if result != expectedResult {
		t.Errorf("Expected %d, but got %d", expectedResult, result)
	}

	// Test case 2: Non-existing chainReferenceID
	expectedResult = uint64(0)
	result = cache.Get(ctx, "chain3", 1)
	if result != expectedResult {
		t.Errorf("Expected %d, but got %d", expectedResult, result)
	}

	// Test case 3: Non-existing msgID
	expectedResult = uint64(0)
	result = cache.Get(ctx, "chain2", 3)
	if result != expectedResult {
		t.Errorf("Expected %d, but got %d", expectedResult, result)
	}
}
