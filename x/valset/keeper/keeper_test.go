package keeper

import (
	"fmt"
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

func TestRegisterRunner(t *testing.T) {
	k, ms, ctx := newValsetKeeper(t)
	val := sdk.ValAddress("validator")
	nonExistingVal := sdk.ValAddress("i dont exist")

	vali := mocks.NewStakingValidatorI(t)

	ms.StakingKeeper.On("Validator", mock.Anything, val).Return(vali)
	ms.StakingKeeper.On("Validator", mock.Anything, nonExistingVal).Return(nil)

	t.Run("it adds a new chain info to the validator", func(t *testing.T) {
		err := k.AddExternalChainInfo(
			ctx,
			val,
			[]*types.ExternalChainInfo{
				{
					ChainID:   "chain-1",
					Address:   "addr1",
					ChainType: "evm",
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
					ChainID:   "chain-1",
					Address:   "addr1",
					ChainType: "evm",
				},
			},
		)
	})

	t.Run("it returns an error if chain info already exists", func(t *testing.T) {
		err := k.AddExternalChainInfo(
			ctx,
			val,
			[]*types.ExternalChainInfo{
				{
					ChainID:   "chain-1",
					Address:   "addr1",
					ChainType: "evm",
				},
			},
		)
		fmt.Println(err)
		require.ErrorIs(t, err, ErrExternalChainAlreadyRegistered)
	})

	t.Run("it returns an error if we try to add to the validator which does not exist", func(t *testing.T) {
		err := k.AddExternalChainInfo(ctx,
			nonExistingVal,
			[]*types.ExternalChainInfo{
				{
					ChainID: "chain-1",
					Address: "addr1",
				},
			},
		)
		require.Error(t, err)
	})

}

func TestCreatingSnapshots(t *testing.T) {
	val1, val2 := sdk.ValAddress("validator1"), sdk.ValAddress("validator2")
	vali1, vali2 := mocks.NewStakingValidatorI(t), mocks.NewStakingValidatorI(t)

	vali1.On("GetOperator").Return(val1)
	vali2.On("GetOperator").Return(val2)

	vali1.On("GetBondedTokens").Return(sdk.NewInt(888))
	vali2.On("GetBondedTokens").Return(sdk.NewInt(222))

	k, ms, ctx := newValsetKeeper(t)

	ms.StakingKeeper.On("Validator", ctx, val1).Return(vali1)
	ms.StakingKeeper.On("Validator", ctx, val2).Return(vali2)

	ms.StakingKeeper.On("IterateBondedValidatorsByPower", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		type fnc = func(int64, stakingtypes.ValidatorI) bool
		f := args.Get(1).(fnc)
		f(0, vali1)
		f(1, vali2)
	})

	err := k.AddExternalChainInfo(ctx, val1, []*types.ExternalChainInfo{
		{
			ChainType: "evm",
			ChainID:   "123",
			Address:   val1.String(),
			Pubkey:    []byte("123"),
		},
		{
			ChainType: "evm",
			ChainID:   "123",
			Address:   val1.String(),
			Pubkey:    []byte("456"),
		},
	})
	require.NoError(t, err)
	err = k.AddExternalChainInfo(ctx, val2, []*types.ExternalChainInfo{
		{
			ChainType: "evm",
			ChainID:   "567",
			Address:   val2.String(),
			Pubkey:    []byte("123"),
		},
	})
	require.NoError(t, err)

	t.Run("adding address which already exists", func(t *testing.T) {
		err = k.AddExternalChainInfo(ctx, val1, []*types.ExternalChainInfo{
			{
				ChainType: "evm",
				ChainID:   "567",
				Address:   val2.String(),
				Pubkey:    []byte("123"),
			},
		})
		require.ErrorIs(t, err, ErrExternalChainAlreadyRegistered)
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
