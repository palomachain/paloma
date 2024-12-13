package types_test

import (
	"testing"

	testutilcommmon "github.com/palomachain/paloma/v2/testutil/common"
	"github.com/palomachain/paloma/v2/x/tokenfactory/types"
	"github.com/stretchr/testify/require"
)

func TestDeconstructDenom(t *testing.T) {
	testutilcommmon.SetupPalomaPrefixes()
	for _, tc := range []struct {
		desc             string
		denom            string
		expectedSubdenom string
		err              error
	}{
		{
			desc:  "empty is invalid",
			denom: "",
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "normal",
			denom:            "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/bitcoin",
			expectedSubdenom: "bitcoin",
		},
		{
			desc:             "multiple slashes in subdenom",
			denom:            "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/bitcoin/1",
			expectedSubdenom: "bitcoin/1",
		},
		{
			desc:             "no subdenom",
			denom:            "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/",
			expectedSubdenom: "",
		},
		{
			desc:  "incorrect prefix",
			denom: "ibc/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/bitcoin",
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "subdenom of only slashes",
			denom:            "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/////",
			expectedSubdenom: "////",
		},
		{
			desc:  "too long name",
			denom: "factory/paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm/adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			err:   types.ErrInvalidDenom,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			expectedCreator := "paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm"
			creator, subdenom, err := types.DeconstructDenom(tc.denom)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, expectedCreator, creator)
				require.Equal(t, tc.expectedSubdenom, subdenom)
			}
		})
	}
}

func TestGetTokenDenom(t *testing.T) {
	testutilcommmon.SetupPalomaPrefixes()
	for _, tc := range []struct {
		desc     string
		creator  string
		subdenom string
		valid    bool
	}{
		{
			desc:     "normal",
			creator:  "paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm",
			subdenom: "bitcoin",
			valid:    true,
		},
		{
			desc:     "multiple slashes in subdenom",
			creator:  "paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm",
			subdenom: "bitcoin/1",
			valid:    true,
		},
		{
			desc:     "no subdenom",
			creator:  "paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm",
			subdenom: "",
			valid:    true,
		},
		{
			desc:     "subdenom of only slashes",
			creator:  "paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm",
			subdenom: "/////",
			valid:    true,
		},
		{
			desc:     "too long name",
			creator:  "paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm",
			subdenom: "adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			valid:    false,
		},
		{
			desc:     "subdenom is exactly max length",
			creator:  "paloma1wm4y8yhppxud6j5wvwr7fyynhh09tmv5x5sfzm",
			subdenom: "bitcoinfsadfsdfeadfsafwefsefsefsdfsdafasefsf",
			valid:    true,
		},
		{
			desc:     "creator is exactly max length",
			creator:  "paloma1t7egva48prqmzl59x5ngv4zx0dtrwewcdqdjr8jhgjhgkhjklhkjhkjhgjhgjgjghelu",
			subdenom: "bitcoin",
			valid:    true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := types.GetTokenDenom(tc.creator, tc.subdenom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
