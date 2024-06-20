package keeper

import (
	"context"
	"errors"
	"fmt"

	cosmosstore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/treasury/types"
)

const storeKey = "treasury"

type Keeper struct {
	cdc         codec.BinaryCodec
	paramstore  paramtypes.Subspace
	bank        types.BankKeeper
	account     types.AccountKeeper
	Scheduler   types.Scheduler
	evm         types.EvmKeeper
	Chains      []xchain.FundCollecter
	KeeperUtil  keeperutil.KeeperUtilI[*types.Fees]
	Store       types.TreasuryStore
	relayerFees keeperutil.KVStoreWrapper[*types.RelayerFeeSetting]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey cosmosstore.KVStoreService, ps paramtypes.Subspace,
	bank types.BankKeeper,
	account types.AccountKeeper,
	scheduler types.Scheduler,
	evm types.EvmKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	k := &Keeper{
		cdc:        cdc,
		paramstore: ps,
		bank:       bank,
		account:    account,
		Scheduler:  scheduler,
		evm:        evm,
		KeeperUtil: keeperutil.KeeperUtil[*types.Fees]{},
		Store: types.Store{
			StoreKey: storeKey,
		},
	}

	k.relayerFees = keeperutil.NewKvStoreWrapper[*types.RelayerFeeSetting](keeperutil.StoreFactory(storeKey, types.RelayerFeeStorePrefix), cdc)

	return k
}

func (k Keeper) ModuleName() string { return types.ModuleName }

func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetCommunityFundFee(ctx context.Context, fee string) error {
	fees, err := k.GetFees(ctx)
	if err != nil {
		return err
	}

	fees.CommunityFundFee = fee

	err = k.setFees(ctx, fees)

	return err
}

func (k Keeper) SetSecurityFee(ctx context.Context, fee string) error {
	fees, err := k.GetFees(ctx)
	if err != nil {
		return err
	}

	fees.SecurityFee = fee

	err = k.setFees(ctx, fees)

	return err
}

func (k Keeper) GetFees(ctx context.Context) (*types.Fees, error) {
	res, err := k.KeeperUtil.Load(k.Store.TreasuryStore(ctx), k.cdc, []byte(storeKey))
	if errors.Is(err, keeperutil.ErrNotFound) {
		return &types.Fees{}, nil
	}
	if err != nil {
		return &types.Fees{}, err
	}
	return res, nil
}

func (k Keeper) GetRelayerFeesByChainReferenceID(ctx context.Context, chainReferenceID string) (map[string]types.RelayerFeeSetting_FeeSetting, error) {
	r := make(map[string]types.RelayerFeeSetting_FeeSetting)
	err := k.relayerFees.Iterate(sdk.UnwrapSDKContext(ctx), func(b []byte, rfs *types.RelayerFeeSetting) bool {
		for _, v := range rfs.Fees {
			if v.ChainReferenceId == chainReferenceID {
				r[rfs.ValAddress] = v
				break
			}
		}
		return true
	})

	return r, err
}

func (k Keeper) setFees(ctx context.Context, fees *types.Fees) error {
	return k.KeeperUtil.Save(k.Store.TreasuryStore(ctx), k.cdc, []byte(storeKey), fees)
}
