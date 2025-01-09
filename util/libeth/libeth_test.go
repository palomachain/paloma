package libeth

import (
	"testing"
)

func TestValidateEthAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "Valid address with 0x prefix",
			address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			wantErr: false,
		},
		{
			name:    "Valid address without 0x prefix",
			address: "742d35Cc6634C0532925a3b844Bc454e4438f44e",
			wantErr: false,
		},
		{
			name:    "Empty address",
			address: "",
			wantErr: true,
		},
		{
			name:    "Invalid hex characters",
			address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44g",
			wantErr: true,
		},
		{
			name:    "Too short address",
			address: "0x742d35",
			wantErr: true,
		},
		{
			name:    "Too long address",
			address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e11",
			wantErr: true,
		},
		{
			name:    "Invalid prefix",
			address: "1x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			wantErr: true,
		},
		{
			name:    "Only 0x prefix",
			address: "0x",
			wantErr: true,
		},
		{
			name:    "Case-insensitive 0X prefix",
			address: "0X742d35Cc6634C0532925a3b844Bc454e4438f44e",
			wantErr: false,
		},
		{
			name:    "Mixed case address",
			address: "0x1DeaDbeeF634C0532925a3b844Bc454e4438f44e",
			wantErr: false,
		},
		{
			name:    "All lowercase address",
			address: "0x1deadbeef634c0532925a3b844bc454e4438f44e",
			wantErr: false,
		},
		{
			name:    "All uppercase address",
			address: "0x1DEADBEEF634C0532925A3B844BC454E4438F44E",
			wantErr: false,
		},
		{
			name:    "Special characters in address",
			address: "0x742d35Cc6634C0532925a3b844Bc454e4438f4$e",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEthAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEthAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
