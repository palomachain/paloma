package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/consensus/keeper/consensus"
	"github.com/palomachain/paloma/x/evm/types"
)

type updateValsetAttester struct {
	attestionParameters
	action *types.UpdateValset
	q      consensus.Queuer
	logger liblog.Logr
	k      *Keeper
}

func newUpdateValsetAttester(k *Keeper, l liblog.Logr, q consensus.Queuer, p attestionParameters) *updateValsetAttester {
	return &updateValsetAttester{
		attestionParameters: p,
		logger:              l,
		k:                   k,
		q:                   q,
	}
}

func (a *updateValsetAttester) Execute(ctx sdk.Context) error {
	a.logger = a.logger.WithFields("action-msg", "Message_UpdateValset")
	a.logger.Debug("Processing update valset message attestation.")

	a.action = a.msg.Action.(*types.Message_UpdateValset).UpdateValset

	switch winner := a.rawEvidence.(type) {
	case *types.TxExecutedProof:
		return a.attest(ctx, winner)
	case *types.SmartContractExecutionErrorProof:
		keeperutil.EmitEvent(a.k, ctx, types.SmartContractExecutionFailedKey,
			types.SmartContractExecutionFailedMessageID.With(fmt.Sprintf("%d", a.msgID)),
			types.SmartContractExecutionFailedChainReferenceID.With(a.chainReferenceID),
			types.SmartContractExecutionFailedError.With(winner.GetErrorMessage()),
			types.SmartContractExecutionMessageType.With(fmt.Sprintf("%T", a.action)),
		)
	default:
		return ErrUnexpectedError.JoinErrorf("unknown type %t when attesting", winner)
	}

	return nil
}

func (a *updateValsetAttester) attest(ctx sdk.Context, evidence *types.TxExecutedProof) error {
	// Set the snapshot as active for this chain
	err := a.k.Valset.SetSnapshotOnChain(ctx, a.action.Valset.ValsetID, a.chainReferenceID)
	if err != nil {
		// We don't want to break here, so we'll just log the error and continue
		a.logger.WithError(err).Error("Failed to set snapshot as active for chain", "valsetID", a.action.Valset.ValsetID)
	}

	// now remove all older update valsets given that new one was uploaded.
	// if there are any, that is.
	keeperutil.EmitEvent(a.k, ctx, types.AttestingUpdateValsetRemoveOldMessagesKey)
	msgs, err := a.q.GetAll(ctx)
	if err != nil {
		return err
	}
	for _, oldMessage := range msgs {

		actionMsg, err := oldMessage.ConsensusMsg(a.k.cdc)
		if err != nil {
			return err
		}
		if _, ok := (actionMsg.(*types.Message).GetAction()).(*types.Message_UpdateValset); ok {
			if oldMessage.GetId() < a.msgID {
				if err := a.q.Remove(ctx, oldMessage.GetId()); err != nil {
					a.logger.WithError(err).Error("error removing old message, attestRouter", "msg-id", oldMessage.GetId(), "msg-nonce", oldMessage.Nonce())
				}
			}
		}
	}

	return nil
}
