package types

import "testing"

func TestL(t *testing.T) {
	m := &ArbitrarySmartContractCall{}
	m.Keccak256(456)
}
