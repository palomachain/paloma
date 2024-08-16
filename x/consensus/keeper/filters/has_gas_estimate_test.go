package filters

import (
	"testing"

	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/stretchr/testify/require"
)

func Test_HasGasEstimate(t *testing.T) {
	tt := []struct {
		name        string
		gasEstimate uint64
		flagMask    uint32
		want        bool
	}{
		{
			name:        "no estimate needed, no gas estimate",
			gasEstimate: 0,
			flagMask:    0,
			want:        true,
		},
		{
			name:        "no estimate needed, with gas estimate",
			gasEstimate: 100,
			flagMask:    0,
			want:        true,
		},
		{
			name:        "estimate needed, no gas estimate",
			gasEstimate: 0,
			flagMask:    1,
			want:        false,
		},
		{
			name:        "estimate needed, with gas estimate",
			gasEstimate: 100,
			flagMask:    1,
			want:        true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			msg := &types.QueuedSignedMessage{
				GasEstimate: tc.gasEstimate,
				FlagMask:    tc.flagMask,
			}
			res := HasGasEstimate(msg)
			require.Equal(t, tc.want, res)
		})
	}
}
