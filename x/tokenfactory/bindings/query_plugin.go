package bindings

import (
	"encoding/json"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	bindingstypes "github.com/palomachain/paloma/v2/x/tokenfactory/bindings/types"
	tokenfactorykeeper "github.com/palomachain/paloma/v2/x/tokenfactory/keeper"
)

type queryPlugin struct {
	// TODO: Use interface abstract to decouple dependencies
	bankKeeper         *bankkeeper.BaseKeeper
	tokenFactoryKeeper *tokenfactorykeeper.Keeper
}

func NewQueryPlugin(b *bankkeeper.BaseKeeper, tfk *tokenfactorykeeper.Keeper) wasmkeeper.CustomQuerier {
	return customQuerier(&queryPlugin{
		bankKeeper:         b,
		tokenFactoryKeeper: tfk,
	})
}

func customQuerier(qp *queryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var q bindingstypes.Query
		if err := json.Unmarshal(request, &q); err != nil {
			return nil, sdkerrors.Wrap(err, "osmosis query")
		}
		switch {
		case q.FullDenom != nil:
			creator := q.FullDenom.CreatorAddr
			subdenom := q.FullDenom.Subdenom

			fullDenom, err := getFullDenom(creator, subdenom)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "full denom query")
			}

			res := bindingstypes.FullDenomResponse{
				Denom: fullDenom,
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "failed to marshal FullDenomResponse")
			}

			return bz, nil

		case q.Admin != nil:
			res, err := qp.getDenomAdmin(ctx, q.Admin.Denom)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON marshal AdminResponse: %w", err)
			}

			return bz, nil

		case q.Metadata != nil:
			res, err := qp.getMetadata(ctx, q.Metadata.Denom)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON marshal MetadataResponse: %w", err)
			}

			return bz, nil

		case q.DenomsByCreator != nil:
			res, err := qp.getDenomsByCreator(ctx, q.DenomsByCreator.Creator)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON marshal DenomsByCreatorResponse: %w", err)
			}

			return bz, nil

		case q.Params != nil:
			res, err := qp.getParams(ctx)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON marshal ParamsResponse: %w", err)
			}

			return bz, nil

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown token query variant"}
		}
	}
}

func (qp queryPlugin) getDenomAdmin(ctx sdk.Context, denom string) (*bindingstypes.AdminResponse, error) {
	metadata, err := qp.tokenFactoryKeeper.GetAuthorityMetadata(ctx, denom)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin for denom: %s", denom)
	}
	return &bindingstypes.AdminResponse{Admin: metadata.Admin}, nil
}

func (qp queryPlugin) getDenomsByCreator(ctx sdk.Context, creator string) (*bindingstypes.DenomsByCreatorResponse, error) {
	if _, err := parseAddress(creator); err != nil {
		return nil, fmt.Errorf("invalid creator address: %s", creator)
	}

	denoms := qp.tokenFactoryKeeper.GetDenomsFromCreator(ctx, creator)
	return &bindingstypes.DenomsByCreatorResponse{Denoms: denoms}, nil
}

func (qp queryPlugin) getMetadata(ctx sdk.Context, denom string) (*bindingstypes.MetadataResponse, error) {
	metadata, found := qp.bankKeeper.GetDenomMetaData(ctx, denom)
	var parsed *bindingstypes.Metadata
	if found {
		parsed = sdkMetadataToWasm(metadata)
	}
	return &bindingstypes.MetadataResponse{Metadata: parsed}, nil
}

func (qp queryPlugin) getParams(ctx sdk.Context) (*bindingstypes.ParamsResponse, error) {
	params := qp.tokenFactoryKeeper.GetParams(ctx)
	return &bindingstypes.ParamsResponse{
		Params: bindingstypes.Params{
			DenomCreationFee: convertSdkCoinsToWasmCoins(params.DenomCreationFee),
		},
	}, nil
}

func convertSdkCoinsToWasmCoins(coins []sdk.Coin) []wasmvmtypes.Coin {
	var toSend []wasmvmtypes.Coin
	for _, coin := range coins {
		c := convertSdkCoinToWasmCoin(coin)
		toSend = append(toSend, c)
	}
	return toSend
}

func convertSdkCoinToWasmCoin(coin sdk.Coin) wasmvmtypes.Coin {
	return wasmvmtypes.Coin{
		Denom: coin.Denom,
		// Note: tokens have 18 decimal places, so 10^22 is common, no longer in u64 range
		Amount: coin.Amount.String(),
	}
}
