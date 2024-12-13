package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/palomachain/paloma/v2/x/tokenfactory/types"
)

func (k Keeper) CreateDenom(ctx context.Context, creatorAddr string, subdenom string) (newTokenDenom string, err error) {
	denom, err := k.validateCreateDenom(ctx, creatorAddr, subdenom)
	if err != nil {
		return "", err
	}

	err = k.chargeForCreateDenom(ctx, creatorAddr, subdenom)
	if err != nil {
		return "", err
	}

	err = k.createDenomAfterValidation(ctx, creatorAddr, denom)
	return denom, err
}

func (k Keeper) createDenomAfterValidation(ctx context.Context, creatorAddr string, denom string) (err error) {
	denomMetaData := banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    denom,
			Exponent: 0,
		}},
		Base: denom,
	}

	k.bankKeeper.SetDenomMetaData(ctx, denomMetaData)

	authorityMetadata := types.DenomAuthorityMetadata{
		Admin: creatorAddr,
	}
	err = k.setAuthorityMetadata(ctx, denom, authorityMetadata)
	if err != nil {
		return err
	}

	k.addDenomFromCreator(ctx, creatorAddr, denom)
	return nil
}

func (k Keeper) validateCreateDenom(ctx context.Context, creatorAddr string, subdenom string) (newTokenDenom string, err error) {
	if k.bankKeeper.HasSupply(ctx, subdenom) {
		return "", fmt.Errorf("temporary error until IBC bug is sorted out, " +
			"can't create subdenoms that are the same as a native denom")
	}

	denom, err := types.GetTokenDenom(creatorAddr, subdenom)
	if err != nil {
		return "", err
	}

	_, found := k.bankKeeper.GetDenomMetaData(ctx, denom)
	if found {
		return "", types.ErrDenomExists
	}

	return denom, nil
}

func (k Keeper) chargeForCreateDenom(ctx context.Context, creatorAddr string, subdenom string) (err error) {
	creationFee := k.GetParams(ctx).DenomCreationFee
	accAddr, err := sdk.AccAddressFromBech32(creatorAddr)
	if err != nil {
		return err
	}
	if creationFee != nil {
		if err := k.communityPoolKeeper.FundCommunityPool(ctx, creationFee, accAddr); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) addDenomFromCreator(ctx context.Context, creator, denom string) {
	store := k.GetCreatorPrefixStore(ctx, creator)
	store.Set([]byte(denom), []byte(denom))
}

func (k Keeper) GetDenomsFromCreator(ctx context.Context, creator string) []string {
	store := k.GetCreatorPrefixStore(ctx, creator)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	denoms := []string{}
	for ; iterator.Valid(); iterator.Next() {
		denoms = append(denoms, string(iterator.Key()))
	}
	return denoms
}

func (k Keeper) GetAllDenomsIterator(ctx context.Context) storetypes.Iterator {
	return k.GetCreatorsPrefixStore(ctx).Iterator(nil, nil)
}
