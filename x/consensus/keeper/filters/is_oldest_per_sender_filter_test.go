package filters

import (
	"testing"

	evmtypes "github.com/palomachain/paloma/x/evm/types"
)

func Test_IsOldestMsgPerSender(t *testing.T) {
	tests := []struct {
		name string
		lut  map[string]struct{}
		msg  *evmtypes.Message
		want bool
	}{
		{
			name: "Test IsOldestMsgPerSender with non SLC",
			lut:  nil,
			msg:  &evmtypes.Message{},
			want: true,
		},
		{
			name: "Test IsOldestMsgPerSender without sender address",
			lut:  nil,
			msg: &evmtypes.Message{
				Action: &evmtypes.Message_SubmitLogicCall{
					SubmitLogicCall: &evmtypes.SubmitLogicCall{
						SenderAddress: nil,
					},
				},
			},
			want: true,
		},
		{
			name: "Test IsOldestMsgPerSender without sender present in LUT",
			lut:  map[string]struct{}{},
			msg: &evmtypes.Message{
				Action: &evmtypes.Message_SubmitLogicCall{
					SubmitLogicCall: &evmtypes.SubmitLogicCall{
						SenderAddress: []byte("sender"),
					},
				},
			},
			want: true,
		},
		{
			name: "Test IsOldestMsgPerSender with sender present in LUT",
			lut:  map[string]struct{}{"sender": {}},
			msg: &evmtypes.Message{
				Action: &evmtypes.Message_SubmitLogicCall{
					SubmitLogicCall: &evmtypes.SubmitLogicCall{
						SenderAddress: []byte("sender"),
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsOldestMsgPerSender(tt.lut, tt.msg); got != tt.want {
				t.Errorf("IsOldestMsgPerSender() = %v, want %v", got, tt.want)
			}
		})
	}
}
