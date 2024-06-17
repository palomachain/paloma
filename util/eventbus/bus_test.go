package eventbus_test

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/eventbus"
	"github.com/stretchr/testify/require"
)

func TestEventBus(t *testing.T) {
	ctx := sdk.Context{}.
		WithLogger(log.NewNopLogger()).
		WithContext(context.Background())
	eventbus.GravityBatchBuilt().Publish(ctx, eventbus.GravityBatchBuiltEvent{
		ChainReferenceID: "test-chain",
	})

	calls := make(map[string]int)
	fn := func(_ context.Context, e eventbus.GravityBatchBuiltEvent) error {
		calls[e.ChainReferenceID]++
		return nil
	}

	eventbus.GravityBatchBuilt().Subscribe("test-1", fn)
	require.Len(t, calls, 0, "should be empty")

	eventbus.GravityBatchBuilt().Publish(ctx, eventbus.GravityBatchBuiltEvent{
		ChainReferenceID: "test-chain",
	})

	require.NotNil(t, calls["test-chain"], "should have executed one.")
	require.Equal(t, 1, calls["test-chain"], "should have executed one.")

	eventbus.GravityBatchBuilt().Subscribe("test-2", fn)
	eventbus.GravityBatchBuilt().Publish(ctx, eventbus.GravityBatchBuiltEvent{
		ChainReferenceID: "test-chain",
	})
	require.Equal(t, 3, calls["test-chain"], "should execute both subscribers.")

	eventbus.GravityBatchBuilt().Unsubscribe("test-1")
	eventbus.GravityBatchBuilt().Publish(ctx, eventbus.GravityBatchBuiltEvent{
		ChainReferenceID: "test-chain",
	})
	require.Equal(t, 4, calls["test-chain"], "should have removed one subscriber.")
}
