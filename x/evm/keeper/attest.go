package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
	metrixtypes "github.com/palomachain/paloma/x/metrix/types"
)

const (
	cMaxSubmitLogicCallRetries     uint32 = 2
	cMaxUploadSmartContractRetries uint32 = 2
)

type msgAttester func(sdk.Context, consensus.Queuer, consensustypes.QueuedSignedMessageI, any) error

func (k Keeper) attestMessageWrapper(ctx context.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI, fn msgAttester) (retErr error) {
	if len(msg.GetEvidence()) == 0 {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(ctx).WithFields(
		"component", "attest-message",
		"msg-id", msg.GetId(),
		"msg-nonce", msg.Nonce())
	logger.Debug("attest-message")

	cacheCtx, writeCache := sdkCtx.CacheContext()
	defer func() {
		if retErr == nil {
			writeCache()
		}
	}()

	result, err := k.consensusChecker.VerifyEvidence(cacheCtx, msg.GetEvidence())
	if err != nil {
		if errors.Is(err, ErrConsensusNotAchieved) {
			logger.WithFields(
				"total-shares", result.TotalShares,
				"total-votes", result.TotalVotes,
				"distribution", result.Distribution,
			).WithError(err).Error("Consensus not achieved.")
			return nil
		}
		return err
	}

	defer func() {
		// given that there was enough evidence for a proof, regardless of the outcome,
		// we should remove this from the queue as there isn't much that we can do about it.
		if err := q.Remove(cacheCtx, msg.GetId()); err != nil {
			k.Logger(sdkCtx).Error("error removing message, attestMessage", "msg-id", msg.GetId(), "msg-nonce", msg.Nonce())
		}
	}()

	return fn(cacheCtx, q, msg, result.Winner)
}

func (k Keeper) attestRouter(ctx context.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) (err error) {
	return k.attestMessageWrapper(ctx, q, msg, k.routerAttester)
}

func (k Keeper) routerAttester(sdkCtx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI, winner any) error {
	consensusMsg, err := msg.ConsensusMsg(k.cdc)
	if err != nil {
		k.Logger(sdkCtx).WithError(err).Error("failed to cast to consensus message")
		return err
	}

	message := consensusMsg.(*types.Message)

	defer func() {
		success := false

		// if the input type is a TX, then regardles, we want to set it as already processed
		switch winner := winner.(type) {
		case *types.TxExecutedProof:
			success = true
			tx, err := winner.GetTX()
			if err == nil {
				k.setTxAsAlreadyProcessed(sdkCtx, tx)
			}
		}

		handledAt := msg.GetHandledAtBlockHeight()
		if handledAt == nil {
			handledAt = func(i math.Int) *math.Int { return &i }(sdkmath.NewInt(sdkCtx.BlockHeight()))
		}
		publishMessageAttestedEvent(sdkCtx, &k, msg.GetId(), message.Assignee, message.AssignedAtBlockHeight, *handledAt, success)
	}()

	rawAction := message.GetAction()
	_, chainReferenceID := q.ChainInfo()
	logger := k.Logger(sdkCtx).WithFields("chain-reference-id", chainReferenceID)

	params := attestionParameters{
		msgID:            msg.GetId(),
		chainReferenceID: chainReferenceID,
		rawEvidence:      winner,
		msg:              message,
	}
	switch rawAction.(type) {
	case *types.Message_UploadSmartContract:
		return newUploadSmartContractAttester(&k, logger, params).Execute(sdkCtx)
	case *types.Message_UpdateValset:
		return newUpdateValsetAttester(&k, logger, q, params).Execute(sdkCtx)
	case *types.Message_SubmitLogicCall:
		return newSubmitLogicCallAttester(&k, logger, params).Execute(sdkCtx)
	}

	return nil
}

func publishMessageAttestedEvent(ctx context.Context, k *Keeper, msgID uint64, assignee string, assignedAt math.Int, handledAt math.Int, successful bool) {
	valAddr, err := sdk.ValAddressFromBech32(assignee)
	if err != nil {
		liblog.FromSDKLogger(k.Logger(ctx)).WithError(err).WithFields("assignee", assignee, "msg-id", msgID).Error("failed to get validator address from bech32.")
	}

	for _, v := range k.onMessageAttestedListeners {
		v.OnConsensusMessageAttested(ctx, metrixtypes.MessageAttestedEvent{
			AssignedAtBlockHeight:  assignedAt,
			HandledAtBlockHeight:   handledAt,
			Assignee:               valAddr,
			MessageID:              msgID,
			WasRelayedSuccessfully: successful,
		})
	}
}

func attestTransactionIntegrity(
	ctx context.Context,
	k *Keeper,
	proof *types.TxExecutedProof,
	verifyTx func(tx *ethtypes.Transaction) error,
) (*ethtypes.Transaction, error) {
	// check if correct thing was called
	tx, err := proof.GetTX()
	if err != nil {
		return nil, fmt.Errorf("failed to get TX: %w", err)
	}
	if k.isTxProcessed(ctx, tx) {
		// somebody submitted the old transaction that was already processed?
		// punish those validators!!
		return nil, ErrUnexpectedError.JoinErrorf("transaction %s is already processed", tx.Hash())
	}
	err = verifyTx(tx)
	if err != nil {
		// passed in transaction doesn't seem to be created from this smart contract
		return nil, fmt.Errorf("tx failed to verify: %w", err)
	}

	return tx, nil
}

func (k Keeper) SetSmartContractAsActive(ctx context.Context, smartContractID uint64, chainReferenceID string) (err error) {
	logger := liblog.FromSDKLogger(k.Logger(ctx))
	defer func() {
		if err == nil {
			logger.With("smart-contract-id", smartContractID).Debug("removing deployment.")
			k.DeleteSmartContractDeploymentByContractID(ctx, smartContractID, chainReferenceID)
		}
	}()

	deployment, _ := k.getSmartContractDeploymentByContractID(ctx, smartContractID, chainReferenceID)
	if deployment.GetStatus() != types.SmartContractDeployment_WAITING_FOR_ERC20_OWNERSHIP_TRANSFER {
		logger.WithError(err).Error("Deployment not awaiting transfer")
		return ErrCannotActiveSmartContractThatIsNotDeploying
	}

	smartContract, err := k.getSmartContract(ctx, deployment.GetSmartContractID())
	if err != nil {
		logger.WithError(err).Error("Failed to get contract")
		return err
	}

	err = k.ActivateChainReferenceID(
		ctx,
		chainReferenceID,
		smartContract,
		deployment.NewSmartContractAddress,
		deployment.GetUniqueID(),
	)
	if err != nil {
		logger.WithError(err).Error("Failed to activate chain")
		return err
	}

	return nil
}

func (k Keeper) txAlreadyProcessedStore(ctx context.Context) storetypes.KVStore {
	s := runtime.KVStoreAdapter(k.storeKey.OpenKVStore(ctx))
	return prefix.NewStore(s, []byte("tx-processed"))
}

func (k Keeper) setTxAsAlreadyProcessed(ctx context.Context, tx *ethtypes.Transaction) {
	kv := k.txAlreadyProcessedStore(ctx)
	kv.Set(tx.Hash().Bytes(), []byte{1})
}

func (k Keeper) isTxProcessed(ctx context.Context, tx *ethtypes.Transaction) bool {
	kv := k.txAlreadyProcessedStore(ctx)
	return kv.Has(tx.Hash().Bytes())
}
