package types

import "math/big"

func (msg *ChainInfo) IsActive() bool {
	return msg.SmartContractAddr != ""
}

func (msg *ChainInfo) GetMinOnChainBalanceBigInt() (*big.Int, error) {
	num, ok := new(big.Int).SetString(msg.MinOnChainBalance, 10)
	if !ok {
		return nil, ErrInvalidBalance.Format(msg.MinOnChainBalance)
	}
	return num, nil
}
