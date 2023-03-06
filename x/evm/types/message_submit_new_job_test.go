package types

import (
	"math/big"
	"strings"
	"testing"

	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/palomachain/paloma/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgSubmitNewJob_ValidateBasic(t *testing.T) {
	for _, tt := range []struct {
		name   string
		msg    *MsgSubmitNewJob
		expErr bool
	}{
		{
			name:   "empty msg returns an error",
			expErr: true,
		},
		{
			name: "invalid creator returns an error",
			msg: &MsgSubmitNewJob{
				Creator: "invalid",
			},
			expErr: true,
		},
		{
			name: "invalid smart contract address returns an error",
			msg: &MsgSubmitNewJob{
				Creator:                 sdk.AccAddress("bla").String(),
				HexSmartContractAddress: "oh no this is invalid",
			},
			expErr: true,
		},
		{
			name: "on invalid JSON ABI it returns an error",
			msg: &MsgSubmitNewJob{
				Creator:                 sdk.AccAddress("bla").String(),
				HexSmartContractAddress: common.BytesToAddress([]byte("abc")).String(),
				Abi:                     "oh no this is invalid",
			},
			expErr: true,
		},
		{
			name: "on valid ABI but method is invalid it returns an error",
			msg: &MsgSubmitNewJob{
				Creator:                 sdk.AccAddress("bla").String(),
				HexSmartContractAddress: common.BytesToAddress([]byte("abc")).String(),
				Abi:                     sample.SimpleABI,
				Method:                  "invalid",
			},
			expErr: true,
		},
		{
			name: "on invalid hex payload it returns an error",
			msg: &MsgSubmitNewJob{
				Creator:                 sdk.AccAddress("bla").String(),
				HexSmartContractAddress: common.BytesToAddress([]byte("abc")).String(),
				Abi:                     sample.SimpleABI,
				Method:                  "store",
				HexPayload:              "invalid",
			},
			expErr: true,
		},
		{
			name: "on happy path there are no errors",
			msg: &MsgSubmitNewJob{
				Creator:                 sdk.AccAddress("bla").String(),
				HexSmartContractAddress: common.BytesToAddress([]byte("abc")).String(),
				Abi:                     sample.SimpleABI,
				Method:                  "store",
				HexPayload: func() string {
					evm := whoops.Must(abi.JSON(strings.NewReader(sample.SimpleABI)))
					bz := whoops.Must(evm.Pack("store", big.NewInt(1337)))
					return common.Bytes2Hex(bz)
				}(),
			},
			expErr: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
