package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

// nolint: exhaustruct
func TestGenesisStateValidate(t *testing.T) {
	specs := map[string]struct {
		src    *GenesisState
		expErr bool
	}{
		"default params": {src: DefaultGenesisState(), expErr: false},
		"empty params": {src: &GenesisState{
			Params: &Params{
				ContractSourceHash:           "",
				BridgeEthereumAddress:        "",
				BridgeChainId:                0,
				SignedBatchesWindow:          0,
				TargetBatchTimeout:           0,
				AverageBlockTime:             0,
				AverageEthereumBlockTime:     0,
				SlashFractionBatch:           math.LegacyDec{},
				SlashFractionBadEthSignature: math.LegacyDec{},
			},
			GravityNonces:      GravityNonces{},
			Batches:            []OutgoingTxBatch{},
			BatchConfirms:      []MsgConfirmBatch{},
			Attestations:       []Attestation{},
			Erc20ToDenoms:      []ERC20ToDenom{},
			UnbatchedTransfers: []OutgoingTransferTx{},
		}, expErr: true},
		"invalid params": {src: &GenesisState{
			Params: &Params{
				ContractSourceHash:           "laksdjflasdkfja",
				BridgeEthereumAddress:        "invalid-eth-address",
				BridgeChainId:                3279089,
				SignedBatchesWindow:          0,
				TargetBatchTimeout:           0,
				AverageBlockTime:             0,
				AverageEthereumBlockTime:     0,
				SlashFractionBatch:           math.LegacyDec{},
				SlashFractionBadEthSignature: math.LegacyDec{},
			},
			GravityNonces:      GravityNonces{},
			Batches:            []OutgoingTxBatch{},
			BatchConfirms:      []MsgConfirmBatch{},
			Attestations:       []Attestation{},
			Erc20ToDenoms:      []ERC20ToDenom{},
			UnbatchedTransfers: []OutgoingTransferTx{},
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
