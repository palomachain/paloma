package keeper

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/treasury/types"
)

const storeKey = "treasury"

type Keeper struct {
	cdc        codec.BinaryCodec
	memKey     storetypes.StoreKey
	paramstore paramtypes.Subspace
	bank       types.BankKeeper
	account    types.AccountKeeper
	Scheduler  types.Scheduler
	Chains     []xchain.FundCollecter
	KeeperUtil keeperutil.KeeperUtilI[*types.Fees]
	Store      types.TreasuryStore
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bank types.BankKeeper,
	account types.AccountKeeper,
	scheduler types.Scheduler,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		memKey:     memKey,
		paramstore: ps,
		bank:       bank,
		account:    account,
		Scheduler:  scheduler,
		KeeperUtil: keeperutil.KeeperUtil[*types.Fees]{},
		Store: types.Store{
			StoreKey: storeKey,
		},
	}
}

func (k Keeper) ModuleName() string { return types.ModuleName }

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetCommunityFundFee(ctx sdk.Context, fee string) error {
	fees, err := k.GetFees(ctx)
	if err != nil {
		return err
	}

	fees.CommunityFundFee = fee

	err = k.setFees(ctx, fees)

	return err
}

func (k Keeper) SetSecurityFee(ctx sdk.Context, fee string) error {
	fees, err := k.GetFees(ctx)
	if err != nil {
		return err
	}

	fees.SecurityFee = fee

	err = k.setFees(ctx, fees)

	return err
}

func (k Keeper) GetFees(ctx sdk.Context) (*types.Fees, error) {
	res, err := k.KeeperUtil.Load(k.Store.TreasuryStore(ctx), k.cdc, []byte(storeKey))
	if errors.Is(err, keeperutil.ErrNotFound) {
		return &types.Fees{}, nil
	}
	if err != nil {
		return &types.Fees{}, err
	}
	return res, nil
}

func (k Keeper) setFees(ctx sdk.Context, fees *types.Fees) error {
	return k.KeeperUtil.Save(k.Store.TreasuryStore(ctx), k.cdc, []byte(storeKey), fees)
}

// func (k Keeper) PayPigeonsForTheirService(
// 	ctx sdk.Context,
// ) error {
// 	return nil
// }

// func (k Keeper) PayForRunningAJob(
// 	ctx sdk.Context,
// 	jobAddr,
// 	to sdk.AccAddress,
// 	chainType xchain.Type,
// 	chainRefID xchain.ReferenceID,
// 	amount sdk.Int,
// ) error {
// 	return nil
// }

// func (k Keeper) KeepingTheNetworkAlive(
// 	ctx sdk.Context,
// 	valAddr sdk.ValAddress,
// 	chainType xchain.Type,
// 	chainRefID xchain.ReferenceID,
// 	amount sdk.Int,
// 	reason string, //e.g. valset update
// ) error {
// 	return nil
// }

func (k Keeper) AddFunds(
	ctx sdk.Context,
	chainType xchain.Type,
	chainRefID xchain.ReferenceID,
	jobID string,
	amount sdk.Int,
) error {
	job, err := k.Scheduler.GetJob(ctx, jobID)
	if err != nil {
		return err
	}
	addr := job.GetAddress()

	denom := fmt.Sprintf("%s/%s", chainType, chainRefID)
	if amount.IsZero() {
		return types.ErrCannotAddZeroFunds.Wrapf("chain type: %s, chainRefID: %s, addr: %s", chainType, chainRefID, addr)
	}
	if amount.IsNegative() {
		return types.ErrCannotAddNegativeFunds.Wrapf("chain type: %s, chainRefID: %s, addr: %s", chainType, chainRefID, addr)
	}

	ctx, writeCtx := ctx.CacheContext()

	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))

	err = k.bank.MintCoins(ctx, k.ModuleName(), coins)
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

// func (k Keeper) TriggerFundEvents(ctx sdk.Context) {
// 	var g whoops.Group
// 	for _, c := range k.Chains {
// 		g.Add(c.CollectJobFundEvents(ctx))
// 	}
// 	if g.Err() {
// 		k.Logger(ctx).With("err", g).Error("error triggering fund events")
// 	}
// }
