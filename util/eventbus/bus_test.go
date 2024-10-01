package eventbus_test

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/eventbus"
	"github.com/stretchr/testify/require"
)

func TestEventBus(t *testing.T) {
	ctx := sdk.Context{}.
		WithLogger(log.NewNopLogger()).
		WithContext(context.Background())
	eventbus.SkywayBatchBuilt().Publish(ctx, eventbus.SkywayBatchBuiltEvent{
		ChainReferenceID: "test-chain",
	})

	calls := make(map[string]int)
	fn := func(_ context.Context, e eventbus.SkywayBatchBuiltEvent) error {
		calls[e.ChainReferenceID]++
		return nil
	}

	eventbus.SkywayBatchBuilt().Subscribe("test-1", fn)
	require.Len(t, calls, 0, "should be empty")

	eventbus.SkywayBatchBuilt().Publish(ctx, eventbus.SkywayBatchBuiltEvent{
		ChainReferenceID: "test-chain",
	})

	require.NotNil(t, calls["test-chain"], "should have executed one.")
	require.Equal(t, 1, calls["test-chain"], "should have executed one.")

	eventbus.SkywayBatchBuilt().Subscribe("test-2", fn)
	eventbus.SkywayBatchBuilt().Publish(ctx, eventbus.SkywayBatchBuiltEvent{
		ChainReferenceID: "test-chain",
	})
	require.Equal(t, 3, calls["test-chain"], "should execute both subscribers.")

	eventbus.SkywayBatchBuilt().Unsubscribe("test-1")
	eventbus.SkywayBatchBuilt().Publish(ctx, eventbus.SkywayBatchBuiltEvent{
		ChainReferenceID: "test-chain",
	})
	require.Equal(t, 4, calls["test-chain"], "should have removed one subscriber.")
}
