package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/consensus/keeper/consensus"
	consensustypes "github.com/palomachain/paloma/v2/x/consensus/types"
	"github.com/palomachain/paloma/v2/x/evm/types"
)

func (k Keeper) attestReferenceBlock(ctx context.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI) (retErr error) {
	return k.attestMessageWrapper(ctx, q, msg, k.referenceBlockAttester)
}

func (k Keeper) referenceBlockAttester(sdkCtx sdk.Context, q consensus.Queuer, msg consensustypes.QueuedSignedMessageI, winner any) error {
	switch ref := winner.(type) {
	case *types.ReferenceBlockAttestationRes:
		_, chainReferenceID := q.ChainInfo()
		logger := k.Logger(sdkCtx).WithFields(
			"msg-id", msg.GetId(),
			"block-height", ref.BlockHeight,
			"block-hash", ref.BlockHash,
		)
		logger.Debug("Changing chain reference block.")
		err := k.UpdateChainReferenceBlock(sdkCtx, chainReferenceID, ref.BlockHeight, ref.BlockHash)
		if err != nil {
			logger.WithError(err).Error("Error updating chain reference block")
		}
		return err
	default:
		return ErrUnexpectedError.JoinErrorf("unknown type %t when attesting", winner)
	}
}
