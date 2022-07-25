package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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

	tx, err := k.attestTransaction(ctx, msg.GetEvidence())
	if err != nil {
		if errors.Is(err, ErrConsensusNotAchieved) {
			return nil
		}
		return err
	}

	// If we got up to here it means that the enough evidence was provided
	actionMsg := consensusMsg.(*types.Message).GetAction()

	switch origMsg := actionMsg.(type) {
	case *types.Message_UploadSmartContract:
		_, chainReferenceID := q.ChainInfo()
		if err != nil {
			return err
		}

		smartContract, err := k.getSmartContract(ctx, origMsg.UploadSmartContract.GetId())

		if err != nil {
			return err
		}

		ethMsg, err := tx.AsMessage(ethtypes.NewEIP2930Signer(tx.ChainId()), big.NewInt(0))
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

		k.removeSmartContractDeployment(ctx, origMsg.UploadSmartContract.GetId(), chainReferenceID)
	case *types.Message_UpdateValset:
		// nothing
	case *types.Message_SubmitLogicCall:
		// nothing
	}

	q.Remove(ctx, msg.GetId())

	return nil

}

func (k Keeper) attestTransaction(
	ctx sdk.Context,
	evidences []*consensustypes.Evidence,
) (*ethtypes.Transaction, error) {
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
		tx         *ethtypes.Transaction
		validators []sdk.ValAddress
	})

	var g whoops.Group
	for _, evidence := range evidences {
		proof := evidence.GetProof()
		tx := &ethtypes.Transaction{}

		// is rlp deterministic?
		err := tx.UnmarshalBinary(proof)
		g.Add(err)
		if err != nil {
			continue
		}

		hash := hex.EncodeToString(hashSha256(proof))
		val := groups[hash]
		if val.tx == nil {
			val.tx = tx
		}
		val.validators = append(val.validators, evidence.ValAddress)
		groups[hash] = val
	}

	// TODO: gas managment
	// TODO: punishing validators who misbehave
	// TODO: check for every tx if it seems genuine

	for _, group := range groups {

		if k.isTxProcessed(ctx, group.tx) {
			// TODO: punish those validators??
			continue
		}

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
			k.setTxAsAlreadyProcessed(ctx, group.tx)
			return group.tx, nil
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
