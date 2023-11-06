package libmeta_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
	"github.com/stretchr/testify/require"
)

type metadata struct {
	creator string
	signers []string
}

// GetCreator implements libmeta.Metadata.
func (m metadata) GetCreator() string {
	return m.creator
}

// GetSigners implements libmeta.Metadata.
func (m metadata) GetSigners() []string {
	return m.signers
}

type testMsg struct {
	md metadata
}

func (msg *testMsg) GetMetadata() libmeta.Metadata {
	return msg.md
}

func Test_GetSigners(t *testing.T) {
	for i, v := range []struct {
		name    string
		creator string
		signers []string
		panics  bool
	}{
		{
			name:    "with one valid signer",
			signers: []string{sdk.AccAddress("address-1").String()},
			creator: "",
			panics:  false,
		},
		{
			name:    "with one invalid signer",
			signers: []string{"foo"},
			creator: "",
			panics:  true,
		},
		{
			name:    "with multiple valid signers",
			signers: []string{sdk.AccAddress("address-1").String(), sdk.AccAddress("address-2").String(), sdk.AccAddress("address-3").String()},
			creator: "",
			panics:  false,
		},
		{
			name:    "with an invalid signer among multiple valid signers",
			signers: []string{"foo", sdk.AccAddress("address-2").String(), sdk.AccAddress("address-3").String()},
			creator: "",
			panics:  true,
		},
	} {
		t.Run(fmt.Sprintf("%d. %s", i+1, v.name), func(t *testing.T) {
			r := require.New(t)
			msg := &testMsg{md: metadata{signers: v.signers, creator: v.creator}}
			if v.panics {
				r.Panics(func() { libmeta.GetSigners(msg) }, "should panic")
				return
			}

			expected := make([]sdk.AccAddress, len(v.signers))
			for i, v := range v.signers {
				addr, err := sdk.AccAddressFromBech32(v)
				r.NoError(err)
				expected[i] = addr
			}
			r.Equal(expected, libmeta.GetSigners(msg))
		})
	}
}
