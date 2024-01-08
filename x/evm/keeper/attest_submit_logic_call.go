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
	_, err = attestTransactionIntegrity(ctx, a.k, evidence, a.action.VerifyAgainstTX)
	if err != nil {
		a.logger.WithError(err).Error("Failed to verify transaction integrity.")
		return err
	}

	if a.k.deploymentCache.Has(ctx, a.chainReferenceID, a.msgID) {
		smartContractID := a.k.deploymentCache.Get(ctx, a.chainReferenceID, a.msgID)
		deployment, _ := a.k.getSmartContractDeploymentByContractID(ctx, smartContractID, a.chainReferenceID)
		if deployment == nil {
			a.logger.WithError(err).Error("Deployment not found")
			return ErrCannotActiveSmartContractThatIsNotDeploying
		}

		hasPendingTransfers := false
		for i, v := range deployment.Erc20Transfers {
			if v.GetMsgID() == a.msgID {
				if v.GetStatus() != types.SmartContractDeployment_ERC20Transfer_PENDING {
					err = fmt.Errorf("invalid transfer status: %v", v.GetStatus())
					a.logger.WithError(err).WithFields("transfer-status", v.GetStatus()).Error("Failed to attest ERC20 transfer message, invalid status.")
					return err
				}
				deployment.Erc20Transfers[i].Status = types.SmartContractDeployment_ERC20Transfer_OK
			}

			if deployment.Erc20Transfers[i].Status != types.SmartContractDeployment_ERC20Transfer_OK {
				hasPendingTransfers = true
			}
		}

		if err := a.k.updateSmartContractDeployment(ctx, smartContractID, a.chainReferenceID, deployment); err != nil {
			a.logger.WithError(err).Error("Failed to update smart contract deployment.")
			return err
		}

		if !hasPendingTransfers {
			if err := a.k.SetSmartContractAsActive(ctx, smartContractID, a.chainReferenceID); err != nil {
				a.logger.WithError(err).Error("Failed to set smart contract as active")
				return err
			}

			a.logger.Debug("attestation successful")
		}

		a.k.deploymentCache.Delete(ctx, a.chainReferenceID, a.msgID)
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
	var newMsgID uint64
	slc := a.action
	if slc.Retries < cMaxSubmitLogicCallRetries {
		slc.Retries++
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

	// Update deployment accordingly if exists
	// Sets existing transfer state to FAILAED and removes it from the cache.
	// If retry is happening, creates a new transfer record on the
	// deployment and adds it to the cache.
	var deployment *types.SmartContractDeployment
	if a.k.deploymentCache.Has(ctx, a.chainReferenceID, a.msgID) {
		smartContractID := a.k.deploymentCache.Get(ctx, a.chainReferenceID, a.msgID)
		deployment, _ = a.k.getSmartContractDeploymentByContractID(ctx, smartContractID, a.chainReferenceID)
		if deployment == nil {
			err := fmt.Errorf("no matching deployment found for contract ID %v on chain %v", smartContractID, a.chainReferenceID)
			a.logger.WithError(err).Error(err.Error())
			return err
		}

		var failedTransfer types.SmartContractDeployment_ERC20Transfer
		for i, v := range deployment.Erc20Transfers {
			if v.GetMsgID() == a.msgID {
				if v.GetStatus() != types.SmartContractDeployment_ERC20Transfer_PENDING {
					a.logger.WithFields("transfer-status", v.GetStatus()).Error("Unexpected status of failed message")
				}
				deployment.Erc20Transfers[i].Status = types.SmartContractDeployment_ERC20Transfer_FAIL
				failedTransfer = deployment.Erc20Transfers[i]
			}
		}

		if newMsgID != 0 {
			deployment.Erc20Transfers = append(deployment.Erc20Transfers, types.SmartContractDeployment_ERC20Transfer{
				Denom:  failedTransfer.Denom,
				Erc20:  failedTransfer.Erc20,
				MsgID:  newMsgID,
				Status: types.SmartContractDeployment_ERC20Transfer_PENDING,
			})

			defer func() {
				if err == nil {
					a.k.deploymentCache.Add(ctx, a.chainReferenceID, smartContractID, newMsgID)
				}
			}()
		}

		if err := a.k.updateSmartContractDeployment(ctx, smartContractID, a.chainReferenceID, deployment); err != nil {
			a.logger.WithError(err).Error("Failed to update smart contract deployment.")
			return err
		}

		a.k.deploymentCache.Delete(ctx, a.chainReferenceID, a.msgID)
	}

	return nil
}
