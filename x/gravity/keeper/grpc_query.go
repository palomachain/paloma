package keeper

import (
	"context"
	"strings"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	utilkeeper "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
)

// nolint: exhaustruct
var _ types.QueryServer = Keeper{}

const (
	MERCURY_UPGRADE_HEIGHT   uint64 = 1282013
	QUERY_ATTESTATIONS_LIMIT uint64 = 1000
)

// Params queries the params of the gravity module
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params := k.GetParams(c)
	return &types.QueryParamsResponse{Params: params}, nil
}

// LastPendingBatchRequestByAddr queries the LastPendingBatchRequestByAddr of
// the gravity module.
func (k Keeper) LastPendingBatchRequestByAddr(
	c context.Context,
	req *types.QueryLastPendingBatchRequestByAddrRequest,
) (*types.QueryLastPendingBatchRequestByAddrResponse, error) {
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, "address invalid")
	}

	var pendingBatchReq types.InternalOutgoingTxBatches

	found := false
	err = k.IterateOutgoingTxBatches(sdk.UnwrapSDKContext(c), func(_ []byte, batch types.InternalOutgoingTxBatch) bool {
		batchConfirm, err := k.GetBatchConfirm(sdk.UnwrapSDKContext(c), batch.BatchNonce, batch.TokenContract, addr)
		if err != nil {
			return false
		}
		foundConfirm := batchConfirm != nil
		if !foundConfirm {
			pendingBatchReq = append(pendingBatchReq, batch)
			found = true

			return true
		}

		return false
	})
	if err != nil {
		return nil, err
	}
	if found {
		ref := pendingBatchReq.ToExternalArray()
		return &types.QueryLastPendingBatchRequestByAddrResponse{Batch: ref}, nil
	}

	return &types.QueryLastPendingBatchRequestByAddrResponse{Batch: nil}, nil
}

const MaxResults = 100

// OutgoingTxBatches queries the OutgoingTxBatches of the gravity module
func (k Keeper) OutgoingTxBatches(
	c context.Context,
	req *types.QueryOutgoingTxBatchesRequest,
) (*types.QueryOutgoingTxBatchesResponse, error) {
	var batches []types.OutgoingTxBatch
	err := k.IterateOutgoingTxBatches(sdk.UnwrapSDKContext(c), func(_ []byte, batch types.InternalOutgoingTxBatch) bool {
		batchChainReferenceID := batch.ChainReferenceID
		reqChainReferenceID := req.ChainReferenceId
		batchAssignee := batch.Assignee
		reqAssignee := req.Assignee
		if reqChainReferenceID != "" && batchChainReferenceID != reqChainReferenceID {
			return false
		}
		if reqAssignee != "" && batchAssignee != reqAssignee {
			return false
		}

		batches = append(batches, batch.ToExternal())
		return len(batches) == MaxResults
	})
	if err != nil {
		return nil, err
	}
	return &types.QueryOutgoingTxBatchesResponse{Batches: batches}, nil
}

// BatchRequestByNonce queries the BatchRequestByNonce of the gravity module.
func (k Keeper) BatchRequestByNonce(
	c context.Context,
	req *types.QueryBatchRequestByNonceRequest,
) (*types.QueryBatchRequestByNonceResponse, error) {
	addr, err := types.NewEthAddress(req.ContractAddress)
	if err != nil {
		return nil, errors.Wrap(sdkerrors.ErrUnknownRequest, err.Error())
	}

	foundBatch, err := k.GetOutgoingTXBatch(sdk.UnwrapSDKContext(c), *addr, req.Nonce)
	if err != nil {
		return nil, err
	}
	if foundBatch == nil {
		return nil, errors.Wrap(sdkerrors.ErrUnknownRequest, "cannot find tx batch")
	}

	return &types.QueryBatchRequestByNonceResponse{Batch: foundBatch.ToExternal()}, nil
}

