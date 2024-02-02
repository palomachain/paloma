package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
	metrixtypes "github.com/palomachain/paloma/x/metrix/types"
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
		c.runningSum = sdk.NewInt(0)
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
	return c.runningSum.Mul(sdk.NewInt(3)).GTE(
		c.totalPower.Mul(sdk.NewInt(2)),
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

	// Use transactionalsed context only!
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

	message := consensusMsg.(*types.Message)

	// If we got up to here it means that the enough evidence was provided
	defer func() {
		success := false

		// if the input type is a TX, then regardles, we want to set it as already processed
		switch winner := evidence.(type) {
		case *types.TxExecutedProof:
			success = true
			tx, err := winner.GetTX()
			if err == nil {
				k.setTxAsAlreadyProcessed(ctx, tx)
			}
		}

		handledAt := msg.GetHandledAtBlockHeight()
		if handledAt == nil {
			handledAt = func(i math.Int) *math.Int { return &i }(sdk.NewInt(ctx.BlockHeight()))
		}
		publishMessageAttestedEvent(ctx, &k, msg.GetId(), message.Assignee, message.AssignedAtBlockHeight, *handledAt, success)

		// given that there was enough evidence for a proof, regardless of the outcome,
		// we should remove this from the queue as there isn't much that we can do about it.
		if err := q.Remove(ctx, msg.GetId()); err != nil {
			logger.WithError(err).Error("error removing message, attestRouter")
		}
	}()

	rawAction := message.GetAction()
	_, chainReferenceID := q.ChainInfo()
	logger = logger.WithFields("chain-reference-id", chainReferenceID)

	params := attestionParameters{
		msgID:            msg.GetId(),
		chainReferenceID: chainReferenceID,
		rawEvidence:      evidence,
		msg:              message,
	}
	switch rawAction.(type) {
	case *types.Message_UploadSmartContract:
		return newUploadSmartContractAttester(&k, logger, params).Execute(ctx)
	case *types.Message_UpdateValset:
		return newUpdateValsetAttester(&k, logger, q, params).Execute(ctx)
	case *types.Message_SubmitLogicCall:
		return newSubmitLogicCallAttester(&k, logger, params).Execute(ctx)
	}

	return nil
}

func publishMessageAttestedEvent(ctx sdk.Context, k *Keeper, msgID uint64, assignee string, assignedAt math.Int, handledAt math.Int, successful bool) {
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
	ctx sdk.Context,
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

func (k Keeper) SetSmartContractAsActive(ctx sdk.Context, smartContractID uint64, chainReferenceID string) (err error) {
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

func (k Keeper) txAlreadyProcessedStore(ctx sdk.Context) sdk.KVStore {
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
