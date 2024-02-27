package keeper_test

import (
	"context"
	gocontext "context"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/x/gravity/keeper"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/stretchr/testify/require"
)

// nolint: exhaustruct
func TestQueryGetAttestations(t *testing.T) {
	input := keeper.CreateTestEnv(t)
	encCfg := app.MakeEncodingConfig()
	k := input.GravityKeeper
	ctx := input.Context
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Some query functions use additional logic to determine if they should look up values using the v1 key, or the new
	// hashed bytes keys used post-Mercury, so we must set the block height high enough here for the correct data to be found
	ctx = sdkCtx.WithBlockHeight(int64(keeper.MERCURY_UPGRADE_HEIGHT) + sdkCtx.BlockHeight())

	queryHelper := baseapp.NewQueryServerTestHelper(sdkCtx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, k)
	queryClient := types.NewQueryClient(queryHelper)

	numAttestations := 10
	createAttestations(t, k, ctx, numAttestations)

	testCases := []struct {
		name      string
		req       *types.QueryAttestationsRequest
		numResult int
		nonces    []uint64
		expectErr bool
	}{
		{
			name:      "no params (all attestations ascending)",
			req:       &types.QueryAttestationsRequest{},
			numResult: numAttestations,
			nonces:    []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			expectErr: false,
		},
		{
			name: "all attestations descending",
			req: &types.QueryAttestationsRequest{
				OrderBy: "desc",
			},
			numResult: numAttestations,
			nonces:    []uint64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			expectErr: false,
		},
		{
			name: "all attestations descending",
			req: &types.QueryAttestationsRequest{
				OrderBy: "desc",
			},
			numResult: numAttestations,
			nonces:    []uint64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			expectErr: false,
		},
		{
			name: "filter by height and limit",
			req: &types.QueryAttestationsRequest{
				Height: 1,
				Limit:  5,
			},
			numResult: 5,
			nonces:    []uint64{1, 2, 3, 4, 5},
			expectErr: false,
		},
		{
			name: "filter by nonce and limit",
			req: &types.QueryAttestationsRequest{
				Nonce: 7,
				Limit: 5,
			},
			numResult: 1,
			nonces:    []uint64{7},
			expectErr: false,
		},
		{
			name: "filter by missing nonce",
			req: &types.QueryAttestationsRequest{
				Nonce: 100000,
				Limit: 5,
			},
			numResult: 0,
			nonces:    []uint64{},
			expectErr: false,
		},
		{
			name: "filter by invalid claim type",
			req: &types.QueryAttestationsRequest{
				ClaimType: "foo",
				Limit:     5,
			},
			numResult: 0,
			nonces:    []uint64{},
			expectErr: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			result, err := queryClient.GetAttestations(gocontext.Background(), tc.req)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Lenf(t, result.Attestations, tc.numResult, "unexpected number of results; tc #%d", i)

				nonces := make([]uint64, len(result.Attestations))
				for i, att := range result.Attestations {
					claim, err := k.UnpackAttestationClaim(&att)
					require.NoError(t, err)
					nonces[i] = claim.GetEventNonce()
				}
				require.Equal(t, tc.nonces, nonces)
			}
		})
	}
}

func createAttestations(t *testing.T, k keeper.Keeper, ctx context.Context, length int) {
	t.Helper()

	for i := 0; i < length; i++ {
		nonce := uint64(1 + i)
		msg := types.MsgSendToPalomaClaim{
			EventNonce:     nonce,
			EthBlockHeight: 1,
			TokenContract:  "0x00000000000000000001",
			Amount:         math.NewInt(10000000000 + int64(i)),
			EthereumSender: "0x00000000000000000002",
			PalomaReceiver: "0x00000000000000000003",
			Orchestrator:   "0x00000000000000000004",
		}

		any, err := codectypes.NewAnyWithValue(&msg)
		require.NoError(t, err)
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		att := &types.Attestation{
			Observed: false,
			Votes:    []string{},
			Height:   uint64(sdkCtx.BlockHeight()),
			Claim:    any,
		}

		hash, err := msg.ClaimHash()
		require.NoError(t, err)

		k.SetAttestation(ctx, nonce, hash, att)
	}
}
