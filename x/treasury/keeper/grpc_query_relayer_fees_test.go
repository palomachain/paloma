package keeper

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/btcutil/bech32"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutilmocks "github.com/palomachain/paloma/util/keeper/mocks"
	"github.com/palomachain/paloma/x/treasury/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestKeeper_QueryRelayerFee(t *testing.T) {
	val := sdk.ValAddress("validator-1")
	testcases := []struct {
		expectedErr error
		setup       func() (Keeper, *types.QueryRelayerFeeRequest)
		expected    *types.RelayerFee
		name        string
	}{
		{
			name: "returns fees when set",
			setup: func() (Keeper, *types.QueryRelayerFeeRequest) {
				sm := keeperutilmocks.NewKVStoreWrapper[*types.RelayerFee](t)
				sm.On("Get", mock.Anything, mock.Anything).Return(&types.RelayerFee{
					Validator: val.String(),
					Fee: sdk.Coin{
						Denom:  "test",
						Amount: math.NewInt(100),
					},
				}, nil)

				k := Keeper{
					relayerFeeStore: sm,
				}

				return k, &types.QueryRelayerFeeRequest{
					Validator: val.String(),
				}
			},
			expected: &types.RelayerFee{
				Validator: val.String(),
				Fee: sdk.Coin{
					Denom:  "test",
					Amount: math.NewInt(100),
				},
			},
		},
		{
			name: "returns nil when not set",
			setup: func() (Keeper, *types.QueryRelayerFeeRequest) {
				sm := keeperutilmocks.NewKVStoreWrapper[*types.RelayerFee](t)
				sm.On("Get", mock.Anything, mock.Anything).Return((*types.RelayerFee)(nil), nil)

				k := Keeper{
					relayerFeeStore: sm,
				}

				return k, &types.QueryRelayerFeeRequest{
					Validator: val.String(),
				}
			},
		},
		{
			name: "returns error when store returns error",
			setup: func() (Keeper, *types.QueryRelayerFeeRequest) {
				sm := keeperutilmocks.NewKVStoreWrapper[*types.RelayerFee](t)
				sm.On("Get", mock.Anything, mock.Anything).Return((*types.RelayerFee)(nil), fmt.Errorf("fail"))

				k := Keeper{
					relayerFeeStore: sm,
				}

				return k, &types.QueryRelayerFeeRequest{
					Validator: val.String(),
				}
			},
			expectedErr: fmt.Errorf("fail"),
		},
		{
			name: "returns error when invalid validator address request",
			setup: func() (Keeper, *types.QueryRelayerFeeRequest) {
				return Keeper{}, &types.QueryRelayerFeeRequest{
					Validator: "val-1",
				}
			},
			expectedErr: fmt.Errorf("decoding bech32 failed: %w", bech32.ErrInvalidLength(5)),
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(_ *testing.T) {
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			k, req := tt.setup()
			actual, actualErr := k.QueryRelayerFee(ctx, req)

			asserter.Equal(tt.expected, actual)
			asserter.Equal(tt.expectedErr, actualErr)
		})
	}
}

func TestKeeper_QueryRelayerFees(t *testing.T) {
	vals := []sdk.ValAddress{
		sdk.ValAddress("validator-1"),
		sdk.ValAddress("validator-2"),
		sdk.ValAddress("validator-3"),
	}
	testcases := []struct {
		expectedErr error
		setup       func() Keeper
		expected    *types.QueryRelayerFeesResponse
		name        string
	}{
		{
			name: "returns fees when set",
			setup: func() Keeper {
				sm := keeperutilmocks.NewKVStoreWrapper[*types.RelayerFee](t)
				sm.On("Iterate", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					callback := args.Get(1).(func([]byte, *types.RelayerFee) bool)
					callback(nil, &types.RelayerFee{
						Validator: vals[0].String(),
						Fee: sdk.Coin{
							Denom:  "test",
							Amount: math.NewInt(100),
						},
					})
					callback(nil, &types.RelayerFee{
						Validator: vals[1].String(),
						Fee: sdk.Coin{
							Denom:  "test",
							Amount: math.NewInt(500),
						},
					})
					callback(nil, &types.RelayerFee{
						Validator: vals[2].String(),
						Fee: sdk.Coin{
							Denom:  "test",
							Amount: math.NewInt(50),
						},
					})
				}).Return(nil)

				k := Keeper{
					relayerFeeStore: sm,
				}

				return k
			},
			expected: &types.QueryRelayerFeesResponse{
				Fees: []types.RelayerFee{
					{
						Validator: vals[0].String(),
						Fee: sdk.Coin{
							Denom:  "test",
							Amount: math.NewInt(100),
						},
					},
					{
						Validator: vals[1].String(),
						Fee: sdk.Coin{
							Denom:  "test",
							Amount: math.NewInt(500),
						},
					},
					{
						Validator: vals[2].String(),
						Fee: sdk.Coin{
							Denom:  "test",
							Amount: math.NewInt(50),
						},
					},
				},
			},
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(_ *testing.T) {
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			k := tt.setup()
			actual, actualErr := k.QueryRelayerFees(ctx, &types.Empty{})

			asserter.Equal(tt.expected, actual)
			asserter.Equal(tt.expectedErr, actualErr)
		})
	}
}
