package keeper

import (
	"context"
	"testing"

	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetLatestPublishedSnapshot(t *testing.T) {
	testcases := []struct {
		name          string
		setup         func(context.Context, *Keeper, mockedServices) *types.QueryGetLatestPublishedSnapshotResponse
		request       *types.QueryGetLatestPublishedSnapshotRequest
		expectedError error
	}{
		{
			name: "returns the latest published snapshot",
			setup: func(ctx context.Context, k *Keeper, ms mockedServices) *types.QueryGetLatestPublishedSnapshotResponse {
				ms.StakingKeeper.On("IterateValidators", mock.Anything, mock.Anything).Return(nil)

				snapshot, err := k.createNewSnapshot(ctx)
				require.NoError(t, err)

				err = k.setSnapshotAsCurrent(ctx, snapshot)
				require.NoError(t, err)

				snapshot, err = k.GetCurrentSnapshot(ctx)
				require.NoError(t, err)

				k.SetSnapshotOnChain(ctx, snapshot.Id, "test-chain")

				snapshot, err = k.GetCurrentSnapshot(ctx)
				require.NoError(t, err)

				return &types.QueryGetLatestPublishedSnapshotResponse{
					Snapshot: snapshot,
				}
			},
			request: &types.QueryGetLatestPublishedSnapshotRequest{
				ChainReferenceID: "test-chain",
			},
		},
		{
			name: "returns error when no snapshot published",
			setup: func(ctx context.Context, k *Keeper, ms mockedServices) *types.QueryGetLatestPublishedSnapshotResponse {
				return nil
			},
			request: &types.QueryGetLatestPublishedSnapshotRequest{
				ChainReferenceID: "test-chain",
			},
			expectedError: keeperutil.ErrNotFound.Format(&types.Snapshot{}, []byte{0, 0, 0, 0, 0, 0, 0, 0}),
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			k, ms, ctx := newValsetKeeper(t)

			expected := tt.setup(ctx, k, ms)

			actual, actualErr := k.GetLatestPublishedSnapshot(ctx, tt.request)
			asserter.Equal(expected, actual)
			asserter.Equal(tt.expectedError, actualErr)
		})
	}
}
