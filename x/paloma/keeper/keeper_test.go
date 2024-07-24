package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/palomachain/paloma/app/params"
	"github.com/palomachain/paloma/testutil/common"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/paloma/keeper"
	"github.com/palomachain/paloma/x/paloma/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	clientAddr  = "paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk"
	creatorAddr = "paloma1l2j8vaykh03zenzytntj3cza6zfxwlj68dd0l3"
)

func TestLightNodeClientFeegranter(t *testing.T) {
	common.SetupPalomaPrefixes()

	k, _, ctx := newMockedKeeper(t)

	t.Run("Should return not found on empty store", func(t *testing.T) {
		_, err := k.LightNodeClientFeegranter(ctx)
		require.ErrorIs(t, err, keeperutil.ErrNotFound)
	})

	t.Run("Should save a new feegranter", func(t *testing.T) {
		accAddr, err := sdk.AccAddressFromBech32(clientAddr)
		require.NoError(t, err)

		err = k.SetLightNodeClientFeegranter(ctx, accAddr)
		require.NoError(t, err)

		res, err := k.LightNodeClientFeegranter(ctx)
		require.NoError(t, err)

		require.Equal(t, accAddr, res.Account)
	})
}

func TestLightNodeClientLicensesCRUD(t *testing.T) {
	k, _, ctx := newMockedKeeper(t)

	t.Run("Should return not found on empty store", func(t *testing.T) {
		_, err := k.GetLightNodeClientLicense(ctx, clientAddr)
		require.ErrorIs(t, err, keeperutil.ErrNotFound)
	})

	t.Run("Should return empty slice on empty store", func(t *testing.T) {
		res, err := k.AllLightNodeClientLicenses(ctx)
		require.NoError(t, err)
		require.Empty(t, res)
	})

	t.Run("Should save a new license", func(t *testing.T) {
		license := &types.LightNodeClientLicense{
			ClientAddress: clientAddr,
			Amount:        sdk.Coin{Amount: math.NewInt(100), Denom: testBondDenom},
			VestingMonths: 12,
		}

		err := k.SetLightNodeClientLicense(ctx, clientAddr, license)
		require.NoError(t, err)

		res, err := k.GetLightNodeClientLicense(ctx, clientAddr)
		require.NoError(t, err)
		require.Equal(t, res, license)

		all, err := k.AllLightNodeClientLicenses(ctx)
		require.NoError(t, err)
		require.Len(t, all, 1)
		require.Equal(t, all[0], license)
	})
}

func TestCreateLightNodeClientLicense(t *testing.T) {
	setup := func(level int, hasAccount bool) (*keeper.Keeper, context.Context) {
		k, ms, ctx := newMockedKeeper(t)

		// We always want the same mocks, but different test errors will use
		// different mocks, so, instead of using `Maybe()`, we use level to
		// control the expectations. This way tests fail unless they call
		// exactly what should be called.
		if level > 0 {
			ms.AccountKeeper.On("AddressCodec").
				Return(authcodec.NewBech32Codec(params.AccountAddressPrefix)).
				Once()
			ms.AccountKeeper.On("HasAccount", mock.Anything, mock.Anything).
				Return(hasAccount).
				Once()
		}

		if level > 1 {
			ms.AccountKeeper.On("NewAccount", mock.Anything, mock.Anything).
				Return(&authtypes.BaseAccount{}).
				Once()
			ms.AccountKeeper.On("SetAccount", mock.Anything, mock.Anything).
				Return().
				Once()
		}

		if level > 2 {
			ms.BankKeeper.On("SendCoinsFromAccountToModule", mock.Anything,
				mock.Anything, mock.Anything, mock.Anything).
				Return(nil).
				Once()
		}

		return k, ctx
	}

	license := &types.LightNodeClientLicense{
		ClientAddress: clientAddr,
		Amount:        sdk.Coin{Amount: math.NewInt(100), Denom: testBondDenom},
		VestingMonths: 12,
	}

	t.Run("Should fail with invalid coin", func(t *testing.T) {
		k, ctx := setup(0, false)

		err := k.CreateLightNodeClientLicense(ctx, creatorAddr, clientAddr,
			sdk.Coin{}, license.VestingMonths)
		require.ErrorIs(t, err, types.ErrInvalidParameters)
	})

	t.Run("Should fail if account already exists", func(t *testing.T) {
		k, ctx := setup(1, true)

		err := k.CreateLightNodeClientLicense(ctx, creatorAddr, clientAddr,
			license.Amount, license.VestingMonths)
		require.ErrorIs(t, err, types.ErrAccountExists)
	})

	t.Run("Should fail if license already exists", func(t *testing.T) {
		k, ctx := setup(3, false)

		// Add a license
		err := k.CreateLightNodeClientLicense(ctx, creatorAddr, clientAddr,
			license.Amount, license.VestingMonths)
		require.NoError(t, err)

		err = k.CreateLightNodeClientLicense(ctx, creatorAddr, clientAddr,
			license.Amount, license.VestingMonths)
		require.ErrorIs(t, err, types.ErrLicenseExists)
	})

	t.Run("Should create a new license", func(t *testing.T) {
		k, ctx := setup(3, false)

		noGrantLicense := &types.LightNodeClientLicense{
			ClientAddress: license.ClientAddress,
			Amount:        license.Amount,
			VestingMonths: license.VestingMonths,
		}

		err := k.CreateLightNodeClientLicense(ctx, creatorAddr, clientAddr,
			noGrantLicense.Amount, noGrantLicense.VestingMonths)
		require.NoError(t, err)

		res, err := k.GetLightNodeClientLicense(ctx, clientAddr)
		require.NoError(t, err)
		require.Equal(t, res, noGrantLicense)
	})
}

