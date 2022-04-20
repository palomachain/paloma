package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/volumefi/cronchain/x/valset/types/mocks"
)

func TestRegisterRunner(t *testing.T) {
	val := sdk.ValAddress("validator")
	t.Run("it returns error if validator does not exist", func(t *testing.T) {
		k, ms, ctx := newValsetKeeper(t)

		// returns no validator
		ms.StakingKeeper.On("Validator", ctx, val).Return(nil).Once()

		err := k.Register(ctx, val)
		require.ErrorIs(t, err, ErrValidatorWithAddrNotFound)
	})

	t.Run("it registers the validator", func(t *testing.T) {
		k, ms, ctx := newValsetKeeper(t)
		val := sdk.ValAddress("validator")

		vali := mocks.NewStakingValidatorI(t)

		ms.StakingKeeper.On("Validator", ctx, val).Return(vali).Twice()
		vali.On("GetOperator").Return(val).Once()

		err := k.Register(ctx, val)
		require.NoError(t, err)
		t.Run("it returns an error if validator is already registered", func(t *testing.T) {
			err := k.Register(ctx, val)
			require.ErrorIs(t, err, ErrValidatorAlreadyRegistered)
		})
	})

}

func TestCreatingSnapshots(t *testing.T) {
	val1, val2 := sdk.ValAddress("validator1"), sdk.ValAddress("validator2")
	vali1, vali2 := mocks.NewStakingValidatorI(t), mocks.NewStakingValidatorI(t)

	vali1.On("GetOperator").Return(val1).Once()
	vali2.On("GetOperator").Return(val2).Once()

	k, ms, ctx := newValsetKeeper(t)

	ms.StakingKeeper.On("Validator", ctx, val1).Return(vali1).Once()
	ms.StakingKeeper.On("Validator", ctx, val2).Return(vali2).Once()

	err := k.Register(ctx, val1)
	require.NoError(t, err)

	err = k.Register(ctx, val2)
	require.NoError(t, err)

	err = k.CreateSnapshot(ctx)
	require.NoError(t, err)

	t.Run("getting the snapshot gets all validators", func(t *testing.T) {
		snapshot, err := k.GetCurrentSnapshot(ctx)
		require.NoError(t, err)
		require.Len(t, snapshot.Validators, 2)
		require.Equal(t, snapshot.Validators[0].Address, val1.String())
		require.Equal(t, snapshot.Validators[1].Address, val2.String())
	})

}
