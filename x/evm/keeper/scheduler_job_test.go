package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestAddressToHex(t *testing.T) {
	testcases := []struct {
		name     string
		input    sdk.AccAddress
		expected string
	}{
		{
			name:     "successfully converts address to hex",
			input:    sdk.AccAddress("test-address"),
			expected: "746573742d61646472657373",
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual := addressToHex(tt.input)
			asserter.Equal(tt.expected, actual)
		})
	}
}
