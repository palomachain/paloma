package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/evm/types"
)

type compassHandoverAttester struct {
	attestionParameters
	action *types.CompassHandover
	logger liblog.Logr
	k      *Keeper
}

func newCompassHandoverAttester(k *Keeper, l liblog.Logr, p attestionParameters) *compassHandoverAttester {
	return &compassHandoverAttester{
		attestionParameters: p,
		logger:              l,
		k:                   k,
	}
}

func (a *compassHandoverAttester) Execute(ctx sdk.Context) error {
	a.logger = a.logger.WithFields("action-msg", "Message_CompassHandover")
	a.logger.Debug("Processing compass handover message attestation.")

	a.action = a.msg.Action.(*types.Message_CompassHandover).CompassHandover

	switch winner := a.rawEvidence.(type) {
	case *types.TxExecutedProof:
		return a.attest(ctx, winner)
	case *types.SmartContractExecutionErrorProof:
		a.logger.WithFields(
			"smart-contract-error", winner.GetErrorMessage()).
			Warn("CompassHandover failed")
		keeperutil.EmitEvent(a.k, ctx, types.SmartContractExecutionFailedKey,
			types.SmartContractExecutionFailedMessageID.With(fmt.Sprintf("%d", a.msgID)),
			types.SmartContractExecutionFailedChainReferenceID.With(a.chainReferenceID),
			types.SmartContractExecutionFailedError.With(winner.GetErrorMessage()),
			types.SmartContractExecutionMessageType.With(fmt.Sprintf("%T", a.action)),
		)
		return nil
	default:
		return ErrUnexpectedError.JoinErrorf("unknown type %t when attesting", winner)
	}
}

func (a *compassHandoverAttester) attest(ctx sdk.Context, evidence *types.TxExecutedProof) (err error) {
	_, err = attestTransactionIntegrity(ctx, a.originalMessage, a.k, evidence,
		a.chainReferenceID, a.msg.AssigneeRemoteAddress, a.action.VerifyAgainstTX)
	if err != nil {
		a.logger.WithError(err).Error("Failed to verify transaction integrity.")
		return err
	}

	return nil
}
