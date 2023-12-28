package types

import (
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// DefaultParamspace defines the default auth module parameter subspace
const (
	// todo: implement oracle constants as params
	DefaultParamspace = ModuleName
)

var (
	// AttestationVotesPowerThreshold threshold of votes power to succeed
	AttestationVotesPowerThreshold = math.NewInt(66)

	// ParamsStoreKeyContractHash stores the contract hash
	ParamsStoreKeyContractHash = []byte("ContractHash")

	// ParamsStoreKeyBridgeContractAddress stores the ethereum address
	ParamsStoreKeyBridgeEthereumAddress = []byte("BridgeEthereumAddress")

	// ParamsStoreKeyBridgeContractChainID stores the bridge chain id
	ParamsStoreKeyBridgeContractChainID = []byte("BridgeChainID")

	// ParamsStoreKeySignedBatchesWindow stores the signed blocks window
	ParamsStoreKeySignedBatchesWindow = []byte("SignedBatchesWindow")

	// ParamsStoreKeySignedClaimsWindow stores the signed blocks window
	ParamsStoreKeyTargetBatchTimeout = []byte("TargetBatchTimeout")

	// ParamsStoreKeySignedClaimsWindow stores the signed blocks window
	ParamsStoreKeyAverageBlockTime = []byte("AverageBlockTime")

	// ParamsStoreKeySignedClaimsWindow stores the signed blocks window
	ParamsStoreKeyAverageEthereumBlockTime = []byte("AverageEthereumBlockTime")

	// ParamsStoreSlashFractionBatch stores the slash fraction Batch
	ParamsStoreSlashFractionBatch = []byte("SlashFractionBatch")

	// ParamStoreSlashFractionBadEthSignature stores the amount by which a validator making a fraudulent eth signature will be slashed
	ParamStoreSlashFractionBadEthSignature = []byte("SlashFractionBadEthSignature")

	// ResetBridgeState boolean indicates the oracle events of the bridge history should be reset
	ParamStoreResetBridgeState = []byte("ResetBridgeState")

	// ResetBridgeHeight stores the nonce after which oracle events should be discarded when resetting the bridge
	ParamStoreResetBridgeNonce = []byte("ResetBridgeNonce")

	// Ensure that params implements the proper interface
	_ paramtypes.ParamSet = &Params{
		ContractSourceHash:           "",
		BridgeEthereumAddress:        "",
		BridgeChainId:                0,
		SignedBatchesWindow:          0,
		TargetBatchTimeout:           0,
		AverageBlockTime:             0,
		AverageEthereumBlockTime:     0,
		SlashFractionBatch:           math.LegacyDec{},
		SlashFractionBadEthSignature: math.LegacyDec{},
	}
)

// ValidateBasic validates genesis state by looping through the params and
// calling their validation functions
func (s GenesisState) ValidateBasic() error {
	if err := s.Params.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "params")
	}
	return nil
}

// DefaultGenesisState returns empty genesis state
// nolint: exhaustruct
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:             DefaultParams(),
		GravityNonces:      GravityNonces{},
		Batches:            []OutgoingTxBatch{},
		BatchConfirms:      []MsgConfirmBatch{},
		Attestations:       []Attestation{},
		Erc20ToDenoms:      []ERC20ToDenom{},
		UnbatchedTransfers: []OutgoingTransferTx{},
	}
}
