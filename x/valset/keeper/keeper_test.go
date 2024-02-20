package keeper

import (
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/palomachain/paloma/x/valset/types/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIfValidatorCanBeAccepted(t *testing.T) {
	k, ms, ctx := newValsetKeeper(t)
	val := sdk.ValAddress("validator")
	nonExistingVal := sdk.ValAddress("i dont exist")

	vali := mocks.NewStakingValidatorI(t)

	ms.StakingKeeper.On("Validator", mock.Anything, val).Return(vali, nil)
	ms.StakingKeeper.On("Validator", mock.Anything, nonExistingVal).Return(nil, nil)

	t.Run("when validator is jailed it returns an error", func(t *testing.T) {
		vali.On("IsJailed").Return(true).Once()
		err := k.CanAcceptValidator(ctx, val)
		require.ErrorIs(t, err, ErrValidatorCannotBePigeon)
	})

	t.Run("when validator is not bonded it returns an error", func(t *testing.T) {
		vali.On("IsJailed").Return(false)
		vali.On("IsBonded").Return(false)
		err := k.CanAcceptValidator(ctx, val)
		require.ErrorIs(t, err, ErrValidatorCannotBePigeon)
	})

	t.Run("when validator is jailed it returns an error", func(t *testing.T) {
		err := k.CanAcceptValidator(ctx, val)
		require.ErrorIs(t, err, ErrValidatorCannotBePigeon)
	})

	t.Run("when validator does not exist it return an error", func(t *testing.T) {
		err := k.CanAcceptValidator(ctx, nonExistingVal)
		require.ErrorIs(t, err, ErrValidatorWithAddrNotFound)
	})

	// t.Run("Retu ")
}

func TestRegisteringPigeon(t *testing.T) {
	k, ms, ctx := newValsetKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ctx = sdkCtx.WithBlockHeight(3000)
	val := sdk.ValAddress("validator")
	val2 := sdk.ValAddress("validator2")
	nonExistingVal := sdk.ValAddress("i dont exist")

	vali := mocks.NewStakingValidatorI(t)
	vali.On("IsJailed").Return(false)
	vali.On("IsBonded").Return(true)

	vali2 := mocks.NewStakingValidatorI(t)
	vali2.On("IsJailed").Return(false)
	vali2.On("IsBonded").Return(true)

	ms.StakingKeeper.On("Validator", mock.Anything, val).Return(vali, nil)
	ms.StakingKeeper.On("Validator", mock.Anything, val2).Return(vali2, nil)
	ms.StakingKeeper.On("Validator", mock.Anything, nonExistingVal).Return(nil, nil)

	t.Run("if validator has been alive before, but it's not now, then it returns an error", func(t *testing.T) {
		err := k.KeepValidatorAlive(sdkCtx.WithBlockHeight(500), val, "v1.4.0")
		require.NoError(t, err)
		alive, err := k.IsValidatorAlive(ctx, val)
		require.NoError(t, err)
		require.False(t, alive)
	})

	t.Run("setting the validator to current ctx's block time", func(t *testing.T) {
		err := k.KeepValidatorAlive(ctx, val, "v1.4.0")
		require.NoError(t, err)
		err = k.KeepValidatorAlive(ctx, val2, "v1.4.0")
		require.NoError(t, err)
	})

	t.Run("it adds a new chain info to the validator", func(t *testing.T) {
		err := k.AddExternalChainInfo(
			ctx,
			val,
			[]*types.ExternalChainInfo{
				{
					ChainReferenceID: "chain-1",
					Address:          "addr1",
					ChainType:        "evm",
				},
			},
		)
		require.NoError(t, err)

		externalChainInfo, err := k.GetValidatorChainInfos(ctx, val)
		require.NoError(t, err)

		require.Equal(
			t,
			externalChainInfo,
			[]*types.ExternalChainInfo{
				{
					ChainReferenceID: "chain-1",
					Address:          "addr1",
					ChainType:        "evm",
				},
			},
		)
	})

	stateForVal1 := []*types.ExternalChainInfo{
		{
			ChainReferenceID: "chain-1",
			Address:          "addr1",
			ChainType:        "evm",
		},
	}
	t.Run("it overrides the state if we are writing new data for existing validator", func(t *testing.T) {
		err := k.AddExternalChainInfo(
			ctx,
			val,
			stateForVal1,
		)
		require.ErrorIs(t, err, nil)
	})

	t.Run("it returns an error if we try to add chain info that other validator already has", func(t *testing.T) {
		err := k.AddExternalChainInfo(
			ctx,
			val2,
			stateForVal1,
		)
		require.ErrorIs(t, err, ErrExternalChainAlreadyRegistered)
	})

	t.Run("it returns an error if we try to add to the validator which does not exist", func(t *testing.T) {
		err := k.AddExternalChainInfo(ctx,
			nonExistingVal,
			[]*types.ExternalChainInfo{
				{
					ChainReferenceID: "chain-1",
					Address:          "addr1",
				},
			},
		)
		require.ErrorIs(t, err, ErrValidatorWithAddrNotFound)
	})
}

func TestCreatingSnapshots(t *testing.T) {
	val1, val2 := sdk.ValAddress("validator1"), sdk.ValAddress("validator2")
	vali1, vali2 := mocks.NewStakingValidatorI(t), mocks.NewStakingValidatorI(t)

	vali1.On("GetOperator").Return(val1.String())
	vali2.On("GetOperator").Return(val2.String())

	vali1.On("IsJailed").Return(false)
	vali2.On("IsJailed").Return(false)

	vali1.On("GetBondedTokens").Return(sdkmath.NewInt(888))
	vali2.On("GetBondedTokens").Return(sdkmath.NewInt(222))

	vali1.On("IsBonded").Return(true)
	vali2.On("IsBonded").Return(true)

	k, ms, ctx := newValsetKeeper(t)
	ms.StakingKeeper.On("Validator", ctx, val1).Return(vali1, nil)
	ms.StakingKeeper.On("Validator", ctx, val2).Return(vali2, nil)

	ms.StakingKeeper.On("IterateValidators", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		type fnc = func(int64, stakingtypes.ValidatorI) bool
		f := args.Get(1).(fnc)
		f(0, vali1)
		f(1, vali2)
	}).Return(nil)

	ms.EvmKeeper.On("MissingChains", mock.Anything, mock.Anything).Return([]string(nil), nil)
	ms.EvmKeeper.On("MissingChains", mock.Anything, mock.Anything).Return([]string(nil), nil)

	state1 := []*types.ExternalChainInfo{
		{
			ChainType:        "evm",
			ChainReferenceID: "123",
			Address:          val1.String(),
			Pubkey:           []byte("whoop"),
		},
	}
	err := k.AddExternalChainInfo(ctx, val1, state1)
	require.NoError(t, err)

	gotState1, err := k.GetValidatorChainInfos(ctx, val1)
	require.NoError(t, err)
	require.Equal(t, state1, gotState1)

	state2 := []*types.ExternalChainInfo{
		{
			ChainType:        "evm",
			ChainReferenceID: "123",
			Address:          val1.String(),
			Pubkey:           []byte("123"),
		},
		{
			ChainType:        "evm",
			ChainReferenceID: "123",
			Address:          val1.String(),
			Pubkey:           []byte("456"),
		},
	}

	err = k.AddExternalChainInfo(ctx, val1, state2)
	require.NoError(t, err)

	gotState2, err := k.GetValidatorChainInfos(ctx, val1)
	require.NoError(t, err)
	require.Equal(t, state2, gotState2)

	err = k.AddExternalChainInfo(ctx, val2, []*types.ExternalChainInfo{
		{
			ChainType:        "evm",
			ChainReferenceID: "567",
			Address:          val2.String(),
			Pubkey:           []byte("6666"),
		},
	})
	require.NoError(t, err)

	t.Run("adding address which already exists simply overrides the state", func(t *testing.T) {
		state := []*types.ExternalChainInfo{
			{
				ChainType:        "evm",
				ChainReferenceID: "56799",
				Address:          val2.String(),
				Pubkey:           []byte("6666777"),
			},
		}
		err = k.AddExternalChainInfo(ctx, val2, state)
		require.ErrorIs(t, err, nil)
		gotState, err := k.GetValidatorChainInfos(ctx, val2)
		require.NoError(t, err)
		require.Equal(t, state, gotState)
	})

	k.TriggerSnapshotBuild(ctx)

	t.Run("getting the snapshot gets all validators", func(t *testing.T) {
		snapshot, err := k.GetCurrentSnapshot(ctx)
		require.NoError(t, err)
		require.Len(t, snapshot.Validators, 2)
		require.Equal(t, snapshot.Validators[0].Address, val1)
		require.Equal(t, snapshot.Validators[1].Address, val2)
	})
}

