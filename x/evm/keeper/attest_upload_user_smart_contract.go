package keeper

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/evm/types"
)

var contractDeployedEvent = crypto.Keccak256Hash([]byte(
	"ContractDeployed(address,address,uint256)",
))

type uploadUserSmartContractAttester struct {
	attestionParameters
	action *types.UploadUserSmartContract
	logger liblog.Logr
	k      *Keeper
}

func newUploadUserSmartContractAttester(
	k *Keeper,
	l liblog.Logr,
	p attestionParameters,
) *uploadUserSmartContractAttester {
	return &uploadUserSmartContractAttester{
		attestionParameters: p,
		logger:              l,
		k:                   k,
	}
}

func (a *uploadUserSmartContractAttester) Execute(ctx sdk.Context) error {
	a.logger = a.logger.WithFields("action-msg", "Message_UploadUserSmartContract")
	a.logger.Debug("Processing upload user smart contract message attestation.")

	a.action = a.msg.Action.(*types.Message_UploadUserSmartContract).UploadUserSmartContract

	switch evidence := a.rawEvidence.(type) {
	case *types.TxExecutedProof:
		return a.attest(ctx, evidence)
	case *types.SmartContractExecutionErrorProof:
		a.logger.Debug("User smart contract execution error proof",
			"user-smart-contract-error", evidence.GetErrorMessage())
		keeperutil.EmitEvent(a.k, ctx, types.SmartContractExecutionFailedKey,
			types.SmartContractExecutionFailedMessageID.With(fmt.Sprintf("%d", a.msgID)),
			types.SmartContractExecutionFailedChainReferenceID.With(a.chainReferenceID),
			types.SmartContractExecutionFailedError.With(evidence.GetErrorMessage()),
			types.SmartContractExecutionMessageType.With(fmt.Sprintf("%T", a.action)),
		)
		a.attemptRetry(ctx)
		return nil
	default:
		a.logger.Error("unknown type %t when attesting", evidence)
		return ErrUnexpectedError.JoinErrorf("unknown type %t when attesting", evidence)
	}
}

func (a *uploadUserSmartContractAttester) attest(
	ctx sdk.Context,
	evidence *types.TxExecutedProof,
) error {
	_, err := attestTransactionIntegrity(ctx, a.originalMessage, a.k, evidence,
		a.chainReferenceID, a.msg.AssigneeRemoteAddress, a.action.VerifyAgainstTX)
	if err != nil {
		a.logger.WithError(err).Error("Failed to verify transaction integrity.")
		return err
	}

	// TODO this may fail if we deploy a different compass with changes to this
	// event between uploading the contract and attesting
	chainInfo, err := a.k.GetChainInfo(ctx, a.chainReferenceID)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get chain info")
		return fmt.Errorf("failed to get chain info: %w", err)
	}

	compassABI, err := abi.JSON(strings.NewReader(chainInfo.Abi))
	if err != nil {
		fmt.Printf("ABI: %v\n", chainInfo.Abi)
		a.logger.WithError(err).Error("Failed to unpack compass ABI")
		return fmt.Errorf("failed to unpack compass ABI: %w", err)
	}

	receipt, err := evidence.GetReceipt()
	if err != nil {
		a.logger.WithError(err).Error("Failed to get evidence receipt")
		return fmt.Errorf("failed to get evidence receipt: %w", err)
	}

	for i := range receipt.Logs {
		if len(receipt.Logs[i].Topics) == 0 {
			a.logger.Error("Invalid topics in receipt logs")
			return errors.New("invalid topics in receipt logs")
		}

		if receipt.Logs[i].Topics[0] != contractDeployedEvent {
			continue
		}

		// This is the event we want, so we get the contract address from it
		event, err := compassABI.Unpack("ContractDeployed", receipt.Logs[i].Data)
		if err != nil {
			a.logger.WithError(err).Error("Failed to unpack event")
			return fmt.Errorf("failed to unpack event: %w", err)
		}

		contractAddr, ok := event[0].(common.Address)
		if !ok {
			a.logger.Error("Invalid contract address in event")
			return errors.New("invalid smart contract address")
		}

		// Update user smart contract deployment
		return a.k.SetUserSmartContractDeploymentActive(ctx,
			sdk.ValAddress(a.action.SenderAddress).String(), a.action.Id,
			a.action.BlockHeight, a.chainReferenceID, contractAddr.String())
	}

	// Either the tx didn't generate logs, or it didn't have the one we want
	return errors.New("failed to find deployed contract address")
}

func (a *uploadUserSmartContractAttester) attemptRetry(ctx sdk.Context) {
	if a.action.Retries >= cMaxSubmitLogicCallRetries {
		a.logger.Error("Max retries for UploadUserSmartContract message reached",
			"message-id", a.msgID,
			"retries", a.action.Retries,
			"chain-reference-id", a.chainReferenceID)

		author := sdk.ValAddress(a.action.SenderAddress)

		err := a.k.SetUserSmartContractDeploymentError(ctx, author.String(),
			a.action.Id, a.action.BlockHeight, a.chainReferenceID)
		if err != nil {
			a.logger.WithError(err).Error("Failed to set UserSmartContract error")
		}

		return
	}

	a.action.Retries++
	// We must clear fees before retry or the signature verification fails
	a.action.Fees = nil

	a.logger.Info("Retrying failed UploadUserSmartContract message",
		"message-id", a.msgID,
		"retries", a.action.Retries,
		"chain-reference-id", a.chainReferenceID)

	newMsgID, err := a.k.AddUploadUserSmartContractToConsensus(ctx,
		a.chainReferenceID, a.msg.TurnstoneID, a.action)
	if err != nil {
		a.logger.WithError(err).Error("Failed to retry UploadUserSmartContract")
		return
	}

	a.logger.Info("Retried failed UploadUserSmartContract message",
		"message-id", a.msgID,
		"new-message-id", newMsgID,
		"retries", a.action.Retries,
		"chain-reference-id", a.chainReferenceID)
}
