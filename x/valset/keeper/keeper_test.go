package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/vizualni/whoops"
	"github.com/volumefi/cronchain/x/valset/types"
	"github.com/volumefi/cronchain/x/valset/types/mocks"
)

var (
	priv1 = ed25519.GenPrivKey()
	pk1   = priv1.PubKey()

	priv2 = ed25519.GenPrivKey()
	pk2   = priv2.PubKey()
)

func TestRegisterRunner(t *testing.T) {
	val := sdk.ValAddress([]byte("validator"))
	t.Run("it returns error if validator does not exist", func(t *testing.T) {
		k, ms, ctx := newValsetKeeper(t)

		// returns no validator
		ms.StakingKeeper.On("Validator", ctx, val).Return(nil).Once()

		err := k.Register(ctx, &types.MsgRegisterConductor{
			Creator:      val.String(),
			PubKey:       pk1.Bytes(),
			SignedPubKey: whoops.Must(priv1.Sign(pk1.Bytes())),
		})
		require.ErrorIs(t, err, ErrValidatorWithAddrNotFound)
	})

	t.Run("it returns error if signature is not valid", func(t *testing.T) {
		k, ms, ctx := newValsetKeeper(t)
		vali := mocks.NewStakingValidatorI(t)

		ms.StakingKeeper.On("Validator", ctx, val).Return(vali).Once()

		err := k.Register(ctx, &types.MsgRegisterConductor{
			Creator:      val.String(),
			PubKey:       pk1.Bytes(),
			SignedPubKey: []byte("invalid signature"),
		})
		require.ErrorIs(t, err, ErrPublicKeyOrSignatureIsInvalid)
	})

	t.Run("it registers the validator", func(t *testing.T) {
		k, ms, ctx := newValsetKeeper(t)
		val := sdk.ValAddress("validator")

		vali := mocks.NewStakingValidatorI(t)

		ms.StakingKeeper.On("Validator", ctx, val).Return(vali).Twice()
		vali.On("GetOperator").Return(val).Once()

		err := k.Register(ctx, &types.MsgRegisterConductor{
			Creator:      val.String(),
			PubKey:       pk1.Bytes(),
			SignedPubKey: whoops.Must(priv1.Sign(pk1.Bytes())),
		})
		require.NoError(t, err)

		t.Run("it returns an error if validator is already registered", func(t *testing.T) {
			err := k.Register(ctx, &types.MsgRegisterConductor{
				Creator:      val.String(),
				PubKey:       pk1.Bytes(),
				SignedPubKey: whoops.Must(priv1.Sign(pk1.Bytes())),
			})
			require.ErrorIs(t, err, ErrValidatorAlreadyRegistered)
		})

		t.Run("it adds a new chain info to the validator", func(t *testing.T) {
			err := k.addExternalChainInfo(ctx, &types.MsgAddExternalChainInfoForValidator{
				Creator: string(val.Bytes()),
				ChainInfos: []*types.MsgAddExternalChainInfoForValidator_ChainInfo{
					{
						ChainID: "chain-1",
						Address: "addr1",
					},
				},
			})
			require.NoError(t, err)
			validator, err := k.getValidator(ctx, val)
			require.NoError(t, err)
			require.Len(t, validator.ExternalChainInfos, 1)
			require.Equal(t, validator.ExternalChainInfos[0].ChainID, "chain-1")
			require.Equal(t, validator.ExternalChainInfos[0].Address, "addr1")
		})

		t.Run("it returns an error if chain info is chain info already exists", func(t *testing.T) {
			err := k.addExternalChainInfo(ctx, &types.MsgAddExternalChainInfoForValidator{
				Creator: string(val.Bytes()),
				ChainInfos: []*types.MsgAddExternalChainInfoForValidator_ChainInfo{
					{
						ChainID: "chain-1",
						Address: "addr1",
					},
				},
			})
			require.ErrorIs(t, err, ErrExternalChainAlreadyRegistered)
		})

		t.Run("it returns an error if we try to add to the validator which does not exist", func(t *testing.T) {
			err := k.addExternalChainInfo(ctx, &types.MsgAddExternalChainInfoForValidator{
				Creator: "i don't exist",
				ChainInfos: []*types.MsgAddExternalChainInfoForValidator_ChainInfo{
					{
						ChainID: "chain-1",
						Address: "addr1",
					},
				},
			})
			require.Error(t, err)
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

	err := k.Register(ctx, &types.MsgRegisterConductor{
		Creator:      val1.String(),
		PubKey:       pk1.Bytes(),
		SignedPubKey: whoops.Must(priv1.Sign(pk1.Bytes())),
	})
	require.NoError(t, err)

	err = k.Register(ctx, &types.MsgRegisterConductor{
		Creator:      val2.String(),
		PubKey:       pk2.Bytes(),
		SignedPubKey: whoops.Must(priv2.Sign(pk2.Bytes())),
	})
	require.NoError(t, err)

	err = k.TriggerSnapshotBuild(ctx)
	require.NoError(t, err)

	t.Run("getting the snapshot gets all validators", func(t *testing.T) {
		snapshot, err := k.GetCurrentSnapshot(ctx)
		require.NoError(t, err)
		require.Len(t, snapshot.Validators, 2)
		require.Equal(t, snapshot.Validators[0].Address, val1.String())
		require.Equal(t, snapshot.Validators[1].Address, val2.String())
	})

}