func TestIsNewSnapshotWorthy(t *testing.T) {
	testcases := []struct {
		name   string
		curr   *types.Snapshot
		neww   *types.Snapshot
		expRes bool
	}{
		{
			expRes: true,
			name:   "current snapshot is nil",
		},
		{
			expRes: true,
			name:   "two snapshots have different validator sizes",
			curr: &types.Snapshot{
				Validators: []types.Validator{
					{},
				},
			},
			neww: &types.Snapshot{
				Validators: []types.Validator{
					{}, {},
				},
			},
		},
		{
			expRes: true,
			name:   "two snapshots are the same length but have different validators",
			curr: &types.Snapshot{
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123")},
					{Address: sdk.ValAddress("456")},
				},
			},
			neww: &types.Snapshot{
				Validators: []types.Validator{
					{Address: sdk.ValAddress("abc")},
					{Address: sdk.ValAddress("def")},
				},
			},
		},
		{
			expRes: true,
			name:   "two snapshots have same validators but different power orders",
			curr: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdkmath.NewInt(20)},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(80)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdkmath.NewInt(80)},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(20)},
				},
			},
		},
		{
			expRes: true,
			name:   "two snapshots have same validators and same relative power orders, but differ in their absolute power more than 1 percent",
			curr: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdkmath.NewInt(20)},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(80)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdkmath.NewInt(30)},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(80)},
				},
			},
		},
		{
			expRes: true,
			name:   "two snapshots are same, but they have removed one of their accounts",
			curr: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{
						Address:    sdk.ValAddress("123"),
						ShareCount: sdkmath.NewInt(20),
						ExternalChainInfos: []*types.ExternalChainInfo{
							{}, {},
						},
					},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(80)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{
						Address:    sdk.ValAddress("123"),
						ShareCount: sdkmath.NewInt(20),
						ExternalChainInfos: []*types.ExternalChainInfo{
							{},
						},
					},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(80)},
				},
			},
		},
		{
			expRes: true,
			name:   "two snapshots are same, but they have different external chain infos",
			curr: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{
						Address:    sdk.ValAddress("123"),
						ShareCount: sdkmath.NewInt(20),
						ExternalChainInfos: []*types.ExternalChainInfo{
							{Address: "abc"}, {Address: "def"},
						},
					},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(80)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{
						Address:    sdk.ValAddress("123"),
						ShareCount: sdkmath.NewInt(20),
						ExternalChainInfos: []*types.ExternalChainInfo{
							{Address: "abc"}, {Address: "123"},
						},
					},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(80)},
				},
			},
		},
		{
			expRes: false,
			name:   "two snapshots have same validators and same relative power orders",
			curr: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{
						Address:    sdk.ValAddress("123"),
						ShareCount: sdkmath.NewInt(20),
						ExternalChainInfos: []*types.ExternalChainInfo{
							{Address: "abc"}, {Address: "def"},
						},
					},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(80)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdkmath.NewInt(1000),
				Validators: []types.Validator{
					{
						Address:    sdk.ValAddress("123"),
						ShareCount: sdkmath.NewInt(200),
						ExternalChainInfos: []*types.ExternalChainInfo{
							{Address: "abc"}, {Address: "def"},
						},
					},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(800)},
				},
			},
		},
		{
			expRes: false,
			name:   "two snapshots have same validators and same traits",
			curr: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{
						Address:    sdk.ValAddress("123"),
						ShareCount: sdkmath.NewInt(20),
						ExternalChainInfos: []*types.ExternalChainInfo{
							{Address: "abc", Traits: []string{"abc"}}, {Address: "def", Traits: []string{"abc"}},
						},
					},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(80)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdkmath.NewInt(1000),
				Validators: []types.Validator{
					{
						Address:    sdk.ValAddress("123"),
						ShareCount: sdkmath.NewInt(200),
						ExternalChainInfos: []*types.ExternalChainInfo{
							{Address: "abc", Traits: []string{"abc"}}, {Address: "def", Traits: []string{"abc"}},
						},
					},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(800)},
				},
			},
		},
		{
			expRes: true,
			name:   "two snapshots have same validators and different traits",
			curr: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{
						Address:    sdk.ValAddress("123"),
						ShareCount: sdkmath.NewInt(20),
						ExternalChainInfos: []*types.ExternalChainInfo{
							{Address: "abc", Traits: []string{"abc"}}, {Address: "def", Traits: []string{"abc"}},
						},
					},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(80)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdkmath.NewInt(1000),
				Validators: []types.Validator{
					{
						Address:    sdk.ValAddress("123"),
						ShareCount: sdkmath.NewInt(200),
						ExternalChainInfos: []*types.ExternalChainInfo{
							{Address: "abc", Traits: []string{"abc"}}, {Address: "def", Traits: []string{"def"}},
						},
					},
					{Address: sdk.ValAddress("456"), ShareCount: sdkmath.NewInt(800)},
				},
			},
		},
		{
			expRes: false,
			name:   "if the powers are still the same then it's not worthy",
			curr: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdkmath.NewInt(20)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdkmath.NewInt(100),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdkmath.NewInt(20)},
				},
			},
		},
		{
			expRes: false,
			name:   "if the powers have change for less than 1%, then it's not worthy",
			curr: &types.Snapshot{
				TotalShares: sdkmath.NewInt(1000),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdkmath.NewInt(20)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdkmath.NewInt(1000),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdkmath.NewInt(21)},
				},
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var k Keeper
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			res := k.isNewSnapshotWorthy(ctx, tt.curr, tt.neww)

			require.Equal(t, tt.expRes, res)
		})
	}
}

