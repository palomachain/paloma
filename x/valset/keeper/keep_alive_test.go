package keeper

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/palomachain/paloma/x/valset/types/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestJailingInactiveValidators(t *testing.T) {
	k, ms, ctx := newValsetKeeper(t)
	ctx = ctx.WithBlockTime(time.Unix(1000000000, 0))

	valBuild := func(id int, alive bool) *mocks.StakingValidatorI {
		val := sdk.ValAddress(fmt.Sprintf("validator_%d", id))
		vali := mocks.NewStakingValidatorI(t)
		ms.StakingKeeper.On("Validator", mock.Anything, val).Return(vali)
		vali.On("IsJailed").Return(false)
		vali.On("IsBonded").Return(true)
		consAddr := sdk.ConsAddress(val)
		if alive {
			err := k.KeepValidatorAlive(ctx.WithBlockTime(time.Unix(1000, 0)), val)
			require.NoError(t, err)

			vali.On("GetConsAddr").Return(consAddr, nil)
			ms.StakingKeeper.On("Jail", mock.Anything, consAddr)
		} else {
			err := k.KeepValidatorAlive(ctx.WithBlockTime(ctx.BlockTime().Add(-defaultKeepAliveDuration/2)), val)
			require.NoError(t, err)
		}
		return vali
	}

	v1 := valBuild(1, false)
	v2 := valBuild(2, false)
	v3 := valBuild(3, true)
	v4 := valBuild(4, true)

	ms.StakingKeeper.On("IterateValidators", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		callback := args.Get(1).(func(int64, stakingtypes.ValidatorI) bool)
		callback(0, v1)
		callback(0, v2)
		callback(0, v3)
		callback(0, v4)
	}).Return(false)

	err := k.JailInactiveValidators(ctx)
	require.NoError(t, err)
}
