package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/VolumeFi/whoops"

	// "cosmossdk.io/store/prefix"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
)

const cMaxSubmitLogicCallRetries uint32 = 2

func hashSha256(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

type consensusPower struct {
	runningSum sdkmath.Int
	totalPower sdkmath.Int
}

func (c *consensusPower) setTotal(total sdkmath.Int) {
	c.totalPower = total
}

func (c *consensusPower) add(power sdkmath.Int) {
	var zero sdkmath.Int
	if c.runningSum == zero {
		c.runningSum = sdkmath.NewInt(0)
	}
	c.runningSum = c.runningSum.Add(power)
}

func (c *consensusPower) consensus() bool {
	var zero sdkmath.Int
	if c.runningSum == zero {
		return false
	}
	/*
		sum >= totalPower * 2 / 3
		===
		3 * sum >= totalPower * 2
	*/
	return c.runningSum.Mul(sdkmath.NewInt(3)).GTE(
		c.totalPower.Mul(sdkmath.NewInt(2)),
	)
}

func (k Keeper) attestRouter(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) (err error) {
	logger := k.Logger(ctx).WithFields(
		"component", "attest-router",
		"msg-id", msg.GetId(),
		"msg-nonce", msg.Nonce())
	logger.Debug("attest-router")

	if len(msg.GetEvidence()) == 0 {
		return nil
	}

	ctx, writeCache := ctx.CacheContext()
	defer func() {
		if err != nil {
			logger.WithError(err).Error("failed to attest. Skipping writeback.")
			return
		}
		writeCache()
	}()

	consensusMsg, err := msg.ConsensusMsg(k.cdc)
	if err != nil {
		logger.WithError(err).Error("failed to cast to consensus message")
		return err
	}

	evidence, err := k.findEvidenceThatWon(ctx, msg.GetEvidence())
	if err != nil {
		if errors.Is(err, ErrConsensusNotAchieved) {
			logger.WithError(err).Error("consensus not achieved")
			return nil
		}
		logger.WithError(err).Error("failed to find evidence")
		return err
	}

	defer func() {
		// given that there was enough evidence for a proof, regardless of the outcome,
		// we should remove this from the queue as there isn't much that we can do about it.
		if err := q.Remove(ctx, msg.GetId()); err != nil {
			logger.WithError(err).Error("error removing message, attestRouter")
		}
	}()

	defer func() {
		// if the input type is a TX, then regardles, we want to set it as already processed
		switch winner := evidence.(type) {
		case *types.TxExecutedProof:
			tx, err := winner.GetTX()
			if err == nil {
				k.setTxAsAlreadyProcessed(ctx, tx)
			}
		}
	}()
	// If we got up to here it means that the enough evidence was provided
	actionMsg := consensusMsg.(*types.Message).GetAction()
	_, chainReferenceID := q.ChainInfo()
	logger = logger.WithFields("chain-reference-id", chainReferenceID)

	switch origMsg := actionMsg.(type) {
	case *types.Message_TransferERC20Ownership:
		logger = logger.WithFields("action-msg", "Message_TransferERC20Ownership")
		logger.Debug("Processing attestation.")
		defer func() {
			// regardless of the outcome, this upload/deployment should be removed
			id := origMsg.TransferERC20Ownership.GetSmartContractID()
			logger.With("smart-contract-id", id).Debug("removing deployment.")
			k.DeleteSmartContractDeploymentByContractID(ctx, id, chainReferenceID)
		}()
		switch winner := evidence.(type) {
		case *types.TxExecutedProof:
			logger.Debug("TX Execution Proof")
			tx, err := winner.GetTX()
			if err != nil {
				logger.WithError(err).Error("Failed to get TX")
				return err
			}
			if k.isTxProcessed(ctx, tx) {
				// somebody submitted the old transaction that was already processed?
				// punish those validators!!
				logger.WithError(err).Error("TX already processed")
				// hack hack hack: We want to keep this check in place, but there is no real TX to work with here yet.
				// return ErrUnexpectedError.WrapS("transaction %s is already processed", tx.Hash())
			}
			err = origMsg.TransferERC20Ownership.VerifyAgainstTX(tx)
			if err != nil {
				// passed in transaction doesn't seem to be created from this smart contract
				logger.WithError(err).Error("TX failed to verify")
				return err
			}

			deployment, _ := k.getSmartContractDeploymentByContractID(ctx, origMsg.TransferERC20Ownership.SmartContractID, chainReferenceID)
			if deployment == nil {
				logger.WithError(err).Error("Deployment not found")
				return ErrCannotActiveSmartContractThatIsNotDeploying
			}
			if deployment.GetStatus() != types.SmartContractDeployment_WAITING_FOR_ERC20_OWNERSHIP_TRANSFER {
				logger.WithError(err).Error("Deployment not awaiting transfer")
				return ErrCannotActiveSmartContractThatIsNotDeploying
			}

			smartContract, err := k.getSmartContract(ctx, deployment.GetSmartContractID())
			if err != nil {
				logger.WithError(err).Error("Failed to get contract")
				return err
			}

			smartContractAddr := common.BytesToAddress(origMsg.TransferERC20Ownership.NewCompassAddress).Hex()
			err = k.ActivateChainReferenceID(
				ctx,
				chainReferenceID,
				smartContract,
				smartContractAddr,
				deployment.GetUniqueID(),
			)
			if err != nil {
				logger.WithError(err).Error("Failed to activate chain")
				return err
			}
			logger.Debug("attestation successful")

		case *types.SmartContractExecutionErrorProof:
			logger.Debug("Error Proof")
			keeperutil.EmitEvent(k, ctx, types.SmartContractExecutionFailedKey,
				types.SmartContractExecutionFailedMessageID.With(fmt.Sprintf("%d", msg.GetId())),
				types.SmartContractExecutionFailedChainReferenceID.With(chainReferenceID),
				types.SmartContractExecutionFailedError.With(winner.GetErrorMessage()),
				types.SmartContractExecutionMessageType.With(fmt.Sprintf("%T", origMsg)),
			)
		default:
			return ErrUnexpectedError.WrapS("unknown type %t when attesting", winner)
		}
	case *types.Message_UploadSmartContract:
		logger = logger.WithFields("action-msg", "Message_UploadSmartContract")
		logger.Debug("Processing upload smart contract message attestation.")
		switch winner := evidence.(type) {
		case *types.TxExecutedProof:
			logger.Debug("TX proof attestation")
			tx, err := winner.GetTX()
			if err != nil {
				logger.WithError(err).Error("Failed to get TX")
				return err
			}
			if k.isTxProcessed(ctx, tx) {
				// somebody submitted the old transaction that was already processed?
				// punish those validators!!
				logger.WithError(err).Error("TX already processed")
				return ErrUnexpectedError.WrapS("transaction %s is already processed", tx.Hash())
			}
			err = origMsg.UploadSmartContract.VerifyAgainstTX(tx)
			if err != nil {
				// passed in transaction doesn't seem to be created from this smart contract
				logger.WithError(err).Error("Failed to verify TX")
				return err
			}

			ethMsg, err := core.TransactionToMessage(tx, ethtypes.NewLondonSigner(tx.ChainId()), big.NewInt(0))
			if err != nil {
				logger.WithError(err).Error("Failed to extract ethMsg")
				return err
			}

			smartContractID := origMsg.UploadSmartContract.GetId()
			deployment, _ := k.getSmartContractDeploymentByContractID(ctx, smartContractID, chainReferenceID)
			if deployment == nil {
				logger.WithError(err).WithFields("smart-contract-id", smartContractID).Error("Smart contract not found")
				return ErrCannotActiveSmartContractThatIsNotDeploying
			}

			if deployment.GetStatus() != types.SmartContractDeployment_IN_FLIGHT {
				logger.WithError(err).Error("deployment not in right state")
				return ErrCannotActiveSmartContractThatIsNotDeploying
			}

			turnstoneID := consensusMsg.(*types.Message).GetTurnstoneID()
			newCompassAddr := crypto.CreateAddress(ethMsg.From, tx.Nonce())
			err = k.initiateERC20TokenOwnershipTransfer(
				ctx,
				chainReferenceID,
				turnstoneID,
				&types.TransferERC20Ownership{
					SmartContractID:   smartContractID,
					NewCompassAddress: newCompassAddr.Bytes(),
				})
			if err != nil {
				logger.WithError(err).Error("failed to initiateERC20TokenOwnershipTransfer")
				return err
			}

			err = k.SetSmartContractDeploymentStatusByContractID(ctx, smartContractID, chainReferenceID, types.SmartContractDeployment_WAITING_FOR_ERC20_OWNERSHIP_TRANSFER)
			if err != nil {
				logger.WithError(err).Error("failed to update deployment status!")
				return err
			}
			logger.Debug("attestation successful")

		case *types.SmartContractExecutionErrorProof:
			logger.Debug("smart contract execution error proof", "smart-contract-error", winner.GetErrorMessage())
			keeperutil.EmitEvent(k, ctx, types.SmartContractExecutionFailedKey,
				types.SmartContractExecutionFailedMessageID.With(fmt.Sprintf("%d", msg.GetId())),
				types.SmartContractExecutionFailedChainReferenceID.With(chainReferenceID),
				types.SmartContractExecutionFailedError.With(winner.GetErrorMessage()),
				types.SmartContractExecutionMessageType.With(fmt.Sprintf("%T", origMsg)),
			)
		default:
			logger.Error("unknown type %t when attesting", winner)
			return ErrUnexpectedError.WrapS("unknown type %t when attesting", winner)
		}

	case *types.Message_UpdateValset:
		switch winner := evidence.(type) {
		case *types.TxExecutedProof:
		// check if the correct valset was updated
		case *types.SmartContractExecutionErrorProof:
			keeperutil.EmitEvent(k, ctx, types.SmartContractExecutionFailedKey,
				types.SmartContractExecutionFailedMessageID.With(fmt.Sprintf("%d", msg.GetId())),
				types.SmartContractExecutionFailedChainReferenceID.With(chainReferenceID),
				types.SmartContractExecutionFailedError.With(winner.GetErrorMessage()),
				types.SmartContractExecutionMessageType.With(fmt.Sprintf("%T", origMsg)),
			)
		default:
			return ErrUnexpectedError.WrapS("unknown type %t when attesting", winner)
		}

		// Set the snapshot as active for this chain
		err := k.Valset.SetSnapshotOnChain(ctx, origMsg.UpdateValset.Valset.ValsetID, chainReferenceID)
		if err != nil {
			// We don't want to break here, so we'll just log the error and continue
			logger.WithError(err).Error("Failed to set snapshot as active for chain", "valsetID", origMsg.UpdateValset.Valset.ValsetID)
		}

		// now remove all older update valsets given that new one was uploaded.
		// if there are any, that is.
		keeperutil.EmitEvent(k, ctx, types.AttestingUpdateValsetRemoveOldMessagesKey)
		msgs, err := q.GetAll(ctx)
		if err != nil {
			return err
		}
		for _, oldMessage := range msgs {

			actionMsg, err := oldMessage.ConsensusMsg(k.cdc)
			if err != nil {
				return err
			}
			if _, ok := (actionMsg.(*types.Message).GetAction()).(*types.Message_UpdateValset); ok {
				if oldMessage.GetId() < msg.GetId() {
					if err := q.Remove(ctx, oldMessage.GetId()); err != nil {
						logger.WithError(err).Error("error removing old message, attestRouter", "msg-id", oldMessage.GetId(), "msg-nonce", oldMessage.Nonce())
					}
				}
			}
		}
	case *types.Message_SubmitLogicCall:
		switch winner := evidence.(type) {
		case *types.TxExecutedProof:
		// check if correct thing was called
		case *types.SmartContractExecutionErrorProof:
			keeperutil.EmitEvent(k, ctx, types.SmartContractExecutionFailedKey,
				types.SmartContractExecutionFailedMessageID.With(fmt.Sprintf("%d", msg.GetId())),
				types.SmartContractExecutionFailedChainReferenceID.With(chainReferenceID),
				types.SmartContractExecutionFailedError.With(winner.GetErrorMessage()),
				types.SmartContractExecutionMessageType.With(fmt.Sprintf("%T", origMsg)),
			)

			rawMsg, ok := consensusMsg.(*types.Message)
			if !ok {
				return nil
			}

			slc := origMsg.SubmitLogicCall
			if slc.Retries < cMaxSubmitLogicCallRetries {
				slc.Retries++
				logger.Info("retrying failed SubmitLogicCall message",
					"message-id", msg.GetId(),
					"retries", slc.Retries,
					"chain-reference-id", chainReferenceID)
				newMsgID, err := k.AddSmartContractExecutionToConsensus(ctx, chainReferenceID, rawMsg.GetTurnstoneID(), slc)
				if err != nil {
					logger.WithError(err).Error("Failed to retry SubmitLogicCall")
				} else {
					logger.Info("retried failed SubmitLogicCall message",
						"message-id", msg.GetId(),
						"new-message-id", newMsgID,
						"retries", slc.Retries,
						"chain-reference-id", chainReferenceID)
				}
			} else {
				logger.Error("max retries for message reached",
					"message-id", msg.GetId(),
					"retries", slc.Retries,
					"chain-reference-id", chainReferenceID)
			}
		default:
			return ErrUnexpectedError.WrapS("unknown type %t when attesting", winner)
		}
	}

	return nil
}

func (k Keeper) findEvidenceThatWon(
	ctx sdk.Context,
	evidences []*consensustypes.Evidence,
) (any, error) {
	snapshot, err := k.Valset.GetCurrentSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	// check if there is enough power to reach the consensus
	// in the best case scenario
	var cp consensusPower
	cp.setTotal(snapshot.TotalShares)

	for _, evidence := range evidences {
		val, found := snapshot.GetValidator(evidence.GetValAddress())
		if !found {
			continue
		}
		cp.add(val.ShareCount)
	}

	if !cp.consensus() {
		return nil, ErrConsensusNotAchieved
	}

	groups := make(map[string]struct {
		evidence   types.Hashable
		validators []sdk.ValAddress
	})

	var g whoops.Group
	for _, evidence := range evidences {
		rawProof := evidence.GetProof()
		var hashable types.Hashable
		err := k.cdc.UnpackAny(rawProof, &hashable)
		if err != nil {
			return nil, err
		}

		bytesToHash, err := hashable.BytesToHash()
		if err != nil {
			return nil, err
		}
		hash := hex.EncodeToString(hashSha256(bytesToHash))
		val := groups[hash]
		if val.evidence == nil {
			val.evidence = hashable
		}
		val.validators = append(val.validators, evidence.ValAddress)
		groups[hash] = val
	}

	// TODO: gas management
	// TODO: punishing validators who misbehave
	// TODO: check for every tx if it seems genuine

	for _, group := range slice.FromMapValues(groups) {

		var cp consensusPower
		cp.setTotal(snapshot.TotalShares)

		for _, val := range group.validators {
			snapshotVal, ok := snapshot.GetValidator(val)
			if !ok {
				// strange...
				continue
			}
			cp.add(snapshotVal.ShareCount)
		}

		if cp.consensus() {
			// consensus reached
			return group.evidence, nil
		}

		// TODO: punish other validators that are a part of different groups?
	}

	if g.Err() {
		return nil, g
	}

	return nil, ErrConsensusNotAchieved
}

func (k Keeper) initiateERC20TokenOwnershipTransfer(
	ctx sdk.Context,
	chainReferenceID,
	turnstoneID string,
	transfer *types.TransferERC20Ownership,
) error {
	assignee, err := k.PickValidatorForMessage(ctx, chainReferenceID, nil)
	if err != nil {
		return err
	}

	msgID, err := k.ConsensusKeeper.PutMessageInQueue(
		ctx,
		consensustypes.Queue(
			ConsensusTurnstoneMessage,
			xchainType,
			chainReferenceID,
		),
		&types.Message{
			ChainReferenceID: chainReferenceID,
			TurnstoneID:      turnstoneID,
			Action: &types.Message_TransferERC20Ownership{
				TransferERC20Ownership: transfer,
			},
			Assignee: assignee,
		}, nil)
	if err != nil {
		return err
	}

	liblog.FromSDKLogger(k.Logger(ctx)).WithFields("new-message-id", msgID).Debug("initiateERC20TokenOwnershipTransfer triggered")
	return nil
}

func (k Keeper) txAlreadyProcessedStore(ctx sdk.Context) prefix.Store {
	kv := ctx.KVStore(k.storeKey)
	return prefix.NewStore(kv, []byte("tx-processed"))
}

func (k Keeper) setTxAsAlreadyProcessed(ctx sdk.Context, tx *ethtypes.Transaction) {
	kv := k.txAlreadyProcessedStore(ctx)
	kv.Set(tx.Hash().Bytes(), []byte{1})
}

func (k Keeper) isTxProcessed(ctx sdk.Context, tx *ethtypes.Transaction) bool {
	kv := k.txAlreadyProcessedStore(ctx)
	return kv.Has(tx.Hash().Bytes())
}
