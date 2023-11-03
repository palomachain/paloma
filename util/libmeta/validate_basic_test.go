package libmeta_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
	"github.com/stretchr/testify/require"
)

func Test_GetValidateBasic(t *testing.T) {
	for i, v := range []struct {
		name    string
		creator string
		signers []string
		errors  bool
	}{
		{
			name:    "with missing signers",
			signers: []string{},
			creator: sdk.AccAddress("address-1").String(),
			errors:  true,
		},
		{
			name:    "with one valid signer and creator",
			signers: []string{sdk.AccAddress("address-1").String()},
			creator: sdk.AccAddress("address-1").String(),
			errors:  false,
		},
		{
			name:    "with one valid signer and invalid creator",
			signers: []string{sdk.AccAddress("address-1").String()},
			creator: "foo",
			errors:  true,
		},
		{
			name:    "with multiple valid signers and creator",
			signers: []string{sdk.AccAddress("address-1").String(), sdk.AccAddress("address-2").String(), sdk.AccAddress("address-3").String()},
			creator: sdk.AccAddress("address-1").String(),
			errors:  false,
		},
		{
			name:    "with multiple valid signers and invalid creator",
			signers: []string{sdk.AccAddress("address-1").String(), sdk.AccAddress("address-2").String(), sdk.AccAddress("address-3").String()},
			creator: "foo",
			errors:  true,
		},
		{
			name:    "with faulty signer among multiple valid signers and creator",
			signers: []string{sdk.AccAddress("address-1").String(), "foo", sdk.AccAddress("address-3").String()},
			creator: sdk.AccAddress("address-1").String(),
			errors:  true,
		},
	} {
		t.Run(fmt.Sprintf("%d. %s", i+1, v.name), func(t *testing.T) {
			r := require.New(t)
			msg := &testMsg{md: metadata{signers: v.signers, creator: v.creator}}
			if v.errors {
				r.Error(libmeta.ValidateBasic(msg))
				return
			}

			r.NoError(libmeta.ValidateBasic(msg))
		})
	}
}
