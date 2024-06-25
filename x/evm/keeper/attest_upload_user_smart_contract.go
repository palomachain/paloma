package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/evm/types"
)

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
	tx, err := evidence.GetTX()
	if err != nil {
		return fmt.Errorf("failed to get TX: %w", err)
	}

	ethMsg, err := core.TransactionToMessage(tx,
		ethtypes.NewLondonSigner(tx.ChainId()), big.NewInt(0))
	if err != nil {
		a.logger.WithError(err).Error("Failed to extract ethMsg")
		return err
	}

	contractAddr := crypto.CreateAddress(ethMsg.From, tx.Nonce())

	// Update user smart contract deployment
	return a.k.SetUserSmartContractDeploymentActive(ctx, a.action.ValAddress,
		a.action.Id, a.chainReferenceID, contractAddr.String())
}

func (a *uploadUserSmartContractAttester) attemptRetry(ctx sdk.Context) {
	if a.action.Retries >= cMaxSubmitLogicCallRetries {
		a.logger.Error("Max retries for UploadUserSmartContract message reached",
			"message-id", a.msgID,
			"retries", a.action.Retries,
			"chain-reference-id", a.chainReferenceID)

		err := a.k.SetUserSmartContractDeploymentError(ctx, a.action.ValAddress,
			a.action.Id, a.chainReferenceID)
		if err != nil {
			a.logger.WithError(err).Error("Failed to set UserSmartContract error")
		}

		return
	}

	a.action.Retries++

	a.logger.Info("Retrying failed UploadSmartContract message",
		"message-id", a.msgID,
		"retries", a.action.Retries,
		"chain-reference-id", a.chainReferenceID)

	newMsgID, err := a.k.AddUploadUserSmartContractToConsensus(ctx, a.chainReferenceID, a.action)
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
