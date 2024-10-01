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
	"github.com/palomachain/paloma/v2/util/liblog"
	"github.com/palomachain/paloma/v2/util/slice"
	consensustypes "github.com/palomachain/paloma/v2/x/consensus/types"
)

type FeeArgs struct {
	RelayerFee            *big.Int
	CommunityFee          *big.Int
	SecurityFee           *big.Int
	FeePayerPalomaAddress [32]byte
}

func (m *UploadSmartContract) VerifyAgainstTX(
	ctx context.Context,
	tx *ethtypes.Transaction,
	_ consensustypes.QueuedSignedMessageI,
	_ *Valset,
	_ *SmartContract,
	_ string,
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
	relayer string,
) error {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger()).
		WithFields("tx_hash", tx.Hash().Hex(), "valset_id", valset.ValsetID)

	logger.Debug("SubmitLogicCall VerifyAgainstTX")

	if valset == nil || compass == nil {
		err := errors.New("missing valset or compass for tx verification")
		logger.WithError(err).Error("failed to verify tx")
		return err
	}

	contractABI, err := abi.JSON(strings.NewReader(compass.GetAbiJSON()))
	if err != nil {
		logger.WithError(err).Warn("SubmitLogicCall VerifyAgainstTX failed to parse compass ABI")
		return err
	}

	padding := bytes.Repeat([]byte{0}, 32-len(m.SenderAddress))
	paddedSenderAddress := [32]byte(append(padding, m.SenderAddress...))

	feeArgs := FeeArgs{
		RelayerFee:            big.NewInt(0).SetUint64(m.Fees.RelayerFee),
		CommunityFee:          big.NewInt(0).SetUint64(m.Fees.CommunityFee),
		SecurityFee:           big.NewInt(0).SetUint64(m.Fees.SecurityFee),
		FeePayerPalomaAddress: paddedSenderAddress,
	}

	// Since some validators might have added their signature to the message
	// after a pigeon start relaying it, we iteratively remove the end signature
	// until we get a match, or there are no more signatures.
	for i := len(msg.GetSignData()); i > 0; i-- {
		args := []any{
			BuildCompassConsensus(valset, msg.GetSignData()[0:i]),
			CompassLogicCallArgs{
				LogicContractAddress: common.HexToAddress(m.GetHexContractAddress()),
				Payload:              m.GetPayload(),
			},
			feeArgs,
			new(big.Int).SetInt64(int64(msg.GetId())),
			new(big.Int).SetInt64(m.GetDeadline()),
			common.HexToAddress(relayer),
		}

		input, err := contractABI.Pack("submit_logic_call", args...)
		if err != nil {
			logger.WithError(err).Warn("SubmitLogicCall VerifyAgainstTX failed to pack ABI")
			return err
		}

		if bytes.Equal(tx.Data(), input) {
			logger.Debug("SubmitLogicCall VerifyAgainstTX success")
			return nil
		}
	}

	logger.Warn("SubmitLogicCall VerifyAgainstTX failed")
	return ErrEthTxNotVerified
}

func (m *CompassHandover) VerifyAgainstTX(
	ctx context.Context,
	tx *ethtypes.Transaction,
	msg consensustypes.QueuedSignedMessageI,
	valset *Valset,
	compass *SmartContract,
	relayer string,
) error {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger()).
		WithFields("tx_hash", tx.Hash().Hex(), "valset_id", valset.ValsetID)

	logger.Debug("CompassHandover VerifyAgainstTX")

	if valset == nil || compass == nil {
		err := errors.New("missing valset or compass for tx verification")
		logger.WithError(err).Error("failed to verify tx")
		return err
	}

	contractABI, err := abi.JSON(strings.NewReader(compass.GetAbiJSON()))
	if err != nil {
		logger.WithError(err).Warn("CompassHandover VerifyAgainstTX failed to parse compass ABI")
		return err
	}

	forwardArgs := slice.Map(m.GetForwardCallArgs(), func(arg CompassHandover_ForwardCallArgs) CompassLogicCallArgs {
		return CompassLogicCallArgs{
			common.HexToAddress(arg.GetHexContractAddress()),
			arg.GetPayload(),
		}
	})

	// Since some validators might have added their signature to the message
	// after a pigeon start relaying it, we iteratively remove the end signature
	// until we get a match, or there are no more signatures.
	for i := len(msg.GetSignData()); i > 0; i-- {
		args := []any{
			BuildCompassConsensus(valset, msg.GetSignData()[0:i]),
			forwardArgs,
			new(big.Int).SetInt64(m.GetDeadline()),
			big.NewInt(0).SetUint64(msg.GetGasEstimate()),
			common.HexToAddress(relayer),
		}

		input, err := contractABI.Pack("compass_update_batch", args...)
		if err != nil {
			logger.WithError(err).Warn("CompassHandover VerifyAgainstTX failed to pack ABI")
			return err
		}

		if bytes.Equal(tx.Data(), input) {
			logger.Debug("CompassHandover VerifyAgainstTX success")
			return nil
		}
	}

	logger.Warn("CompassHandover VerifyAgainstTX failed")
	return ErrEthTxNotVerified
}

