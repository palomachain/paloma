package keeper

import (
	"context"
	"errors"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/skyway/types"
)

func initBridgeDataFromGenesis(ctx context.Context, k Keeper, data types.GenesisState) {
	// reset batches in state
	for _, batch := range data.Batches {
		// TODO: block height?
		intBatch, err := batch.ToInternal()
		if err != nil {
			panic(sdkerrors.Wrapf(err, "unable to make batch internal: %v", batch))
		}
		err = k.StoreBatch(ctx, *intBatch)
		if err != nil {
			panic(err)
		}
	}

	// reset batch confirmations in state
	for _, conf := range data.BatchConfirms {
		conf := conf
		_, err := k.SetBatchConfirm(ctx, &conf)
		if err != nil {
			panic(err)
		}
	}

	// reset batch gas estimates in state
	for _, estimate := range data.BatchGasEstimates {
		_, err := k.SetBatchGasEstimate(ctx, &estimate)
		if err != nil {
			panic(err)
		}
	}
}

// InitGenesis starts a chain from a genesis state
func InitGenesis(ctx context.Context, k Keeper, data types.GenesisState) {
	k.SetParams(ctx, *data.Params)

	for _, nonce := range data.SkywayNonces {
		// restore various nonces, this MUST match SkywayNonces in genesis
		err := k.setLastObservedSkywayNonce(ctx, nonce.ChainReferenceId, nonce.LastObservedNonce)
		if err != nil {
			panic(err)
		}
		err = k.SetLastSlashedBatchBlock(ctx, nonce.LastSlashedBatchBlock)
		if err != nil {
			panic(err)
		}
		k.setID(ctx, nonce.LastTxPoolId, []byte(types.KeyLastTXPoolID))
		k.setID(ctx, nonce.LastBatchId, []byte(types.KeyLastOutgoingBatchID))
	}

	if len(data.SkywayNonces) == 0 {
		if err := k.SetLastSlashedBatchBlock(ctx, 0); err != nil {
			panic(err)
		}
	}

	initBridgeDataFromGenesis(ctx, k, data)

	// reset pool transactions in state
	for _, tx := range data.UnbatchedTransfers {
		intTx, err := tx.ToInternal()
		if err != nil {
			panic(sdkerrors.Wrapf(err, "invalid unbatched tx: %v", tx))
		}
		if err := k.addUnbatchedTX(ctx, intTx); err != nil {
			panic(err)
		}
	}

	// reset attestations in state
	for _, att := range data.Attestations {
		att := att
		claim, err := k.UnpackAttestationClaim(&att)
		if err != nil {
			panic("couldn't cast to claim")
		}

		// TODO: block height?
		hash, err := claim.ClaimHash()
		if err != nil {
			panic(fmt.Errorf("error when computing ClaimHash for %v", hash))
		}
		k.SetAttestation(ctx, claim.GetChainReferenceId(), claim.GetSkywayNonce(), hash, &att)
	}

	// reset attestation state of specific validators
	// this must be done after the above to be correct
	for _, att := range data.Attestations {
		att := att
		claim, err := k.UnpackAttestationClaim(&att)
		if err != nil {
			panic("couldn't cast to claim")
		}
		/*
			reconstruct the latest event nonce for every validator
			if somehow this genesis state is saved when all attestations
			have been cleaned up GetLastEventNonceByValidator handles that case

			if we were to save and load the last event nonce for every validator
			then we would need to carry that state forever across all chain restarts
			but since we've already had to handle the edge case of new validators joining
			while all attestations have already been cleaned up we can do this instead and
			not carry around every validator's event nonce counter forever.
		*/
		for _, vote := range att.Votes {
			val, err := sdk.ValAddressFromBech32(vote)
			if err != nil {
				panic(err)
			}
			last, err := k.GetLastSkywayNonceByValidator(ctx, val, claim.GetChainReferenceId())
			if err != nil {
				panic(err)
			}
			if claim.GetSkywayNonce() > last {
				err = k.SetLastSkywayNonceByValidator(ctx, val, claim.GetChainReferenceId(), claim.GetSkywayNonce())
				if err != nil {
					panic(err)
				}
			}
		}
	}

	// populate state with cosmos originated denom-erc20 mapping
	for i, item := range data.Erc20ToDenoms {
		ethAddr, err := types.NewEthAddress(item.Erc20)
		if err != nil {
			panic(fmt.Errorf("invalid erc20 address in Erc20ToDenoms for item %d: %s", i, item.Erc20))
		}
		err = k.setDenomToERC20(ctx, item.ChainReferenceId, item.Denom, *ethAddr)
		if err != nil {
			panic(err)
		}
	}

	for _, tax := range data.BridgeTaxes {
		if err := k.SetBridgeTax(ctx, tax); err != nil {
			panic(err)
		}
	}

	for _, limit := range data.BridgeTransferLimits {
		if err := k.SetBridgeTransferLimit(ctx, limit); err != nil {
			panic(err)
		}
	}

	if len(data.LightNodeSaleContracts) > 0 {
		err := k.SetAllLighNodeSaleContracts(ctx, data.LightNodeSaleContracts)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports all the state needed to restart the chain
// from the current state of the chain
func ExportGenesis(ctx context.Context, k Keeper) types.GenesisState {
	unbatchedTransfers, err := k.GetUnbatchedTransactions(ctx)
	if err != nil {
		panic(err)
	}

	batches, err := k.GetOutgoingTxBatches(ctx)
	if err != nil {
		panic(err)
	}

	denoms, err := k.GetAllERC20ToDenoms(ctx)
	if err != nil {
		panic(err)
	}

	var (
		p              = k.GetParams(ctx)
		batchconfs     = []types.MsgConfirmBatch{}
		batchestimates = []types.MsgEstimateBatchGas{}
		attestations   = []types.Attestation{}
		erc20ToDenoms  = []types.ERC20ToDenom{}
		nonces         = []types.SkywayNonces{}
	)

	lastSlashedBlock, err := k.GetLastSlashedBatchBlock(ctx)
	if err != nil {
		panic(err)
	}

	lastTxPoolId, err := k.getID(ctx, types.KeyLastTXPoolID)
	if err != nil {
		panic(err)
	}

	lastBatchId, err := k.getID(ctx, types.KeyLastOutgoingBatchID)
	if err != nil {
		panic(err)
	}

	for _, chain := range k.GetChainsWithTokens(ctx) {
		lastObservedNonce, err := k.GetLastObservedSkywayNonce(ctx, chain)
		if err != nil {
			panic(err)
		}
		nonces = append(nonces, types.SkywayNonces{
			LastObservedNonce:     lastObservedNonce,
			LastSlashedBatchBlock: lastSlashedBlock,
			LastTxPoolId:          lastTxPoolId,
			LastBatchId:           lastBatchId,
			ChainReferenceId:      chain,
		})

		attmap, attKeys, err := k.GetAttestationMapping(ctx, chain)
		if err != nil {
			panic(err)
		}
		// export attestations from state
		for _, key := range attKeys {
			// TODO: set height = 0?
			attestations = append(attestations, attmap[key]...)
		}
	}

	// export batch confirmations from state
	extBatches := make([]types.OutgoingTxBatch, len(batches))
	for i, batch := range batches {
		// TODO: set height = 0?
		batchConfirms, err := k.GetBatchConfirmByNonceAndTokenContract(ctx, batch.BatchNonce, batch.TokenContract)
		if err != nil {
			panic(err)
		}
		estimates, err := k.GetBatchGasEstimateByNonceAndTokenContract(ctx, batch.BatchNonce, batch.TokenContract)
		if err != nil {
			panic(err)
		}
		batchconfs = append(batchconfs, batchConfirms...)
		batchestimates = append(batchestimates, estimates...)

		extBatches[i] = batch.ToExternal()
	}

	// export erc20 to denom relations
	for _, erc20ToDenom := range denoms {
		erc20ToDenoms = append(erc20ToDenoms, *erc20ToDenom)
	}

	unbatchedTxs := make([]types.OutgoingTransferTx, len(unbatchedTransfers))
	for i, v := range unbatchedTransfers {
		unbatchedTxs[i] = v.ToExternal()
	}

	taxes, err := k.AllBridgeTaxes(ctx)
	if err != nil && !errors.Is(err, keeperutil.ErrNotFound) {
		panic(err)
	}

	limits, err := k.AllBridgeTransferLimits(ctx)
	if err != nil && !errors.Is(err, keeperutil.ErrNotFound) {
		panic(err)
	}

	contracts, err := k.AllLightNodeSaleContracts(ctx)
	if err != nil && !errors.Is(err, keeperutil.ErrNotFound) {
		panic(err)
	}

	return types.GenesisState{
		Params:                 &p,
		SkywayNonces:           nonces,
		Batches:                extBatches,
		BatchConfirms:          batchconfs,
		BatchGasEstimates:      batchestimates,
		Attestations:           attestations,
		Erc20ToDenoms:          erc20ToDenoms,
		UnbatchedTransfers:     unbatchedTxs,
		BridgeTaxes:            taxes,
		BridgeTransferLimits:   limits,
		LightNodeSaleContracts: contracts,
	}
}
