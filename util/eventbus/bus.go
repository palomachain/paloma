package eventbus

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/liblog"
)

var gravityBatchBuilt = newEvent[GravityBatchBuiltEvent]()

type (
	EventHandler[E any] func(context.Context, E) error
	Event[E any]        struct {
		subscribers map[string]EventHandler[E]
	}
)

func newEvent[E any]() Event[E] {
	return Event[E]{
		subscribers: make(map[string]EventHandler[E]),
	}
}

func (e Event[E]) Publish(ctx context.Context, event E) {
	for s, fn := range e.subscribers {
		if fn != nil {
			logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger()).
				WithComponent("eventbus").
				WithFields("event", event).
				WithFields("subscriber", s)
			logger.Debug("Handling event")
			if err := fn(ctx, event); err != nil {
				logger.WithError(err).Error("Failed to handle event")
			}
		}
	}
}

func (e Event[E]) Subscribe(id string, fn EventHandler[E]) {
	e.subscribers[id] = fn
}

func (e Event[E]) Unsubscribe(id string) {
	e.subscribers[id] = nil
}

type GravityBatchBuiltEvent struct {
	ChainReferenceID string
}

func GravityBatchBuilt() *Event[GravityBatchBuiltEvent] {
	return &gravityBatchBuilt
}
