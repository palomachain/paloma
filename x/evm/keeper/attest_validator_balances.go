package keeper

import (
	"context"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/palomachain/paloma/v2/util/liblog"
	"github.com/palomachain/paloma/v2/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/v2/x/consensus/types"
	"github.com/palomachain/paloma/v2/x/evm/types"
)

func (k Keeper) attestValidatorBalances(ctx context.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) (retErr error) {
	return k.attestMessageWrapper(ctx, q, msg, k.validatorBalancesAttester)
}

func (k Keeper) validatorBalancesAttester(sdkCtx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI, winner any) error {
	consensusMsg, err := msg.ConsensusMsg(k.cdc)
	if err != nil {
		return err
	}

	request := consensusMsg.(*types.ValidatorBalancesAttestation)

	_, chainReferenceID := q.ChainInfo()
	ci, err := k.GetChainInfo(sdkCtx, chainReferenceID)
	if err != nil {
		return err
	}

	minBalance, err := ci.GetMinOnChainBalanceBigInt()
	if err != nil {
		return err
	}

	return k.processValidatorBalanceProof(sdkCtx, request, winner, chainReferenceID, minBalance)
}

func (k Keeper) processValidatorBalanceProof(
	ctx sdk.Context,
	request *types.ValidatorBalancesAttestation,
	evidence any,
	chainReferenceID string,
	minBalance *big.Int,
) error {
	switch winner := evidence.(type) {
	case *types.ValidatorBalancesAttestationRes:
		for i := range request.GetHexAddresses() {
			valAddr := request.ValAddresses[i]
			valAddrString, err := k.AddressCodec.BytesToString(valAddr)
			if err != nil {
				k.Logger(ctx).Error("error while getting validator address", err)
			}

			hexAddr, balanceStr := common.HexToAddress(request.HexAddresses[i]), winner.Balances[i]
			balance, ok := new(big.Int).SetString(balanceStr, 10)
			if !ok {
				k.Logger(ctx).Error(
					"invalid balance string when attesting to EVM balance",
					"balance", balanceStr,
					"val-addr", valAddr,
					"eth-addr", hexAddr,
				)

				reason := fmt.Sprintf(types.JailReasonInvalidBalance, chainReferenceID, balanceStr)
				if err = k.Valset.Jail(ctx, valAddr, reason); err != nil {
					k.Logger(ctx).Error(
						"error jailing validator",
						"err", err,
						"val-addr", valAddr,
						"eth-addr", hexAddr,
					)
				}

				continue
			}

			if err := k.Valset.SetValidatorBalance(ctx, valAddr, "evm", chainReferenceID, hexAddr.String(), balance); err != nil {
				k.Logger(ctx).Error(
					"error setting validator balance",
					"err", err,
					"val-addr", valAddr,
					"eth-addr", hexAddr,
				)
			}

			if balance.Cmp(minBalance) == -1 || balance.Cmp(big.NewInt(0)) == 0 {
				isJailed, err := k.Valset.IsJailed(ctx, valAddr)
				if err != nil {
					liblog.FromSDKLogger(k.Logger(ctx)).WithError(err).WithValidator(valAddrString).WithFields("val-addr", valAddr, "hex-addr", hexAddr).Error("attestValidatorBalances: error in checking jailed validator")
				}

				if !isJailed {
					if err := k.Valset.Jail(ctx, valAddr, fmt.Sprintf(types.JailReasonNotEnoughFunds, chainReferenceID, balanceStr, minBalance)); err != nil {
						k.Logger(ctx).Error(
							"error jailing validator",
							"err", err,
							"val-addr", valAddr,
							"eth-addr", hexAddr,
						)
					}
				}
			}
		}
	default:
		return ErrUnexpectedError.JoinErrorf("unknown type %t when attesting", winner)
	}

	return nil
}
