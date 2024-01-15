package types

import (
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// TODO: Implement TX verifications

func (m *TransferERC20Ownership) VerifyAgainstTX(tx *ethtypes.Transaction) error {
	return nil
}

func (m *UploadSmartContract) VerifyAgainstTX(tx *ethtypes.Transaction) error {
	// if !bytes.Equal(tx.Data(), append(m.GetBytecode()[:], m.GetConstructorInput()[:]...)) {
	// 	// return ErrEthTxNotVerified
	// }
	return nil
}

func (m *SubmitLogicCall) VerifyAgainstTX(tx *ethtypes.Transaction) error {
	return nil
}

func (m *UpdateValset) VerifyAgainstTX(tx *ethtypes.Transaction, smartContract *SmartContract) error {
	// // TODO
	// arguments := abi.Arguments{
	// 	// addresses
	// 	{Type: whoops.Must(abi.NewType("address[]", "", nil))},
	// 	// powers
	// 	{Type: whoops.Must(abi.NewType("uint256[]", "", nil))},
	// 	// valset id
	// 	{Type: whoops.Must(abi.NewType("uint256", "", nil))},
	// 	// turnstone id
	// 	{Type: whoops.Must(abi.NewType("bytes32", "", nil))},
	// }

	// contractABI, err := abi.JSON(strings.NewReader(smartContract.GetAbiJSON()))
	// input, err := contractABI.Pack("", smartContract.Ge, types.TransformValsetToABIValset(valset))
	return nil
}
