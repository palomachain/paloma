package filters

import (
	"testing"

	"github.com/palomachain/paloma/x/consensus/types"
)

func Test_IsUnprocessed(t *testing.T) {
	tests := []struct {
		name string
		msg  types.QueuedSignedMessageI
		want bool
	}{
		{
			name: "Test IsUnprocessed with nil public access data and nil error data",
			msg: &types.QueuedSignedMessage{
				PublicAccessData: nil,
				ErrorData:        nil,
			},
			want: true,
		},
		{
			name: "Test IsUnprocessed with non-nil public access data and nil error data",
			msg: &types.QueuedSignedMessage{
				PublicAccessData: &types.PublicAccessData{},
				ErrorData:        nil,
			},
			want: false,
		},
		{
			name: "Test IsUnprocessed with nil public access data and non-nil error data",
			msg: &types.QueuedSignedMessage{
				PublicAccessData: nil,
				ErrorData:        &types.ErrorData{},
			},
			want: false,
		},
		{
			name: "Test IsUnprocessed with non-nil public access data and non-nil error data",
			msg: &types.QueuedSignedMessage{
				PublicAccessData: &types.PublicAccessData{},
				ErrorData:        &types.ErrorData{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnprocessed(tt.msg); got != tt.want {
				t.Errorf("IsUnprocessed() = %v, want %v", got, tt.want)
			}
		})
	}
}
