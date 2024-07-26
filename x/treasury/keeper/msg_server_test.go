package keeper

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	keeperutilmocks "github.com/palomachain/paloma/util/keeper/mocks"
	"github.com/palomachain/paloma/x/treasury/types"
	"github.com/palomachain/paloma/x/treasury/types/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpsertRelayerFee(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("paloma", "pub")
	config.SetBech32PrefixForValidator("palomavaloper", "valoperpub")

	valAddr, err := sdk.ValAddressFromBech32("palomavaloper1tsu8nthuspe4zlkejtj3v27rtq8qz7q6983zt2")
	require.NoError(t, err)

	testcases := []struct {
		name        string
		setup       func() msgServer
		input       *types.RelayerFeeSetting
		expectedErr error
	}{
		{
			name: "with invalid validator address",
			input: &types.RelayerFeeSetting{
				ValAddress: "test",
			},
			setup: func() msgServer {
				k := msgServer{
					Keeper: Keeper{},
				}

				return k
			},
			expectedErr: fmt.Errorf("decoding bech32 failed"),
		},
		{
			name: "with keeper error",
			input: &types.RelayerFeeSetting{
				ValAddress: valAddr.String(),
			},
			setup: func() msgServer {
				evm := mocks.NewEvmKeeper(t)
				m := keeperutilmocks.NewKVStoreWrapper[*types.RelayerFeeSetting](t)
				m.On("Get", mock.Anything, mock.Anything).Return(&types.RelayerFeeSetting{}, fmt.Errorf("FAIL"))
				k := msgServer{
					Keeper: Keeper{
						relayerFees: m,
						evm:         evm,
					},
				}

				return k
			},
			expectedErr: fmt.Errorf("FAIL"),
		},
		{
			name: "with unknown chain",
			input: &types.RelayerFeeSetting{
				ValAddress: valAddr.String(),
				Fees: []types.RelayerFeeSetting_FeeSetting{
					{
						ChainReferenceId: "test-chain",
						Multiplicator:    math.LegacyMustNewDecFromStr("1.2"),
					},
				},
			},
			setup: func() msgServer {
				evm := mocks.NewEvmKeeper(t)
				m := keeperutilmocks.NewKVStoreWrapper[*types.RelayerFeeSetting](t)
				m.On("Get", mock.Anything, mock.Anything).Return(&types.RelayerFeeSetting{}, keeperutil.ErrNotFound)
				m.On("Set",
					mock.Anything,
					valAddr,
					&types.RelayerFeeSetting{
						ValAddress: valAddr.String(),
						Fees: []types.RelayerFeeSetting_FeeSetting{
							{
								ChainReferenceId: "test-chain",
								Multiplicator:    math.LegacyMustNewDecFromStr("1.2"),
							},
						},
					},
				).Return(nil)
				k := msgServer{
					Keeper: Keeper{
						relayerFees: m,
						evm:         evm,
					},
				}

				return k
			},
			expectedErr: nil,
		},
		{
			name: "with no existing record",
			input: &types.RelayerFeeSetting{
				ValAddress: valAddr.String(),
				Fees: []types.RelayerFeeSetting_FeeSetting{
					{
						ChainReferenceId: "test-chain",
						Multiplicator:    math.LegacyMustNewDecFromStr("1.2"),
					},
				},
			},
			setup: func() msgServer {
				evm := mocks.NewEvmKeeper(t)
				m := keeperutilmocks.NewKVStoreWrapper[*types.RelayerFeeSetting](t)
				m.On("Get", mock.Anything, mock.Anything).Return(&types.RelayerFeeSetting{}, keeperutil.ErrNotFound)
				m.On("Set",
					mock.Anything,
					valAddr,
					&types.RelayerFeeSetting{
						ValAddress: valAddr.String(),
						Fees: []types.RelayerFeeSetting_FeeSetting{
							{
								ChainReferenceId: "test-chain",
								Multiplicator:    math.LegacyMustNewDecFromStr("1.2"),
							},
						},
					},
				).Return(nil)
				k := msgServer{
					Keeper: Keeper{
						relayerFees: m,
						evm:         evm,
					},
				}

				return k
			},
			expectedErr: nil,
		},
		{
			name: "with preexisting record",
			input: &types.RelayerFeeSetting{
				ValAddress: valAddr.String(),
				Fees: []types.RelayerFeeSetting_FeeSetting{
					{
						ChainReferenceId: "test-chain",
						Multiplicator:    math.LegacyMustNewDecFromStr("1.2"),
					},
					{
						ChainReferenceId: "test-chain-3",
						Multiplicator:    math.LegacyMustNewDecFromStr("0.2"),
					},
				},
			},
			setup: func() msgServer {
				evm := mocks.NewEvmKeeper(t)
				m := keeperutilmocks.NewKVStoreWrapper[*types.RelayerFeeSetting](t)
				m.On("Get", mock.Anything, mock.Anything).Return(&types.RelayerFeeSetting{
					ValAddress: valAddr.String(),
					Fees: []types.RelayerFeeSetting_FeeSetting{
						{
							ChainReferenceId: "test-chain-2",
							Multiplicator:    math.LegacyMustNewDecFromStr("1.5"),
						},
						{
							ChainReferenceId: "test-chain",
							Multiplicator:    math.LegacyMustNewDecFromStr("1.6"),
						},
					},
				}, nil)
				m.On("Set",
					mock.Anything,
					valAddr,
					&types.RelayerFeeSetting{
						ValAddress: valAddr.String(),
						Fees: []types.RelayerFeeSetting_FeeSetting{
							{
								ChainReferenceId: "test-chain",
								Multiplicator:    math.LegacyMustNewDecFromStr("1.2"),
							},
							{
								ChainReferenceId: "test-chain-3",
								Multiplicator:    math.LegacyMustNewDecFromStr("0.2"),
							},
							{
								ChainReferenceId: "test-chain-2",
								Multiplicator:    math.LegacyMustNewDecFromStr("1.5"),
							},
						},
					},
				).Return(nil)
				k := msgServer{
					Keeper: Keeper{
						relayerFees: m,
						evm:         evm,
					},
				}

				return k
			},
			expectedErr: nil,
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			k := tt.setup()

			_, actualErr := k.UpsertRelayerFee(ctx, &types.MsgUpsertRelayerFee{FeeSetting: tt.input})
			if tt.expectedErr != nil {
				asserter.ErrorContains(actualErr, tt.expectedErr.Error())
			} else {
				asserter.NoError(actualErr)
			}
		})
	}
}
