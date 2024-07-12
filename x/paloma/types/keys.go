package types

const (
	// ModuleName defines the module name
	ModuleName = "paloma"

	// StoreKey defines the primary module store key
	StoreKey = "paloma-store"

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_paloma_store"
)

var LightNodeClientKeyPrefix = []byte("light-node-client")

func KeyPrefix(p string) []byte {
	return []byte(p)
}
