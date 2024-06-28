package types

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/palomachain/paloma/util/liblog"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
)

func (m *UploadSmartContract) VerifyAgainstTX(
	ctx context.Context,
	tx *ethtypes.Transaction,
	_ consensustypes.QueuedSignedMessageI,
	_ *Valset,
	_ *SmartContract,
) error {
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
			logger.WithError(err).Warn("UploadSmartContract VerifyAgainstTX failed to unpack constructor input")
			return err
		}

		input, err := contractABI.Pack("", params...)
		if err != nil {
			logger.WithError(err).Warn("UploadSmartContract VerifyAgainstTX failed to pack constructor params")
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

func (m *SubmitLogicCall) VerifyAgainstTX(
	ctx context.Context,
	tx *ethtypes.Transaction,
	msg consensustypes.QueuedSignedMessageI,
	valset *Valset,
	compass *SmartContract,
) error {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger()).
		WithFields("tx_hash", tx.Hash().Hex())

	logger.Debug("SubmitLogicCall VerifyAgainstTX")

	if valset == nil || compass == nil {
		err := errors.New("missing valset or compass for tx verification")
		logger.WithError(err).Error("failed to verify tx")
		return err
	}

	args := []any{
		BuildCompassConsensus(valset, msg.GetSignData()),
		CompassLogicCallArgs{
			LogicContractAddress: common.HexToAddress(m.GetHexContractAddress()),
			Payload:              m.GetPayload(),
		},
		new(big.Int).SetInt64(int64(msg.GetId())),
		new(big.Int).SetInt64(m.GetDeadline()),
	}

	contractABI, err := abi.JSON(strings.NewReader(compass.GetAbiJSON()))
	if err != nil {
		logger.WithError(err).Warn("SubmitLogicCall VerifyAgainstTX failed to parse compass ABI")
		return err
	}

	input, err := contractABI.Pack("submit_logic_call", args...)
	if err != nil {
		logger.WithError(err).Warn("SubmitLogicCall VerifyAgainstTX failed to pack ABI")
		return err
	}

	if !bytes.Equal(tx.Data(), input) {
		logger.Warn("SubmitLogicCall VerifyAgainstTX failed")
		return ErrEthTxNotVerified
	}

	logger.Debug("SubmitLogicCall VerifyAgainstTX success")

	return nil
}

func (m *UpdateValset) VerifyAgainstTX(
	ctx context.Context,
	tx *ethtypes.Transaction,
	msg consensustypes.QueuedSignedMessageI,
	valset *Valset,
	compass *SmartContract,
) error {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger()).
		WithFields("tx_hash", tx.Hash().Hex())

	logger.Debug("UpdateValset VerifyAgainstTX")

	if valset == nil || compass == nil {
		err := errors.New("missing valset or compass for tx verification")
		logger.WithError(err).Error("failed to verify tx")
		return err
	}

	args := []any{
		BuildCompassConsensus(valset, msg.GetSignData()),
		TransformValsetToCompassValset(m.Valset),
	}

	contractABI, err := abi.JSON(strings.NewReader(compass.GetAbiJSON()))
	if err != nil {
		logger.WithError(err).Warn("UpdateValset VerifyAgainstTX failed to parse compass ABI")
		return err
	}

	input, err := contractABI.Pack("update_valset", args...)
	if err != nil {
		logger.WithError(err).Warn("UpdateValset VerifyAgainstTX failed to pack ABI")
		return err
	}

	if !bytes.Equal(tx.Data(), input) {
		logger.Warn("SubmitLogicCall VerifyAgainstTX failed")
		return ErrEthTxNotVerified
	}

	logger.Debug("UpdateValset VerifyAgainstTX success")

	return nil
}
