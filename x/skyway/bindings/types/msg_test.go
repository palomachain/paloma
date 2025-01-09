package types

import (
	"testing"

	"github.com/palomachain/paloma/v2/testutil/common"
	"github.com/stretchr/testify/require"
)

func TestSendTx_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     SendTx
		wantErr bool
	}{
		{
			name: "valid SendTx",
			msg: SendTx{
				RemoteChainDestinationAddress: "0x1234567890abcdef1234567890abcdef12345678",
				Amount:                        "1000000000ugrain",
				ChainReferenceId:              "eth-main",
			},
			wantErr: false,
		},
		{
			name: "invalid address",
			msg: SendTx{
				RemoteChainDestinationAddress: "invalid_address",
				Amount:                        "1000000000ugrain",
				ChainReferenceId:              "eth-main",
			},
			wantErr: true,
		},
		{
			name: "invalid amount",
			msg: SendTx{
				RemoteChainDestinationAddress: "0x1234567890abcdef1234567890abcdef12345678",
				Amount:                        "invalid_amount",
				ChainReferenceId:              "eth-main",
			},
			wantErr: true,
		},
		{
			name: "empty chain reference id",
			msg: SendTx{
				RemoteChainDestinationAddress: "0x1234567890abcdef1234567890abcdef12345678",
				Amount:                        "1000000000ugrain",
				ChainReferenceId:              "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCancelTx_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     CancelTx
		wantErr bool
	}{
		{
			name: "valid CancelTx",
			msg: CancelTx{
				TransactionId: 1,
			},
			wantErr: false,
		},
		{
			name: "invalid transaction id",
			msg: CancelTx{
				TransactionId: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSetErc20ToDenom_ValidateBasic(t *testing.T) {
	common.SetupPalomaPrefixes()
	denom := "factory/paloma167rf0jmkkkqp9d4awa8rxw908muavqgghtw6tn/testcoin"
	tests := []struct {
		name    string
		msg     SetErc20ToDenom
		wantErr bool
	}{
		{
			name: "valid SetErc20ToDenom",
			msg: SetErc20ToDenom{
				Erc20Address:     "0x1234567890abcdef1234567890abcdef12345678",
				TokenDenom:       denom,
				ChainReferenceId: "eth-main",
			},
			wantErr: false,
		},
		{
			name: "invalid address",
			msg: SetErc20ToDenom{
				Erc20Address:     "invalid_address",
				TokenDenom:       denom,
				ChainReferenceId: "eth-main",
			},
			wantErr: true,
		},
		{
			name: "empty chain reference id",
			msg: SetErc20ToDenom{
				Erc20Address:     "0x1234567890abcdef1234567890abcdef12345678",
				TokenDenom:       denom,
				ChainReferenceId: "",
			},
			wantErr: true,
		},
		{
			name: "invalid token denom",
			msg: SetErc20ToDenom{
				Erc20Address:     "0x1234567890abcdef1234567890abcdef12345678",
				TokenDenom:       "invalid_denom",
				ChainReferenceId: "eth-main",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
