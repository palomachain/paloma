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
	cdc             codec.BinaryCodec
	memKey          storetypes.StoreKey
	paramstore      paramtypes.Subspace
	bank            types.BankKeeper
	account         types.AccountKeeper
	Scheduler       types.Scheduler
	Chains          []xchain.FundCollecter
	treasureStore   *keeperutil.KVStoreWrapper[*types.Fees]
	relayerFeeStore *keeperutil.KVStoreWrapper[*types.RelayerFee]
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
		cdc:             cdc,
		memKey:          memKey,
		paramstore:      ps,
		bank:            bank,
		account:         account,
		Scheduler:       scheduler,
		treasureStore:   keeperutil.NewKvStoreWrapper[*types.Fees](keeperutil.StoreFactory(storeKey, types.TreasureStorePrefix), cdc),
		relayerFeeStore: keeperutil.NewKvStoreWrapper[*types.RelayerFee](keeperutil.StoreFactory(storeKey, types.RelayerFeeStorePrefix), cdc),
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

func (k Keeper) SetRelayerFee(ctx sdk.Context, fee *types.RelayerFee) error {
	if fee == nil {
		return fmt.Errorf("nil fee not allowed")
	}

	return k.KeeperUtil.Save(k.Store.RelayerFeeStore(ctx), k.cdc)
}
