package bindings

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	schedulerkeeper "github.com/palomachain/paloma/v2/x/scheduler/keeper"
)

func RegisterCustomPlugins(
	scheduler *schedulerkeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(scheduler)
	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(scheduler),
	)

	return []wasmkeeper.Option{
		queryPluginOpt,
		messengerDecoratorOpt,
	}
}
