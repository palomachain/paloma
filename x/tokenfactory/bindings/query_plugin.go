package bindings

import (
	"encoding/json"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errtypes "github.com/cosmos/cosmos-sdk/types/errors"

	bindingstypes "github.com/palomachain/paloma/v2/x/tokenfactory/bindings/types"
)

func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindingstypes.TokenFactoryQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "osmosis query")
		}
		if contractQuery.Token == nil {
			return nil, sdkerrors.Wrap(errtypes.ErrUnknownRequest, "nil token field")
		}
		tokenQuery := contractQuery.Token

		switch {
		case tokenQuery.FullDenom != nil:
			creator := tokenQuery.FullDenom.CreatorAddr
			subdenom := tokenQuery.FullDenom.Subdenom

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

		case tokenQuery.Admin != nil:
			res, err := qp.GetDenomAdmin(ctx, tokenQuery.Admin.Denom)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON marshal AdminResponse: %w", err)
			}

			return bz, nil

		case tokenQuery.Metadata != nil:
			res, err := qp.GetMetadata(ctx, tokenQuery.Metadata.Denom)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON marshal MetadataResponse: %w", err)
			}

			return bz, nil

		case tokenQuery.DenomsByCreator != nil:
			res, err := qp.GetDenomsByCreator(ctx, tokenQuery.DenomsByCreator.Creator)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON marshal DenomsByCreatorResponse: %w", err)
			}

			return bz, nil

		case tokenQuery.Params != nil:
			res, err := qp.GetParams(ctx)
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

func ConvertSdkCoinsToWasmCoins(coins []sdk.Coin) []wasmvmtypes.Coin {
	var toSend []wasmvmtypes.Coin
	for _, coin := range coins {
		c := ConvertSdkCoinToWasmCoin(coin)
		toSend = append(toSend, c)
	}
	return toSend
}

func ConvertSdkCoinToWasmCoin(coin sdk.Coin) wasmvmtypes.Coin {
	return wasmvmtypes.Coin{
		Denom: coin.Denom,
		// Note: tokens have 18 decimal places, so 10^22 is common, no longer in u64 range
		Amount: coin.Amount.String(),
	}
}