// BatchConfirms returns the batch confirmations by nonce and token contract
func (k Keeper) BatchConfirms(
	c context.Context,
	req *types.QueryBatchConfirmsRequest,
) (*types.QueryBatchConfirmsResponse, error) {
	var confirms []types.MsgConfirmBatch
	contract, err := types.NewEthAddress(req.ContractAddress)
	if err != nil {
		return nil, errors.Wrap(err, "invalid contract address in request")
	}
	err = k.IterateBatchConfirmByNonceAndTokenContract(sdk.UnwrapSDKContext(c),
		req.Nonce, *contract, func(_ []byte, c types.MsgConfirmBatch) bool {
			confirms = append(confirms, c)
			return false
		})
	if err != nil {
		return nil, err
	}
	return &types.QueryBatchConfirmsResponse{Confirms: confirms}, nil
}

// LastEventNonce returns the last event nonce observed by Paloma.
func (k Keeper) LastEventNonce(
	c context.Context,
	req *types.QueryLastEventNonceRequest,
) (*types.QueryLastEventNonceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var ret types.QueryLastEventNonceResponse
	lastEventNonce, err := k.GetLastObservedEventNonce(ctx)
	if err != nil {
		return nil, err
	}
	ret.EventNonce = lastEventNonce
	return &ret, nil
}

// LastEventNonceByAddr returns the last event nonce for the given validator address,
// this allows eth oracles to figure out where they left off
func (k Keeper) LastEventNonceByAddr(
	c context.Context,
	req *types.QueryLastEventNonceByAddrRequest,
) (*types.QueryLastEventNonceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var ret types.QueryLastEventNonceResponse
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errors.Wrap(sdkerrors.ErrInvalidAddress, req.Address)
	}
	validator, found, err := k.GetOrchestratorValidator(ctx, addr)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.Wrap(types.ErrUnknown, "address")
	}
	valAddress, err := utilkeeper.ValAddressFromBech32(k.AddressCodec, validator.GetOperator())
	if err := sdk.VerifyAddressFormat(valAddress); err != nil {
		return nil, errors.Wrap(err, "invalid validator address")
	}
	lastEventNonce, err := k.GetLastEventNonceByValidator(ctx, valAddress)
	if err != nil {
		return nil, err
	}
	ret.EventNonce = lastEventNonce
	return &ret, nil
}

// DenomToERC20 queries the Paloma Denom that maps to an Ethereum ERC20
func (k Keeper) DenomToERC20(
	c context.Context,
	req *types.QueryDenomToERC20Request,
) (*types.QueryDenomToERC20Response, error) {
	ctx := sdk.UnwrapSDKContext(c)
	erc20, err := k.GetERC20OfDenom(ctx, req.GetChainReferenceId(), req.GetDenom())
	if err != nil {
		return nil, errors.Wrapf(err, "invalid denom (%v) queried", req.Denom)
	}
	var ret types.QueryDenomToERC20Response
	ret.Erc20 = erc20.GetAddress().Hex()

	return &ret, err
}

// ERC20ToDenom queries the ERC20 contract that maps to an Ethereum ERC20 if any
func (k Keeper) ERC20ToDenom(
	c context.Context,
	req *types.QueryERC20ToDenomRequest,
) (*types.QueryERC20ToDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	ethAddr, err := types.NewEthAddress(req.Erc20)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid Erc20 in request: %s", req.Erc20)
	}
	name, err := k.GetDenomOfERC20(ctx, req.GetChainReferenceId(), *ethAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid erc20 (%v) queried", req.Erc20)
	}
	var ret types.QueryERC20ToDenomResponse
	ret.Denom = name

	return &ret, nil
}

// GetLastObservedEthBlock queries the LastObservedEthereumBlockHeight
func (k Keeper) GetLastObservedEthBlock(
	c context.Context,
	req *types.QueryLastObservedEthBlockRequest,
) (*types.QueryLastObservedEthBlockResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	locator := k.GetLastObservedEthereumBlockHeight

	ethHeight := locator(ctx)

	return &types.QueryLastObservedEthBlockResponse{Block: ethHeight.EthereumBlockHeight}, nil
}

