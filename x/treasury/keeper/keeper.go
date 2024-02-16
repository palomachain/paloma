package keeper

import (
	"context"
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

const treasuryFeeStoreKey keeperutil.Key = "treasury"

type Keeper struct {
	paramstore      paramtypes.Subspace
	cdc             codec.BinaryCodec
	memKey          storetypes.StoreKey
	bank            types.BankKeeper
	account         types.AccountKeeper
	staking         types.StakingKeeper
	Scheduler       types.Scheduler
	treasureStore   keeperutil.KVStoreWrapper[*types.Fees]
	relayerFeeStore keeperutil.KVStoreWrapper[*types.RelayerFee]
	Chains          []xchain.FundCollecter
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bank types.BankKeeper,
	account types.AccountKeeper,
	staking types.StakingKeeper,
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
		staking:         staking,
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
		return fmt.Errorf("SetCommunityFundFee: %w", err)
	}

	fees.CommunityFundFee = fee
	return k.setFees(ctx, fees)
}

func (k Keeper) SetSecurityFee(ctx sdk.Context, fee string) error {
	fees, err := k.GetFees(ctx)
	if err != nil {
		return fmt.Errorf("SetSecurityFee: %w", err)
	}

	fees.SecurityFee = fee
	return k.setFees(ctx, fees)
}

func (k Keeper) GetFees(ctx sdk.Context) (*types.Fees, error) {
	res, err := k.treasureStore.Get(ctx, treasuryFeeStoreKey)
	if errors.Is(err, keeperutil.ErrNotFound) {
		return &types.Fees{}, nil
	}
	if err != nil {
		return &types.Fees{}, err
	}
	return res, nil
}

func (k Keeper) setFees(ctx sdk.Context, fees *types.Fees) error {
	return k.treasureStore.Set(ctx, treasuryFeeStoreKey, fees)
}

func (k Keeper) setRelayerFee(ctx context.Context, valAddress sdk.ValAddress, fee *types.RelayerFee) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	val := k.staking.Validator(sdkCtx, valAddress)
	if val == nil {
		return fmt.Errorf("validator not found")
	}
	if fee == nil {
		return fmt.Errorf("nil fee not allowed")
	}

	return k.relayerFeeStore.Set(ctx, valAddress, fee)
}

func (k Keeper) getRelayerFee(ctx context.Context, valAddress sdk.ValAddress) (*types.RelayerFee, error) {
	fee, err := k.relayerFeeStore.Get(ctx, valAddress)
	if err != nil {
		if errors.Is(err, keeperutil.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return fee, nil
}

func (k Keeper) getRelayerFees(ctx context.Context) ([]types.RelayerFee, error) {
	fees := make([]types.RelayerFee, 0, 200)
	if err := k.relayerFeeStore.Iterate(ctx, func(_ []byte, f *types.RelayerFee) bool {
		fees = append(fees, *f)
		return true
	}); err != nil {
		return nil, fmt.Errorf("iterate relayer fee store: %w", err)
	}

	return fees, nil
}
