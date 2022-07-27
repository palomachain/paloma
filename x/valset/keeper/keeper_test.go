package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/palomachain/paloma/x/valset/types/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var (
	priv1 = secp256k1.GenPrivKey()
	pk1   = priv1.PubKey()

	priv2 = secp256k1.GenPrivKey()
	pk2   = priv2.PubKey()
)

func TestIfValidatorCanBeAccepted(t *testing.T) {
	k, ms, ctx := newValsetKeeper(t)
	val := sdk.ValAddress("validator")
	nonExistingVal := sdk.ValAddress("i dont exist")

	vali := mocks.NewStakingValidatorI(t)

	ms.StakingKeeper.On("Validator", mock.Anything, val).Return(vali)
	ms.StakingKeeper.On("Validator", mock.Anything, nonExistingVal).Return(nil)

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
}
func TestRegisterRunner(t *testing.T) {
	k, ms, ctx := newValsetKeeper(t)
	val := sdk.ValAddress("validator")
	val2 := sdk.ValAddress("validator2")
	nonExistingVal := sdk.ValAddress("i dont exist")

	vali := mocks.NewStakingValidatorI(t)
	vali.On("IsJailed").Return(false)
	vali.On("IsBonded").Return(true)

	vali2 := mocks.NewStakingValidatorI(t)
	vali2.On("IsJailed").Return(false)
	vali2.On("IsBonded").Return(true)

	ms.StakingKeeper.On("Validator", mock.Anything, val).Return(vali)
	ms.StakingKeeper.On("Validator", mock.Anything, val2).Return(vali2)
	ms.StakingKeeper.On("Validator", mock.Anything, nonExistingVal).Return(nil)

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

		externalChainInfo, err := k.getValidatorChainInfos(ctx, val)
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

	vali1.On("GetOperator").Return(val1)
	vali2.On("GetOperator").Return(val2)

	vali1.On("IsJailed").Return(false)
	vali2.On("IsJailed").Return(false)

	vali1.On("GetBondedTokens").Return(sdk.NewInt(888))
	vali2.On("GetBondedTokens").Return(sdk.NewInt(222))

	vali1.On("IsBonded").Return(true)
	vali2.On("IsBonded").Return(true)

	k, ms, ctx := newValsetKeeper(t)

	ms.StakingKeeper.On("Validator", ctx, val1).Return(vali1)
	ms.StakingKeeper.On("Validator", ctx, val2).Return(vali2)

	ms.StakingKeeper.On("IterateValidators", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		type fnc = func(int64, stakingtypes.ValidatorI) bool
		f := args.Get(1).(fnc)
		f(0, vali1)
		f(1, vali2)
	})

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

	gotState1, err := k.getValidatorChainInfos(ctx, val1)
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

	gotState2, err := k.getValidatorChainInfos(ctx, val1)
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
		gotState, err := k.getValidatorChainInfos(ctx, val2)
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
				TotalShares: sdk.NewInt(100),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdk.NewInt(20)},
					{Address: sdk.ValAddress("456"), ShareCount: sdk.NewInt(80)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdk.NewInt(100),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdk.NewInt(80)},
					{Address: sdk.ValAddress("456"), ShareCount: sdk.NewInt(20)},
				},
			},
		},
		{
			expRes: true,
			name:   "two snapshots have same validators and same relative power orders, but differ in their absolute power more than 1 percent",
			curr: &types.Snapshot{
				TotalShares: sdk.NewInt(100),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdk.NewInt(20)},
					{Address: sdk.ValAddress("456"), ShareCount: sdk.NewInt(80)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdk.NewInt(100),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdk.NewInt(30)},
					{Address: sdk.ValAddress("456"), ShareCount: sdk.NewInt(80)},
				},
			},
		},
		{
			expRes: false,
			name:   "two snapshots have same validators and same relative power orders",
			curr: &types.Snapshot{
				TotalShares: sdk.NewInt(100),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdk.NewInt(20)},
					{Address: sdk.ValAddress("456"), ShareCount: sdk.NewInt(80)},
				},
			},
			neww: &types.Snapshot{
				TotalShares: sdk.NewInt(1000),
				Validators: []types.Validator{
					{Address: sdk.ValAddress("123"), ShareCount: sdk.NewInt(200)},
					{Address: sdk.ValAddress("456"), ShareCount: sdk.NewInt(800)},
				},
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var k Keeper
			res := k.isNewSnapshotWorthy(tt.curr, tt.neww)

			require.Equal(t, tt.expRes, res)
		})
	}
}
