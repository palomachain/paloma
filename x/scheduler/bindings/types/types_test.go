package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJob_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		job     Job
		wantErr bool
	}{
		{
			name: "valid job",
			job: Job{
				JobId:            "job1",
				ChainType:        "evm",
				ChainReferenceId: "eth-main",
				Definition:       "contract ABI",
				Payload:          "data",
			},
			wantErr: false,
		},
		{
			name: "missing job id",
			job: Job{
				ChainType:        "evm",
				ChainReferenceId: "eth-main",
				Definition:       "contract ABI",
				Payload:          "data",
			},
			wantErr: true,
		},
		{
			name: "missing chain type",
			job: Job{
				JobId:            "job1",
				ChainReferenceId: "eth-main",
				Definition:       "contract ABI",
				Payload:          "data",
			},
			wantErr: true,
		},
		{
			name: "missing chain reference id",
			job: Job{
				JobId:      "job1",
				ChainType:  "evm",
				Definition: "contract ABI",
				Payload:    "data",
			},
			wantErr: true,
		},
		{
			name: "missing definition",
			job: Job{
				JobId:            "job1",
				ChainType:        "evm",
				ChainReferenceId: "eth-main",
				Payload:          "data",
			},
			wantErr: true,
		},
		{
			name: "missing payload",
			job: Job{
				JobId:            "job1",
				ChainType:        "evm",
				ChainReferenceId: "eth-main",
				Definition:       "contract ABI",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.ValidateBasic()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