// GetLastObservedEthNonce queries the LastObservedEventNonce
func (k Keeper) GetLastObservedEthNonce(
	c context.Context,
	req *types.QueryLastObservedEthNonceRequest,
) (*types.QueryLastObservedEthNonceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	locator := k.GetLastObservedEventNonce

	nonce, err := locator(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryLastObservedEthNonceResponse{Nonce: nonce}, nil
}

// GetAttestations queries the attestation map
func (k Keeper) GetAttestations(
	c context.Context,
	req *types.QueryAttestationsRequest,
) (*types.QueryAttestationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	iterator := k.IterateAttestations

	limit := req.Limit
	if limit == 0 || limit > QUERY_ATTESTATIONS_LIMIT {
		limit = QUERY_ATTESTATIONS_LIMIT
	}

	var (
		attestations []types.Attestation
		count        uint64
		iterErr      error
	)

	reverse := strings.EqualFold(req.OrderBy, "desc")
	filter := req.Height > 0 || req.Nonce > 0 || req.ClaimType != ""

	err := iterator(ctx, reverse, func(_ []byte, att types.Attestation) (abort bool) {
		claim, err := k.UnpackAttestationClaim(&att)
		if err != nil {
			iterErr = errors.Wrap(sdkerrors.ErrUnpackAny, "failed to unmarshal Ethereum claim")
			return true
		}

		var match bool
		switch {
		case filter && claim.GetEthBlockHeight() == req.Height:
			attestations = append(attestations, att)
			match = true

		case filter && claim.GetEventNonce() == req.Nonce:
			attestations = append(attestations, att)
			match = true

		case filter && claim.GetType().String() == req.ClaimType:
			attestations = append(attestations, att)
			match = true

		case !filter:
			// No filter provided, so we include the attestation. This is equivalent
			// to providing no query params or just limit and/or order_by.
			attestations = append(attestations, att)
			match = true
		}

		if match {
			count++
			if count >= limit {
				return true
			}
		}

		return false
	})
	if iterErr != nil {
		return nil, iterErr
	}
	if err != nil {
		return nil, err
	}

	return &types.QueryAttestationsResponse{Attestations: attestations}, nil
}

func (k Keeper) GetErc20ToDenoms(c context.Context, denoms *types.QueryErc20ToDenoms) (*types.QueryErc20ToDenomsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	erc20ToDenomsPtrs, err := k.GetAllERC20ToDenoms(ctx)
	if err != nil {
		return nil, err
	}

	erc20ToDenoms := make([]types.ERC20ToDenom, len(erc20ToDenomsPtrs))

	for k, v := range erc20ToDenomsPtrs {
		erc20ToDenoms[k] = *v
	}

	return &types.QueryErc20ToDenomsResponse{
		Erc20ToDenom: erc20ToDenoms,
	}, nil
}

func (k Keeper) GetPendingSendToEth(
	c context.Context,
	req *types.QueryPendingSendToEth,
) (*types.QueryPendingSendToEthResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	batches, err := k.GetOutgoingTxBatches(ctx)
	if err != nil {
		return nil, err
	}
	unbatchedTxs, err := k.GetUnbatchedTransactions(ctx)
	if err != nil {
		return nil, err
	}
	senderAddress := req.GetSenderAddress()
	res := types.QueryPendingSendToEthResponse{
		TransfersInBatches: []types.OutgoingTransferTx{},
		UnbatchedTransfers: []types.OutgoingTransferTx{},
	}
	for _, batch := range batches {
		for _, tx := range batch.Transactions {
			if senderAddress == "" || tx.Sender.String() == senderAddress {
				res.TransfersInBatches = append(res.TransfersInBatches, tx.ToExternal())
			}
		}
	}
	for _, tx := range unbatchedTxs {
		if senderAddress == "" || tx.Sender.String() == senderAddress {
			res.UnbatchedTransfers = append(res.UnbatchedTransfers, tx.ToExternal())
		}
	}

	return &res, nil
}