func (m *UpdateValset) VerifyAgainstTX(
	ctx context.Context,
	tx *ethtypes.Transaction,
	msg consensustypes.QueuedSignedMessageI,
	valset *Valset,
	compass *SmartContract,
	relayer string,
) error {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger()).
		WithFields("tx_hash", tx.Hash().Hex(),
			"valset_id", valset.ValsetID,
			"updated_valset_id", m.Valset.ValsetID)

	logger.Debug("UpdateValset VerifyAgainstTX")

	if valset == nil || compass == nil {
		err := errors.New("missing valset or compass for tx verification")
		logger.WithError(err).Error("failed to verify tx")
		return err
	}

	contractABI, err := abi.JSON(strings.NewReader(compass.GetAbiJSON()))
	if err != nil {
		logger.WithError(err).Warn("UpdateValset VerifyAgainstTX failed to parse compass ABI")
		return err
	}

	// Since some validators might have added their signature to the message
	// after a pigeon start relaying it, we iteratively remove the end signature
	// until we get a match, or there are no more signatures.
	for i := len(msg.GetSignData()); i > 0; i-- {
		args := []any{
			BuildCompassConsensus(valset, msg.GetSignData()[0:i]),
			TransformValsetToCompassValset(m.Valset),
			common.HexToAddress(relayer),
			big.NewInt(0).SetUint64(msg.GetGasEstimate()),
		}

		input, err := contractABI.Pack("update_valset", args...)
		if err != nil {
			logger.WithError(err).Warn("UpdateValset VerifyAgainstTX failed to pack ABI")
			return err
		}

		if bytes.Equal(tx.Data(), input) {
			logger.Debug("UpdateValset VerifyAgainstTX success")
			return nil
		}
	}

	logger.Warn("UpdateValset VerifyAgainstTX failed")
	return ErrEthTxNotVerified
}

func (m *UploadUserSmartContract) VerifyAgainstTX(
	ctx context.Context,
	tx *ethtypes.Transaction,
	msg consensustypes.QueuedSignedMessageI,
	valset *Valset,
	compass *SmartContract,
	relayer string,
) error {
	logger := liblog.FromSDKLogger(sdk.UnwrapSDKContext(ctx).Logger()).
		WithFields("tx_hash", tx.Hash().Hex(), "valset_id", valset.ValsetID)

	logger.Debug("UploadUserSmartContract VerifyAgainstTX")

	if valset == nil || compass == nil {
		err := errors.New("missing valset or compass for tx verification")
		logger.WithError(err).Error("failed to verify tx")
		return err
	}

	contractABI, err := abi.JSON(strings.NewReader(compass.GetAbiJSON()))
	if err != nil {
		logger.WithError(err).Warn("UploadUserSmartContract VerifyAgainstTX failed to parse compass ABI")
		return err
	}

	padding := bytes.Repeat([]byte{0}, 32-len(m.SenderAddress))
	paddedAuthor := [32]byte(append(padding, m.SenderAddress...))

	feeArgs := FeeArgs{
		RelayerFee:            big.NewInt(0).SetUint64(m.Fees.RelayerFee),
		CommunityFee:          big.NewInt(0).SetUint64(m.Fees.CommunityFee),
		SecurityFee:           big.NewInt(0).SetUint64(m.Fees.SecurityFee),
		FeePayerPalomaAddress: paddedAuthor,
	}

	// Since some validators might have added their signature to the message
	// after a pigeon start relaying it, we iteratively remove the end signature
	// until we get a match, or there are no more signatures.
	for i := len(msg.GetSignData()); i > 0; i-- {
		args := []any{
			BuildCompassConsensus(valset, msg.GetSignData()[0:i]),
			common.HexToAddress(m.GetDeployerAddress()),
			m.GetBytecode(),
			feeArgs,
			new(big.Int).SetInt64(int64(msg.GetId())),
			new(big.Int).SetInt64(m.GetDeadline()),
			common.HexToAddress(relayer),
		}

		input, err := contractABI.Pack("deploy_contract", args...)
		if err != nil {
			logger.WithError(err).Warn("UploadUserSmartContract VerifyAgainstTX failed to pack ABI")
			return err
		}

		if bytes.Equal(tx.Data(), input) {
			logger.Debug("UploadUserSmartContract VerifyAgainstTX success")
			return nil
		}
	}

	logger.Warn("UploadUserSmartContract VerifyAgainstTX failed")
	return ErrEthTxNotVerified
}
