package types

import (
	fmt "fmt"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type Hashable interface {
	BytesToHash() ([]byte, error)
}

var (
	_ Hashable = &TxExecutedProof{}
	_ Hashable = &SmartContractExecutionErrorProof{}
)

func (h *TxExecutedProof) BytesToHash() ([]byte, error) {
	tx, err := h.GetTX()
	if err != nil {
		return nil, err
	}
	return tx.MarshalBinary()
}

func (h *TxExecutedProof) GetTX() (*ethtypes.Transaction, error) {
	tx := &ethtypes.Transaction{}

	err := tx.UnmarshalBinary(h.SerializedTX)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (h *SmartContractExecutionErrorProof) BytesToHash() ([]byte, error) {
	return []byte(h.ErrorMessage), nil
}

func (h *ValidatorBalancesAttestationRes) BytesToHash() ([]byte, error) {
	var res []byte
	res = append(res, []byte(fmt.Sprintf("%d", h.BlockHeight))...)
	for _, val := range h.Balances {
		res = append(res, []byte("\n"+val)...)
	}
	return res, nil
}
