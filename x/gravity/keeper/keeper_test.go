package keeper

import (
	"testing"

	"cosmossdk.io/math"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
	vtypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint: exhaustruct
func TestPrefixRange(t *testing.T) {
	cases := map[string]struct {
		src      []byte
		expStart []byte
		expEnd   []byte
		expError bool
	}{
		"normal":              {src: []byte{1, 3, 4}, expStart: []byte{1, 3, 4}, expEnd: []byte{1, 3, 5}},
		"normal short":        {src: []byte{79}, expStart: []byte{79}, expEnd: []byte{80}},
		"empty case":          {src: []byte{}},
		"roll-over example 1": {src: []byte{17, 28, 255}, expStart: []byte{17, 28, 255}, expEnd: []byte{17, 29, 0}},
		"roll-over example 2": {
			src:      []byte{15, 42, 255, 255},
			expStart: []byte{15, 42, 255, 255}, expEnd: []byte{15, 43, 0, 0},
		},
		"pathological roll-over": {src: []byte{255, 255, 255, 255}, expStart: []byte{255, 255, 255, 255}},
		"nil prohibited":         {expError: true},
	}

	for testName, tc := range cases {
		tc := tc
		t.Run(testName, func(t *testing.T) {
			start, end, err := prefixRange(tc.src)
			if tc.expError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expStart, start)
				assert.Equal(t, tc.expEnd, end)
			}
		})
	}
}

// nolint: exhaustruct
func TestAttestationIterator(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() {
		sdk.UnwrapSDKContext(input.Context).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	ctx := input.Context
	// add some attestations to the store

	claim1 := &types.MsgSendToPalomaClaim{
		EventNonce:     1,
		TokenContract:  TokenContractAddrs[0],
		Amount:         math.NewInt(100),
		EthereumSender: EthAddrs[0].String(),
		PalomaReceiver: AccAddrs[0].String(),
		Orchestrator:   AccAddrs[0].String(),
		Metadata: vtypes.MsgMetadata{
			Creator: AccAddrs[0].String(),
			Signers: []string{AccAddrs[0].String()},
		},
	}
	ne1, err := codecTypes.NewAnyWithValue(claim1)
	require.NoError(t, err)
	att1 := &types.Attestation{
		Claim:    ne1,
		Observed: true,
		Votes:    []string{ValAddrs[0].String()},
	}

	claim2 := &types.MsgSendToPalomaClaim{
		EventNonce:     2,
		TokenContract:  TokenContractAddrs[0],
		Amount:         math.NewInt(100),
		EthereumSender: EthAddrs[0].String(),
		PalomaReceiver: AccAddrs[0].String(),
		Orchestrator:   AccAddrs[0].String(),
		Metadata: vtypes.MsgMetadata{
			Creator: AccAddrs[0].String(),
			Signers: []string{AccAddrs[0].String()},
		},
	}
	ne2, err := codecTypes.NewAnyWithValue(claim2)
	require.NoError(t, err)
	att2 := &types.Attestation{
		Claim:    ne2,
		Observed: true,
		Votes:    []string{ValAddrs[0].String()},
	}

	hash1, err := claim1.ClaimHash()
	require.NoError(t, err)
	hash2, err := claim2.ClaimHash()
	require.NoError(t, err)

	input.GravityKeeper.SetAttestation(ctx, claim1.EventNonce, hash1, att1)
	err = input.GravityKeeper.setLastObservedEventNonce(ctx, claim1.EventNonce)
	require.NoError(t, err)
	input.GravityKeeper.SetAttestation(ctx, claim2.EventNonce, hash2, att2)
	err = input.GravityKeeper.setLastObservedEventNonce(ctx, claim2.EventNonce)
	require.NoError(t, err)

	atts := []types.Attestation{}
	err = input.GravityKeeper.IterateAttestations(ctx, false, func(_ []byte, att types.Attestation) bool {
		atts = append(atts, att)
		return false
	})
	require.NoError(t, err)

	require.Len(t, atts, 2)
}
