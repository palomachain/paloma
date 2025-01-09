package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateJob_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		job     *Job
		wantErr bool
	}{
		{
			name:    "valid job",
			job:     &Job{JobId: "1", ChainType: "evm", ChainReferenceId: "eth-main", Definition: "{}", Payload: "{}", PayloadModifiable: false, IsMEV: false},
			wantErr: false,
		},
		{
			name:    "nil job",
			job:     nil,
			wantErr: true,
		},
		{
			name:    "empty job id",
			job:     &Job{JobId: "", ChainType: "evm", ChainReferenceId: "eth-main", Definition: "{}", Payload: "{}", PayloadModifiable: false, IsMEV: false},
			wantErr: true,
		},
		{
			name:    "empty chain type",
			job:     &Job{JobId: "1", ChainType: "", ChainReferenceId: "eth-main", Definition: "{}", Payload: "{}", PayloadModifiable: false, IsMEV: false},
			wantErr: true,
		},
		{
			name:    "empty chain reference id",
			job:     &Job{JobId: "1", ChainType: "evm", ChainReferenceId: "", Definition: "{}", Payload: "{}", PayloadModifiable: false, IsMEV: false},
			wantErr: true,
		},
		{
			name:    "empty definition",
			job:     &Job{JobId: "1", ChainType: "evm", ChainReferenceId: "eth-main", Definition: "", Payload: "{}", PayloadModifiable: false, IsMEV: false},
			wantErr: true,
		},
		{
			name:    "empty payload",
			job:     &Job{JobId: "1", ChainType: "evm", ChainReferenceId: "eth-main", Definition: "{}", Payload: "", PayloadModifiable: false, IsMEV: false},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cj := CreateJob{Job: tt.job}
			err := cj.ValidateBasic()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteJob_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		job     ExecuteJob
		wantErr bool
	}{
		{
			name:    "valid execute job",
			job:     ExecuteJob{JobID: "1", Sender: "sender", Payload: []byte("payload")},
			wantErr: false,
		},
		{
			name:    "empty job id",
			job:     ExecuteJob{JobID: "", Sender: "sender", Payload: []byte("payload")},
			wantErr: true,
		},
		{
			name:    "empty sender",
			job:     ExecuteJob{JobID: "1", Sender: "", Payload: []byte("payload")},
			wantErr: true,
		},
		{
			name:    "empty payload",
			job:     ExecuteJob{JobID: "1", Sender: "sender", Payload: []byte("")},
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
