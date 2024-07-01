package filters

import (
	"testing"

	"github.com/palomachain/paloma/x/consensus/types"
)

func Test_IsNotBlockedByValset(t *testing.T) {
	tests := []struct {
		name                 string
		pendingValsetUpdates []types.QueuedSignedMessageI
		msg                  types.QueuedSignedMessageI
		want                 bool
	}{
		{
			name:                 "Test IsNotBlockedByValset with empty valset updates",
			pendingValsetUpdates: nil,
			msg:                  &types.QueuedSignedMessage{},
			want:                 true,
		},
		{
			name: "Test IsNotBlockedByValset with older message than valset update",
			pendingValsetUpdates: []types.QueuedSignedMessageI{
				&types.QueuedSignedMessage{
					Id: 2,
				},
			},
			msg: &types.QueuedSignedMessage{
				Id: 1,
			},
			want: true,
		},
		{
			name: "Test IsNotBlockedByValset with younger message than valset update",
			pendingValsetUpdates: []types.QueuedSignedMessageI{
				&types.QueuedSignedMessage{
					Id: 1,
				},
			},
			msg: &types.QueuedSignedMessage{
				Id: 2,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotBlockedByValset(tt.pendingValsetUpdates, tt.msg); got != tt.want {
				t.Errorf("IsNotBlockedByValset() = %v, want %v", got, tt.want)
			}
		})
	}
}
