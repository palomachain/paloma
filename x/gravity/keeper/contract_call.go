package keeper

import (
	"bytes"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
)

func (k Keeper) contractCallExecuted(ctx sdk.Context, invalidationScope []byte, invalidationNonce uint64) {
	otx := k.GetOutgoingTx(ctx, types.MakeContractCallTxKey(invalidationScope, invalidationNonce))
	if otx == nil {
		k.Logger(ctx).Error("Failed to clean contract calls",
			"invalidation scope", hex.EncodeToString(invalidationScope),
			"invalidation nonce", invalidationNonce)
		return
	}

	completedCallTx, _ := otx.(*types.ContractCallTx)
	k.IterateOutgoingTxsByType(ctx, types.ContractCallTxPrefixByte, func(key []byte, otx types.OutgoingTx) bool {
		// If the iterated contract call's nonce is lower than the one that was just executed, delete it
		cctx, _ := otx.(*types.ContractCallTx)
		if (cctx.InvalidationNonce < completedCallTx.InvalidationNonce) &&
			bytes.Equal(cctx.InvalidationScope, completedCallTx.InvalidationScope) {
			k.DeleteOutgoingTx(ctx, cctx.GetStoreIndex())
		}
		return false
	})

	k.DeleteOutgoingTx(ctx, completedCallTx.GetStoreIndex())
}
