package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/palomachain/paloma/util/slice"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/vizualni/whoops"
)

func hashSha256(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

type consensusPower struct {
	runningSum sdk.Int
	totalPower sdk.Int
}

func (c *consensusPower) setTotal(total sdk.Int) {
	c.totalPower = total
}

func (c *consensusPower) add(power sdk.Int) {
	var zero sdk.Int
	if c.runningSum == zero {
		c.runningSum = sdk.NewInt(0)
	}
	c.runningSum = c.runningSum.Add(power)
}

func (c *consensusPower) consensus() bool {
	var zero sdk.Int
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

func (k Keeper) attestRouter(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) (retErr error) {
	k.Logger(ctx).Debug("attest-router", "msg-id", msg.GetId(), "msg-nonce", msg.Nonce())
	if len(msg.GetEvidence()) == 0 {
		return nil
	}

	ctx, writeCache := ctx.CacheContext()
	defer func() {
		if retErr == nil {
			writeCache()
		}
	}()

	consensusMsg, err := msg.ConsensusMsg(k.cdc)
	if err != nil {
		return err
	}

	evidence, err := k.findEvidenceThatWon(ctx, msg.GetEvidence())
	if err != nil {
		if errors.Is(err, ErrConsensusNotAchieved) {
			return nil
		}
		return err
	}

	defer func() {
		// given that there was enough evidence for a proof, regardless of the outcome,
		// we should remove this from the queue as there isn't much that we can do about it.
		q.Remove(ctx, msg.GetId())
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

	switch origMsg := actionMsg.(type) {
	case *types.Message_UploadSmartContract:
		defer func() {
			// regardless of the outcome, this upload/deployment should be removed
			k.RemoveSmartContractDeployment(ctx, origMsg.UploadSmartContract.GetId(), chainReferenceID)
		}()
		switch winner := evidence.(type) {
		case *types.TxExecutedProof:
			tx, err := winner.GetTX()
			if err != nil {
				return err
			}
			if k.isTxProcessed(ctx, tx) {
				// somebody submitted the old transaction that was already processed?
				// punish those validators!!
				return ErrUnexpectedError.WrapS("transaction %s is already processed", tx.Hash())
			}
			err = origMsg.UploadSmartContract.VerifyAgainstTX(tx)
			if err != nil {
				// passed in transaction doesn't seem to be created from this smart contract
				return err
			}
			if err != nil {
				return err
			}

			smartContract, err := k.getSmartContract(ctx, origMsg.UploadSmartContract.GetId())

			if err != nil {
				return err
			}

			ethMsg, err := tx.AsMessage(ethtypes.NewLondonSigner(tx.ChainId()), big.NewInt(0))
			if err != nil {
				return err
			}

			smartContractAddr := crypto.CreateAddress(ethMsg.From(), tx.Nonce()).Hex()

			if err != nil {
				return err
			}

			deployingSmartContract, _ := k.getSmartContractDeploying(ctx, origMsg.UploadSmartContract.GetId(), chainReferenceID)
			if deployingSmartContract == nil {
				return ErrCannotActiveSmartContractThatIsNotDeploying
			}

			err = k.ActivateChainReferenceID(
				ctx,
				chainReferenceID,
				smartContract,
				smartContractAddr,
				deployingSmartContract.GetUniqueID(),
			)

			if err != nil {
				return err
			}

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
					q.Remove(ctx, oldMessage.GetId())
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

	// TODO: gas managment
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
