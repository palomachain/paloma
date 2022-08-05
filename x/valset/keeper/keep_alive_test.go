package keeper

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/valset/types/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestJailingInactiveValidators(t *testing.T) {
	k, ms, ctx := newValsetKeeper(t)
	ctx = ctx.WithBlockTime(time.Unix(1000000000, 0))

	valBuild := func(id int, alive bool) sdk.ValAddress {
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
		return val
	}
	valBuild(1, false)
	valBuild(2, false)
	valBuild(3, true)
	valBuild(4, true)

	err := k.JailInactiveValidators(ctx)
	require.NoError(t, err)
}