func TestCreateSaleLightNodeClientLicense(t *testing.T) {
	setup := func() (*keeper.Keeper, context.Context) {
		k, ms, ctx := newMockedKeeper(t)

		ms.AccountKeeper.On("AddressCodec").
			Return(authcodec.NewBech32Codec(params.AccountAddressPrefix)).
			Twice()
		ms.AccountKeeper.On("HasAccount", mock.Anything, mock.Anything).
			Return(false).
			Once()
		ms.AccountKeeper.On("NewAccount", mock.Anything, mock.Anything).
			Return(&authtypes.BaseAccount{}).
			Once()
		ms.AccountKeeper.On("SetAccount", mock.Anything, mock.Anything).
			Return().
			Once()
		ms.BankKeeper.On("SendCoinsFromAccountToModule", mock.Anything,
			mock.Anything, mock.Anything, mock.Anything).
			Return(nil).
			Once()
		ms.FeegrantKeeper.On("GrantAllowance", mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).
			Return(nil).
			Once()

		return k, ctx
	}

	amount := math.NewInt(100)
	license := &types.LightNodeClientLicense{
		ClientAddress: clientAddr,
		Amount:        sdk.Coin{Amount: amount, Denom: testBondDenom},
		VestingMonths: 24,
	}

	t.Run("Should fail on feegrant license without feegranter", func(t *testing.T) {
		k, _, ctx := newMockedKeeper(t)

		err := k.CreateSaleLightNodeClientLicense(ctx, clientAddr, amount)
		require.ErrorIs(t, err, types.ErrNoFeegranter)
	})

	t.Run("Should fail on feegrant license without funder", func(t *testing.T) {
		k, _, ctx := newMockedKeeper(t)

		// Set a feegranter
		accAddr, err := sdk.AccAddressFromBech32(creatorAddr)
		require.NoError(t, err)

		err = k.SetLightNodeClientFeegranter(ctx, accAddr)
		require.NoError(t, err)

		err = k.CreateSaleLightNodeClientLicense(ctx, clientAddr, amount)
		require.ErrorIs(t, err, types.ErrNoFunder)
	})

	t.Run("Should create a new license with feegrant", func(t *testing.T) {
		k, ctx := setup()

		// Set a feegranter
		accAddr, err := sdk.AccAddressFromBech32(creatorAddr)
		require.NoError(t, err)

		err = k.SetLightNodeClientFeegranter(ctx, accAddr)
		require.NoError(t, err)

		// Set a funder
		err = k.SetLightNodeClientFunder(ctx, accAddr)
		require.NoError(t, err)

		err = k.CreateSaleLightNodeClientLicense(ctx, clientAddr, amount)
		require.NoError(t, err)

		res, err := k.GetLightNodeClientLicense(ctx, clientAddr)
		require.NoError(t, err)
		require.Equal(t, res, license)
	})
}

func TestCreateLightNodeClientAccount(t *testing.T) {
	setup := func(level int, getAccount bool) (*keeper.Keeper, context.Context) {
		k, ms, ctx := newMockedKeeper(t)

		if level > 0 {
			ms.AccountKeeper.On("AddressCodec").
				Return(authcodec.NewBech32Codec(params.AccountAddressPrefix)).
				Twice()
			ms.AccountKeeper.On("NewAccount", mock.Anything, mock.Anything).
				Return(&authtypes.BaseAccount{}).
				Once()
			ms.AccountKeeper.On("SetAccount", mock.Anything, mock.Anything).
				Return().
				Once()
			ms.BankKeeper.On("SendCoinsFromAccountToModule", mock.Anything,
				mock.Anything, mock.Anything, mock.Anything).
				Return(nil).
				Once()
			ms.AccountKeeper.On("HasAccount", mock.Anything, mock.Anything).
				Return(false).
				Once()

			if getAccount {
				ms.AccountKeeper.On("GetAccount", mock.Anything, mock.Anything).
					Return(&authtypes.BaseAccount{}).Once()
			} else {
				ms.AccountKeeper.On("GetAccount", mock.Anything, mock.Anything).
					Return(nil).Once()
			}
		}

		if level > 1 {
			ms.AccountKeeper.On("SetAccount", mock.Anything, mock.Anything).
				Return().
				Once()
			ms.BankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything,
				mock.Anything, mock.Anything, mock.Anything).
				Return(nil).
				Once()
		}

		return k, ctx
	}

	t.Run("Should return error on no license", func(t *testing.T) {
		k, ctx := setup(0, false)

		err := k.CreateLightNodeClientAccount(ctx, clientAddr)
		require.ErrorIs(t, err, types.ErrNoLicense)
	})

	t.Run("Should return error on no base account", func(t *testing.T) {
		k, ctx := setup(1, false)

		// Create a license
		err := k.CreateLightNodeClientLicense(ctx, creatorAddr, clientAddr,
			sdk.Coin{Amount: math.NewInt(1), Denom: testBondDenom}, 1)
		require.NoError(t, err)

		err = k.CreateLightNodeClientAccount(ctx, clientAddr)
		require.ErrorIs(t, err, types.ErrNoAccount)
	})

	t.Run("Should create the account", func(t *testing.T) {
		k, ctx := setup(2, true)

		// Create a license
		err := k.CreateLightNodeClientLicense(ctx, creatorAddr, clientAddr,
			sdk.Coin{Amount: math.NewInt(1), Denom: testBondDenom}, 1)
		require.NoError(t, err)

		err = k.CreateLightNodeClientAccount(ctx, clientAddr)
		require.NoError(t, err)
	})
}
