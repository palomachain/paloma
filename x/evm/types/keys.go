package types

import fmt "fmt"

const (
	// ModuleName defines the module name
	ModuleName = "evm"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_evm"

	UserSmartContractStoreKeyPrefix = "user-smart-contract-"
	UserSmartContractKeyPrefix      = "contract-"
)

func UserSmartContractStoreKey(addr string) []byte {
	return []byte(UserSmartContractStoreKeyPrefix + addr)
}

func UserSmartContractKey(id uint64) []byte {
	return []byte(fmt.Sprintf("%s-%d", UserSmartContractKeyPrefix, id))
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
