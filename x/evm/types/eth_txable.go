package types

import (
	"bytes"
	"context"
	"encoding/hex"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/palomachain/paloma/util/liblog"
)

// TODO: Implement TX verifications

func (m *TransferERC20Ownership) VerifyAgainstTX(ctx context.Context, tx *ethtypes.Transaction) error {
	return nil
}

func (m *UploadSmartContract) VerifyAgainstTX(ctx context.Context, tx *ethtypes.Transaction) error {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger()).
		WithFields("tx_hash", tx.Hash().Hex(), "msg_id", m.Id)

	contractABI, err := abi.JSON(strings.NewReader(m.GetAbi()))
	if err != nil {
		return err
	}

	mData := make([]byte, len(m.GetBytecode()))
	copy(mData, m.GetBytecode())

	// We should always have ConstructorInput data on smart contract uploads,
	// but just in case we don't, the TX input will be only the bytecode
	if len(m.GetConstructorInput()) > 0 {
		params, err := contractABI.Constructor.Inputs.Unpack(m.GetConstructorInput())
		if err != nil {
			return err
		}

		input, err := contractABI.Pack("", params...)
		if err != nil {
			return err
		}

		mData = append(mData, input...)
	}

	if !bytes.Equal(tx.Data(), mData) {
		logger.Warn("UploadSmartContract VerifyAgainstTX failed")
		return ErrEthTxNotVerified
	}

	logger.Debug("UploadSmartContract VerifyAgainstTX success")

	return nil
}

func (m *SubmitLogicCall) VerifyAgainstTX(ctx context.Context, tx *ethtypes.Transaction) error {
	liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger()).
		WithFields(
			"data", hex.EncodeToString(tx.Data()),
			"hash", tx.Hash().Hex(),
		).
		Debug("VerifyAgainstTX SubmitLogicCall")
	return nil
}

func (m *UpdateValset) VerifyAgainstTX(ctx context.Context, tx *ethtypes.Transaction, smartContract *SmartContract) error {
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
