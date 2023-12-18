package keeper

import (
	"testing"

	"cosmossdk.io/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutilmocks "github.com/palomachain/paloma/util/keeper/mocks"
	"github.com/palomachain/paloma/x/treasury/types"
	"github.com/palomachain/paloma/x/treasury/types/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestKeeper_QueryFees(t *testing.T) {
	testcases := []struct {
		name        string
		setup       func() (Keeper, *types.QueryFeesRequest)
		expected    *types.Fees
		expectedErr error
	}{
		{
			name: "returns fees when set",
			setup: func() (Keeper, *types.QueryFeesRequest) {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load", mock.Anything, mock.Anything, mock.Anything).Return(&types.Fees{
					CommunityFundFee: "0.01",
					SecurityFee:      "0.02",
				}, nil)

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k, &types.QueryFeesRequest{}
			},
			expected: &types.Fees{
				CommunityFundFee: "0.01",
				SecurityFee:      "0.02",
			},
		},
		{
			name: "returns error when nil request",
			setup: func() (Keeper, *types.QueryFeesRequest) {
				return Keeper{}, nil
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid request"),
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			k, req := tt.setup()
			actual, actualErr := k.QueryFees(ctx, req)

			asserter.Equal(tt.expected, actual)
			asserter.Equal(tt.expectedErr, actualErr)
		})
	}
}
