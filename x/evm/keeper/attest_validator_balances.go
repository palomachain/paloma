package keeper

import (
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
)

func (k Keeper) attestValidatorBalances(ctx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) (retErr error) {

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

	request := consensusMsg.(*types.ValidatorBalancesAttestation)

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

	_, chainReferenceID := q.ChainInfo()
	ci, err := k.GetChainInfo(ctx, chainReferenceID)
	if err != nil {
		return err
	}

	ci.GetMinOnChainBalance()

	switch winner := evidence.(type) {
	case *types.ValidatorBalancesAttestationRes:
		for i := range request.GetHexAddresses() {
			valAddr := request.ValAddresses[i]
			hexAddr, balanceStr := common.AddressFromHex(winner.HexAddresses[i]), winner.Balances[i]
			balance, ok := new(big.Int).SetString(balanceStr, 10)
			if !ok {
				// shit
			}

			k.setValidatorBalance(ctx, valAddr, hexAddrchainReferenceID, balance)
			if balance < minCoinsThreshold {
				// jail
				k.valset.Jail(valAddr, fmt.Sprintf(types.JailReasonNotEnoughFunds, chainReferenceID, balanceStr, minCoinsThreshold))
			}
		}
	default:
		return ErrUnexpectedError.WrapS("unknown type %t when attesting", winner)
	}

	return nil
}
