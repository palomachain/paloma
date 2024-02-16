package types

import keeperutil "github.com/palomachain/paloma/util/keeper"

const (
	// ModuleName defines the module name
	ModuleName = "metrix"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_metrix"

	// MetricsStorePrefix defines the prefix for the module's metrics store
	MetricsStorePrefix = "metrics"

	// HistoryStorePrefix defines the prefix for the module's historic relay data store
	HistoryStorePrefix = "history"

	// MessageNonceCacheStorePrefix defines the prefix for the module's message nonce cache store.
	MessageNonceCacheStorePrefix = "message-nonce"

	// MessageNonceCacheKey defines the key to access the singleton resource on the message nonce store.
	MessageNonceCacheKey keeperutil.Key = "highest-message-nonce"
)
