package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisStateValidate(t *testing.T) {
	var nilByteSlice []byte
	specs := map[string]struct {
		src    *GenesisState
		expErr bool
	}{
		"default params": {src: DefaultGenesisState(), expErr: false},
		"empty params":   {src: &GenesisState{Params: &Params{}}, expErr: true},
		"invalid params": {src: &GenesisState{
			Params: &Params{
				GravityId:             "foo",
				ContractSourceHash:    "laksdjflasdkfja",
				BridgeEthereumAddress: "invalid-eth-address",
				BridgeChainId:         3279089,
			},
		}, expErr: true},
		"valid delegate": {src: &GenesisState{
			Params: DefaultParams(),
			DelegateKeys: []*MsgDelegateKeys{
				{
					ValidatorAddress:    "cosmosvaloper13yfm8as7y0mzsxqkfmk5jvgm45aez0u24jk95z",
					OrchestratorAddress: "cosmos1h706wwrghfpydyh735aet8aluhf95dqj0psgyf",
					EthereumAddress:     "0xFDb0aaBD40774BBF3068Bf29E8b0a6C88BE26F83",
					EthSignature:        []byte("0xa2f643b5b919050eb4e200bae900d16fd28adca4d727fb7cc1c68f2517e601d0355340ee0913b1e3b5a5837fbd795857a004e0333913cfb7c59e159ff02115b01c"),
				},
			},
		}, expErr: false},
		"valid delegate with placeholder signature": {src: &GenesisState{
			Params: DefaultParams(),
			DelegateKeys: []*MsgDelegateKeys{
				{
					ValidatorAddress:    "cosmosvaloper13yfm8as7y0mzsxqkfmk5jvgm45aez0u24jk95z",
					OrchestratorAddress: "cosmos1h706wwrghfpydyh735aet8aluhf95dqj0psgyf",
					EthereumAddress:     "0xFDb0aaBD40774BBF3068Bf29E8b0a6C88BE26F83",
					EthSignature:        []byte("unused"), // this will marshal into "dW51c2Vk" as []byte will be encoded as base64
				},
			},
		}, expErr: false},
		"valid delegate with nil signature": {src: &GenesisState{
			Params: DefaultParams(),
			DelegateKeys: []*MsgDelegateKeys{
				{
					ValidatorAddress:    "cosmosvaloper13yfm8as7y0mzsxqkfmk5jvgm45aez0u24jk95z",
					OrchestratorAddress: "cosmos1h706wwrghfpydyh735aet8aluhf95dqj0psgyf",
					EthereumAddress:     "0xFDb0aaBD40774BBF3068Bf29E8b0a6C88BE26F83",
					EthSignature:        nilByteSlice,
				},
			},
		}, expErr: true},
		"delegate with bad validator address": {src: &GenesisState{
			Params: DefaultParams(),
			DelegateKeys: []*MsgDelegateKeys{
				{
					ValidatorAddress:    "cosmosvaloper1wrong",
					OrchestratorAddress: "cosmos1h706wwrghfpydyh735aet8aluhf95dqj0psgyf",
					EthereumAddress:     "0xFDb0aaBD40774BBF3068Bf29E8b0a6C88BE26F83",
					EthSignature:        []byte("0xa2f643b5b919050eb4e200bae900d16fd28adca4d727fb7cc1c68f2517e601d0355340ee0913b1e3b5a5837fbd795857a004e0333913cfb7c59e159ff02115b01c"),
				},
			},
		}, expErr: true},
		"delegate with bad orchestrator address": {src: &GenesisState{
			Params: DefaultParams(),
			DelegateKeys: []*MsgDelegateKeys{
				{
					ValidatorAddress:    "cosmosvaloper13yfm8as7y0mzsxqkfmk5jvgm45aez0u24jk95z",
					OrchestratorAddress: "cosmos1wrong",
					EthereumAddress:     "0xFDb0aaBD40774BBF3068Bf29E8b0a6C88BE26F83",
					EthSignature:        []byte("0xa2f643b5b919050eb4e200bae900d16fd28adca4d727fb7cc1c68f2517e601d0355340ee0913b1e3b5a5837fbd795857a004e0333913cfb7c59e159ff02115b01c"),
				},
			},
		}, expErr: true},
		"delegate with bad eth address": {src: &GenesisState{
			Params: DefaultParams(),
			DelegateKeys: []*MsgDelegateKeys{
				{
					ValidatorAddress:    "cosmosvaloper13yfm8as7y0mzsxqkfmk5jvgm45aez0u24jk95z",
					OrchestratorAddress: "cosmos1h706wwrghfpydyh735aet8aluhf95dqj0psgyf",
					EthereumAddress:     "0xdeadbeef",
					EthSignature:        []byte("0xa2f643b5b919050eb4e200bae900d16fd28adca4d727fb7cc1c68f2517e601d0355340ee0913b1e3b5a5837fbd795857a004e0333913cfb7c59e159ff02115b01c"),
				},
			},
		}, expErr: true},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestStringToByteArray(t *testing.T) {
	specs := map[string]struct {
		testString string
		expErr     bool
	}{
		"16 bytes": {"lakjsdflaksdjfds", false},
		"32 bytes": {"lakjsdflaksdjfdslakjsdflaksdjfds", false},
		"33 bytes": {"€€€€€€€€€€€", true},
	}

	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			_, err := strToFixByteArray(spec.testString)
			if spec.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
