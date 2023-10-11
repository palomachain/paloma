package keeper

import (
	"errors"
	"testing"

	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authmocks "github.com/palomachain/paloma/testutil/third_party_mocks/cosmos/cosmos-sdk/x/auth/types/mocks"
	"github.com/palomachain/paloma/x/scheduler/types"
	"github.com/palomachain/paloma/x/scheduler/types/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExecuteJob(t *testing.T) {
	testcases := []struct {
		name          string
		msgServer     func() msgServer
		expected      *types.MsgExecuteJobResponse
		expectedError error
	}{
		{
			name: "full success scheduling job",
			msgServer: func() msgServer {
				schedulerKeeper := mocks.NewKeeper(t)
				mockAccount := authmocks.NewAccountI(t)
				mockAccount.On("GetAddress").Return(sdk.AccAddress("test-address"))

				schedulerKeeper.On("Logger", mock.Anything).Return(log.NewNopLogger())
				schedulerKeeper.On("GetAccount",
					mock.Anything,
					mock.Anything,
				).Return(mockAccount)

				schedulerKeeper.On("ExecuteJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				msgServer := msgServer{
					schedulerKeeper,
				}

				return msgServer
			},
			expected:      &types.MsgExecuteJobResponse{},
			expectedError: nil,
		},
		{
			name: "error in scheduling.  returns the error",
			msgServer: func() msgServer {
				schedulerKeeper := mocks.NewKeeper(t)
				mockAccount := authmocks.NewAccountI(t)
				mockAccount.On("GetAddress").Return(sdk.AccAddress("test-address"))

				schedulerKeeper.On("Logger", mock.Anything).Return(log.NewNopLogger())
				schedulerKeeper.On("GetAccount",
					mock.Anything,
					mock.Anything,
				).Return(mockAccount)

				schedulerKeeper.On("ExecuteJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("error-3"))

				msgServer := msgServer{
					schedulerKeeper,
				}

				return msgServer
			},
			expected:      nil,
			expectedError: errors.New("error-3"),
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			msg := types.MsgExecuteJob{
				Creator: "cosmos1l2j8vaykh03zenzytntj3cza6zfxwlj6lgqvd5",
				JobID:   "test_job_1",
				Payload: []byte(``),
			}

			actual, actualErr := tt.msgServer().ExecuteJob(ctx, &msg)

			asserter.Equal(tt.expected, actual)
			asserter.Equal(tt.expectedError, actualErr)
		})
	}
}
