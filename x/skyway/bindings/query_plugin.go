package bindings

import (
	"context"
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bindingstypes "github.com/palomachain/paloma/v2/x/skyway/bindings/types"
	skywaytypes "github.com/palomachain/paloma/v2/x/skyway/types"
)

type SkywayKeeper interface {
	GetAllERC20ToDenoms(context.Context) ([]*skywaytypes.ERC20ToDenom, error)
}

type queryPlugin struct {
	k SkywayKeeper
}

func NewQueryPlugin(k SkywayKeeper) wasmkeeper.CustomQuerier {
	return customQuerier(&queryPlugin{
		k: k,
	})
}

func customQuerier(qp *queryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var query bindingstypes.Query
		if err := json.Unmarshal(request, &query); err != nil {
			return nil, sdkerrors.Wrap(err, "query")
		}
		switch {
		case query.Erc20ToDenoms != nil:
			return queryErc20ToDenoms(ctx, qp.k)
		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown scheduler query variant"}
		}
	}
}

func queryErc20ToDenoms(ctx sdk.Context, k SkywayKeeper) ([]byte, error) {
	denoms, err := k.GetAllERC20ToDenoms(ctx)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to query ERC20 to denoms")
	}

	castDenoms := make([]bindingstypes.ERC20ToDenom, len(denoms))
	for i, d := range denoms {
		castDenoms[i] = bindingstypes.ERC20ToDenom{
			Erc20:            d.Erc20,
			Denom:            d.Denom,
			ChainReferenceId: d.ChainReferenceId,
		}
	}

	res := bindingstypes.Erc20ToDenomsResponse{
		Denoms: castDenoms,
	}

	bz, err := json.Marshal(res)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to marshal response")
	}

	return bz, nil
}
