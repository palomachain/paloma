package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	"github.com/palomachain/paloma/x/treasury/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		paramstore paramtypes.Subspace
		bank       types.BankKeeper
		account    types.AccountKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{

		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
	}
}

func (k Keeper) ModuleName() string { return types.ModuleName }

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) PayPigeonsForTheirService(
	ctx sdk.Context,
) error {
	return nil
}

func (k Keeper) PayForRunningAJob(
	ctx sdk.Context,
	jobAddr,
	to sdk.AccAddress,
	chainType xchain.Type,
	chainRefID xchain.ReferenceID,
	amount sdk.Int,
) error {
	return nil
}

func (k Keeper) AddFunds(
	ctx sdk.Context,
	chainType xchain.Type,
	chainRefID xchain.ReferenceID,
	addr sdk.AccAddress,
	amount sdk.Int,
) error {
	denom := fmt.Sprintf("%s/%s", chainType, chainRefID)
	if amount.IsZero() {
		return "cannot add zero amount"
	}

	oldCtx := ctx
	ctx, writeCtx := ctx.CacheContext()

	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))
	err := k.bank.MintCoins(ctx, k.ModuleName(), coins)
	if err != nil {
		return err
	}

	err = k.bank.SendCoinsFromModuleToAccount(ctx, k.ModuleName(), addr, coins)
	if err != nil {
		return err
	}

	writeCtx()
	return nil
}
