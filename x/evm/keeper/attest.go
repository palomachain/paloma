package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
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

func (k Keeper) attestRouter(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) (retErr error) {
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

	err = k.attestTransaction(ctx, msg.GetEvidence())
	if err != nil {
		if errors.Is(err, ErrConsensusNotAchieved) {
			return nil
		}
		return err
	}

	defer func() {
		q.Remove(ctx, msg.GetId())
	}()

	switch origMsg := consensusMsg.(type) {
	case *types.UploadSmartContract:
		smartContract, err := k.getSmartContract(ctx, origMsg.GetId())
		err = keeperutil.Save(k.lastSmartContractStore(ctx), k.cdc, lastSmartContractKey, smartContract)
		if err != nil {
			return err
		}
	case *types.SubmitLogicCall:
		// nothing
	case *types.UpdateValset:
		// nothing
	}

	return nil

}

func (k Keeper) attestTransaction(
	ctx sdk.Context,
	evidences []*consensustypes.Evidence,
) error {
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

		hash := hex.EncodeToString(hashSha256(evidence.GetProof()))
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

	snapshot, err := k.Valset.GetCurrentSnapshot(ctx)
	if err != nil {
		return err
	}
	for _, group := range groups {

		if k.isTxProcessed(ctx, group.tx) {
			// TODO: punish those validators??
			continue
		}

		groupTotal := big.NewInt(0)
		for _, val := range group.validators {
			snapshotVal, ok := snapshot.GetValidator(val)
			if !ok {
				// strange...
				continue
			}

			groupTotal.Add(groupTotal, snapshotVal.ShareCount.BigInt())
		}

		/*
			groupTotal >= total * 2.0 / 3.0  then consensus has been reached
			not to lose precision, we can do this:
			groupTotal * 3 >= total * 2
		*/
		var grInt *big.Int
		var totInt *big.Int

		grInt.Mul(groupTotal, big.NewInt(3))
		totInt.Mul(snapshot.TotalShares.BigInt(), big.NewInt(2))

		cmp := grInt.Cmp(totInt)
		if cmp == 0 || cmp == 1 {
			// consensus reached
			return nil
		}

		// TODO: punish other validators that are a part of different groups?
	}

	if g.Err() {
		return g
	}

	return ErrConsensusNotAchieved
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
