package keeper

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
)

const (
	cErc20HandoverABI  string = `[{"inputs": [{"name": "_compass","type": "address"}],"name": "new_compass","outputs": [],"stateMutability": "nonpayable","type": "function"}]`
	cFeeMgrHandoverABI string = `[{"inputs":[{"name":"_new_compass","type":"address"}],"name":"update_compass","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
)

type attestionParameters struct {
	originalMessage  consensustypes.QueuedSignedMessageI
	rawEvidence      any
	msg              *types.Message
	chainReferenceID string
	msgID            uint64
}

type uploadSmartContractAttester struct {
	attestionParameters
	action *types.UploadSmartContract
	logger liblog.Logr
	k      *Keeper
}

func newUploadSmartContractAttester(k *Keeper, l liblog.Logr, p attestionParameters) *uploadSmartContractAttester {
	return &uploadSmartContractAttester{
		attestionParameters: p,
		logger:              l,
		k:                   k,
	}
}

func (a *uploadSmartContractAttester) Execute(ctx sdk.Context) error {
	a.logger = a.logger.WithFields("action-msg", "Message_UploadSmartContract")
	a.logger.Debug("Processing upload smart contract message attestation.")

	a.action = a.msg.Action.(*types.Message_UploadSmartContract).UploadSmartContract

	switch evidence := a.rawEvidence.(type) {
	case *types.TxExecutedProof:
		return a.attest(ctx, evidence)
	case *types.SmartContractExecutionErrorProof:
		a.logger.Debug("smart contract execution error proof", "smart-contract-error", evidence.GetErrorMessage())
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

func (a *uploadSmartContractAttester) attest(ctx sdk.Context, evidence *types.TxExecutedProof) error {
	tx, err := attestTransactionIntegrity(ctx, a.originalMessage, a.k, evidence,
		a.chainReferenceID, a.msg.AssigneeRemoteAddress, a.action.VerifyAgainstTX)
	if err != nil {
		a.logger.WithError(err).Error("Failed to verify transaction integrity.")
		return err
	}

	smartContractID := a.action.GetId()
	deployment, _ := a.k.getSmartContractDeploymentByContractID(ctx, smartContractID, a.chainReferenceID)
	if deployment == nil {
		a.logger.WithError(err).WithFields("smart-contract-id", smartContractID).Error("Smart contract not found")
		return ErrCannotActiveSmartContractThatIsNotDeploying
	}

	if deployment.GetStatus() != types.SmartContractDeployment_IN_FLIGHT {
		a.logger.WithError(err).Error("deployment not in right state")
		return ErrCannotActiveSmartContractThatIsNotDeploying
	}

	// Update deployment
	ethMsg, err := core.TransactionToMessage(tx, ethtypes.NewLondonSigner(tx.ChainId()), big.NewInt(0))
	if err != nil {
		a.logger.WithError(err).Error("Failed to extract ethMsg")
		return err
	}
	newCompassAddr := crypto.CreateAddress(ethMsg.From, tx.Nonce())
	deployment.NewSmartContractAddress = newCompassAddr.Hex()
	deployment.Status = types.SmartContractDeployment_WAITING_FOR_ERC20_OWNERSHIP_TRANSFER
	if err := a.k.updateSmartContractDeployment(ctx, smartContractID, a.chainReferenceID, deployment); err != nil {
		a.logger.WithError(err).Error("Failed to update smart contract deployment")
		return err
	}

	// If this is the first deployment on chain, it won't have a snapshot
	// assigned to it. We need to assign it a snapshot, or we won't be able to
	// do SubmitLogicCall or UpdateValset
	_, err = a.k.Valset.GetLatestSnapshotOnChain(ctx, a.chainReferenceID)
	if err != nil {
		if !errors.Is(err, keeperutil.ErrNotFound) {
			return err
		}

		snapshot, err := a.k.Valset.GetCurrentSnapshot(ctx)
		if err != nil {
			return err
		}

		err = a.k.Valset.SetSnapshotOnChain(ctx, snapshot.Id, a.chainReferenceID)
		if err != nil {
			return err
		}

		// Since this is the intial deployment, we need to set the smart contract
		// as active and can skip the handover process.
		return a.k.SetSmartContractAsActive(ctx, smartContractID, a.chainReferenceID)
	}

	// If we got this far, it means we have a snapshot on chain and we can
	// proceed with the handover process.
	return a.startCompassHandover(ctx, newCompassAddr)
}

func (a *uploadSmartContractAttester) startCompassHandover(
	ctx sdk.Context,
	newCompassAddr common.Address,
) error {
	ci, err := a.k.GetChainInfo(ctx, a.chainReferenceID)
	if err != nil {
		return fmt.Errorf("get all chain infos: %w", err)
	}
	tokens, err := a.k.Skyway.CastChainERC20ToDenoms(ctx, a.chainReferenceID)
	if err != nil {
		return fmt.Errorf("cast chain erc20 to denoms: %w", err)
	}
	forwardCallArgs := make([]types.CompassHandover_ForwardCallArgs, 0, len(tokens)+1) // +1 for the fee manager handover
	erc20abi, err := abi.JSON(strings.NewReader(cErc20HandoverABI))
	if err != nil {
		return fmt.Errorf("parse erc20 abi json: %w", err)
	}
	erc20payload, err := erc20abi.Pack("new_compass", newCompassAddr)
	if err != nil {
		return fmt.Errorf("pack erc20 payload: %w", err)
	}
	feemgrabi, err := abi.JSON(strings.NewReader(cFeeMgrHandoverABI))
	if err != nil {
		return fmt.Errorf("parse fee manager abi json: %w", err)
	}
	feemgrpayload, err := feemgrabi.Pack("update_compass", newCompassAddr)
	if err != nil {
		return fmt.Errorf("pack fee manager payload: %w", err)
	}

	for _, v := range tokens {
		forwardCallArgs = append(forwardCallArgs, types.CompassHandover_ForwardCallArgs{
			HexContractAddress: v.GetErc20(),
			Payload:            erc20payload,
		})
	}

	forwardCallArgs = append(forwardCallArgs, types.CompassHandover_ForwardCallArgs{
		HexContractAddress: ci.FeeManagerAddr,
		Payload:            feemgrpayload,
	})

	_, err = a.k.scheduleCompassHandover(
		ctx,
		a.chainReferenceID,
		string(ci.GetSmartContractUniqueID()),
		&types.CompassHandover{
			Id:              a.action.GetId(),
			ForwardCallArgs: forwardCallArgs,
			Deadline:        ctx.BlockTime().Add(10 * time.Minute).Unix(),
		},
	)
	if err != nil {
		return fmt.Errorf("try compass handover: %w", err)
	}

	a.logger.Debug("attestation successful")
	return nil
}

func (a *uploadSmartContractAttester) attemptRetry(ctx sdk.Context) {
	contract := a.action

	if contract.Retries >= cMaxSubmitLogicCallRetries {
		a.logger.Error("max retries for UploadSmartContract message reached",
			"message-id", a.msgID,
			"retries", contract.Retries,
			"chain-reference-id", a.chainReferenceID)

		smartContractID := a.action.GetId()
		a.k.DeleteSmartContractDeploymentByContractID(ctx, smartContractID, a.chainReferenceID)
		return
	}

	contract.Retries++

	a.logger.Info("retrying failed UploadSmartContract message",
		"message-id", a.msgID,
		"retries", contract.Retries,
		"chain-reference-id", a.chainReferenceID)

	newMsgID, err := a.k.AddUploadSmartContractToConsensus(ctx, a.chainReferenceID, contract)
	if err != nil {
		a.logger.WithError(err).Error("Failed to retry UploadSmartContract")
		return
	}

	a.logger.Info("retried failed UploadSmartContract message",
		"message-id", a.msgID,
		"new-message-id", newMsgID,
		"retries", contract.Retries,
		"chain-reference-id", a.chainReferenceID)
}
