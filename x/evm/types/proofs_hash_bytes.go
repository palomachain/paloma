package types

import (
	fmt "fmt"
	"slices"

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

	if h.SerializedReceipt == nil {
		// We don't have a receipt, just return the transaction
		return tx.MarshalBinary()
	}

	receipt, err := h.GetReceipt()
	if err != nil {
		return nil, err
	}

	serializedTX, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	serializedReceipt, err := receipt.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return slices.Concat(serializedTX, serializedReceipt), nil
}

func (h *TxExecutedProof) GetTX() (*ethtypes.Transaction, error) {
	tx := &ethtypes.Transaction{}

	err := tx.UnmarshalBinary(h.SerializedTX)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (h *TxExecutedProof) GetReceipt() (*ethtypes.Receipt, error) {
	receipt := &ethtypes.Receipt{}

	err := receipt.UnmarshalBinary(h.SerializedReceipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
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

func (h *ReferenceBlockAttestationRes) BytesToHash() ([]byte, error) {
	var res []byte
	res = append(res, []byte(fmt.Sprintf("%d", h.BlockHeight))...)
	res = append(res, []byte(h.BlockHash)...)

	return res, nil
}