func TestGracePeriodCoverage(t *testing.T) {
	k, _, ctx := newValsetKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockHeight(100)

	t.Run("with unjailed validator covered by grace period", func(t *testing.T) {
		for i := cJailingGracePeriodBlockHeight; i >= 0; i-- {
			val := sdk.ValAddress("validator")
			bh := sdk.Uint64ToBigEndian(uint64(sdkCtx.BlockHeight() - int64(i)))
			k.gracePeriodStore(sdkCtx).Set(val, bh)
			require.True(t, k.isValidatorInGracePeriod(sdkCtx, val))
		}
	})

	t.Run("with unjailed validator no longer covered by grace period", func(t *testing.T) {
		val := sdk.ValAddress("validator")
		bh := sdk.Uint64ToBigEndian(uint64(sdkCtx.BlockHeight() - cJailingGracePeriodBlockHeight - 1))
		k.gracePeriodStore(sdkCtx).Set(val, bh)
		require.False(t, k.isValidatorInGracePeriod(sdkCtx, val), "bh = %d", bh)
	})

	t.Run("with not present in grace period store", func(t *testing.T) {
		val := sdk.ValAddress("validator-bonded")
		require.False(t, k.isValidatorInGracePeriod(sdkCtx, val))
	})
}

func TestJailedValidaotors(t *testing.T) {
	k, ms, ctx := newValsetKeeper(t)
	val := sdk.ValAddress("validator")
	vali := mocks.NewStakingValidatorI(t)

	ms.StakingKeeper.On("Validator", mock.Anything, val).Return(vali, nil)
	t.Run("Return true when validator is jailed", func(t *testing.T) {
		vali.On("IsJailed").Return(false).Once()
		flag, err := k.IsJailed(ctx, val)
		require.NoError(t, err)
		require.False(t, flag)
	})
	t.Run("Return true when the validator is jailed", func(t *testing.T) {
		vali.On("IsJailed").Return(true).Once()
		err := k.Jail(ctx, val, "i don't know")
		require.Error(t, err)
	})
}
