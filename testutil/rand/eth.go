package rand

import (
	"github.com/ethereum/go-ethereum/common"
)

func ETHAddress() common.Address {
	return common.BytesToAddress(Bytes(20))
}
