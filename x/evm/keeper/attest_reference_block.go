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

func (k Keeper) referenceBlockAttester(sdkCtx sdk.Context, q consensus.Queuer, _ consensustypes.QueuedSignedMessageI, winner any) error {
	_, chainReferenceID := q.ChainInfo()

	switch ref := winner.(type) {
	case *types.ReferenceBlockAttestationRes:
		k.Logger(sdkCtx).WithFields(
			"block-height", ref.BlockHeight,
			"block-hash", ref.BlockHash,
		).Debug("Changing chain reference block.")
		return k.UpdateChainReferenceBlock(sdkCtx, chainReferenceID, ref.BlockHeight, ref.BlockHash)
	default:
		return ErrUnexpectedError.JoinErrorf("unknown type %t when attesting", winner)
	}
}
