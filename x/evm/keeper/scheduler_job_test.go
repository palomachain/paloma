package keeper

import (
	"testing"

	"github.com/VolumeFi/whoops"
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

func TestZeroPadBytes(t *testing.T) {
	testcases := []struct {
		name           string
		inputByteArray []byte
		inputSize      int
		expected       []byte
		expectedErr    error
	}{
		{
			name:           "when array length equals size",
			inputByteArray: []byte{byte(1), byte(1), byte(1), byte(1), byte(1)},
			inputSize:      5,
			expected:       []byte{byte(1), byte(1), byte(1), byte(1), byte(1)},
		},
		{
			name:           "when array length is less than size",
			inputByteArray: []byte{byte(1), byte(1), byte(1), byte(1), byte(1)},
			inputSize:      7,
			expected:       []byte{byte(0), byte(0), byte(1), byte(1), byte(1), byte(1), byte(1)},
		},
		{
			name:           "when array length is greater than size",
			inputByteArray: []byte{byte(1), byte(1), byte(1), byte(1), byte(1)},
			inputSize:      4,
			expectedErr:    whoops.String("Can not zero pad byte array of size 5 to 4"),
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual, actualErr := zeroPadBytes(tt.inputByteArray, tt.inputSize)
			asserter.Equal(tt.expected, actual)
			asserter.Equal(tt.expectedErr, actualErr)
		})
	}
}
