package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/evm/types"
)

type submitLogicCallAttester struct {
	attestionParameters
	action *types.SubmitLogicCall
	logger liblog.Logr
	k      *Keeper
}

func newSubmitLogicCallAttester(k *Keeper, l liblog.Logr, p attestionParameters) *submitLogicCallAttester {
	return &submitLogicCallAttester{
		attestionParameters: p,
		logger:              l,
		k:                   k,
	}
}

func (a *submitLogicCallAttester) Execute(ctx sdk.Context) error {
	a.logger = a.logger.WithFields("action-msg", "Message_SubmitLogicCall")
	a.logger.Debug("Processing submit logic call message attestation.")

	a.action = a.msg.Action.(*types.Message_SubmitLogicCall).SubmitLogicCall

	switch winner := a.rawEvidence.(type) {
	case *types.TxExecutedProof:
		return a.attest(ctx, winner)
	case *types.SmartContractExecutionErrorProof:
		return a.attemptRetry(ctx, winner)
	default:
		return ErrUnexpectedError.JoinErrorf("unknown type %t when attesting", winner)
	}
}

func (a *submitLogicCallAttester) attest(ctx sdk.Context, evidence *types.TxExecutedProof) (err error) {
	_, err = attestTransactionIntegrity(ctx, a.originalMessage, a.k, evidence,
		a.chainReferenceID, a.msg.AssigneeRemoteAddress, a.action.VerifyAgainstTX)
	if err != nil {
		a.logger.WithError(err).Error("Failed to verify transaction integrity.")
		return err
	}

	return nil
}

func (a *submitLogicCallAttester) attemptRetry(ctx sdk.Context, proof *types.SmartContractExecutionErrorProof) (err error) {
	keeperutil.EmitEvent(a.k, ctx, types.SmartContractExecutionFailedKey,
		types.SmartContractExecutionFailedMessageID.With(fmt.Sprintf("%d", a.msgID)),
		types.SmartContractExecutionFailedChainReferenceID.With(a.chainReferenceID),
		types.SmartContractExecutionFailedError.With(proof.GetErrorMessage()),
		types.SmartContractExecutionMessageType.With(fmt.Sprintf("%T", a.action)),
	)

	// Retry message if eligible
	// Must be less than cMaxSubmitLogicCallRetries
	var newMsgID uint64
	slc := a.action
	if slc.Retries < cMaxSubmitLogicCallRetries {
		slc.Retries++
		// We must clear fees before retry or the signature verification fails
		slc.Fees = nil
		a.logger.Info("retrying failed SubmitLogicCall message",
			"message-id", a.msgID,
			"retries", slc.Retries,
			"chain-reference-id", a.chainReferenceID)

		var err error
		newMsgID, err = a.k.AddSmartContractExecutionToConsensus(ctx, a.chainReferenceID, a.msg.GetTurnstoneID(), slc)
		if err != nil {
			a.logger.WithError(err).Error("Failed to retry SubmitLogicCall")
		} else {
			a.logger.Info("retried failed SubmitLogicCall message",
				"message-id", a.msgID,
				"new-message-id", newMsgID,
				"retries", slc.Retries,
				"chain-reference-id", a.chainReferenceID)
		}
	} else {
		a.logger.Error("max retries for message reached",
			"message-id", a.msgID,
			"retries", slc.Retries,
			"chain-reference-id", a.chainReferenceID)
	}

	return nil
}
