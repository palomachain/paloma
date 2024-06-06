package keeper

import (
	"context"
	"testing"

	"github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
)

func TestGetPigeonRequirements(t *testing.T) {
	testcases := []struct {
		name          string
		setup         func(context.Context, *Keeper, mockedServices) *types.QueryGetPigeonRequirementsResponse
		expectedError error
	}{
		{
			name: "returns default, without any pigeon requirements set",
			setup: func(ctx context.Context, k *Keeper, ms mockedServices) *types.QueryGetPigeonRequirementsResponse {
				// Clear the requirements
				pigeonStore := k.pigeonStore(ctx)
				pigeonStore.Delete(types.PigeonRequirementsKey)

				return &types.QueryGetPigeonRequirementsResponse{
					PigeonRequirements: &types.PigeonRequirements{
						MinVersion: defaultMinimumPigeonVersion,
					},
				}
			},
		},
		{
			name: "returns current, without any scheduled pigeon requirements",
			setup: func(ctx context.Context, k *Keeper, ms mockedServices) *types.QueryGetPigeonRequirementsResponse {
				k.SetPigeonRequirements(ctx, &types.PigeonRequirements{
					MinVersion: "v20.0.0",
				})

				return &types.QueryGetPigeonRequirementsResponse{
					PigeonRequirements: &types.PigeonRequirements{
						MinVersion: "v20.0.0",
					},
				}
			},
		},
		{
			name: "returns default and scheduled, with only scheduled pigeon requirements set",
			setup: func(ctx context.Context, k *Keeper, ms mockedServices) *types.QueryGetPigeonRequirementsResponse {
				// Clear the requirements
				pigeonStore := k.pigeonStore(ctx)
				pigeonStore.Delete(types.PigeonRequirementsKey)

				k.SetScheduledPigeonRequirements(ctx, &types.ScheduledPigeonRequirements{
					Requirements: &types.PigeonRequirements{
						MinVersion: "v20.0.0",
					},
					TargetBlockHeight: 1000,
				})

				return &types.QueryGetPigeonRequirementsResponse{
					PigeonRequirements: &types.PigeonRequirements{
						MinVersion: defaultMinimumPigeonVersion,
					},
					ScheduledPigeonRequirements: &types.ScheduledPigeonRequirements{
						Requirements: &types.PigeonRequirements{
							MinVersion: "v20.0.0",
						},
						TargetBlockHeight: 1000,
					},
				}
			},
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			k, ms, ctx := newValsetKeeper(t)

			expected := tt.setup(ctx, k, ms)

			actual, actualErr := k.GetPigeonRequirements(ctx, &types.QueryGetPigeonRequirementsRequest{})
			asserter.Equal(expected, actual)
			asserter.Equal(tt.expectedError, actualErr)
		})
	}
}
