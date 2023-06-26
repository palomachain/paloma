package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckFeeValid(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid input",
			input:    "0.00000001",
			expected: true,
		},
		{
			name:     "invalid input - no leading zero",
			input:    ".00000001",
			expected: false,
		},
		{
			name:     "invalid input - no too many decimal places",
			input:    "0.000000001",
			expected: false,
		},
		{
			name:     "invalid input - empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid input - percentage",
			input:    "0.01%",
			expected: false,
		},
	}

	asserter := assert.New(t)

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual := checkFeeValid(tt.input)
			asserter.Equal(tt.expected, actual)
		})
	}
}
