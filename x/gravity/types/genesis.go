package types

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// DefaultParamspace defines the default auth module parameter subspace
const (
	// todo: implement oracle constants as params
	DefaultParamspace = ModuleName
)

var (
	// AttestationVotesPowerThreshold threshold of votes power to succeed
	AttestationVotesPowerThreshold = sdk.NewInt(66)

	// ParamsStoreKeyGravityID stores the gravity id
	ParamsStoreKeyGravityID = []byte("GravityID")

	// ParamsStoreKeyContractHash stores the contract hash
	ParamsStoreKeyContractHash = []byte("ContractHash")

	// ParamsStoreKeyBridgeContractAddress stores the ethereum address
	ParamsStoreKeyBridgeEthereumAddress = []byte("BridgeEthereumAddress")

	// ParamsStoreKeyBridgeContractChainID stores the bridge chain id
	ParamsStoreKeyBridgeContractChainID = []byte("BridgeChainID")

	// ParamsStoreKeySignedValsetsWindow stores the signed blocks window
	ParamsStoreKeySignedValsetsWindow = []byte("SignedValsetsWindow")

	// ParamsStoreKeySignedBatchesWindow stores the signed blocks window
	ParamsStoreKeySignedBatchesWindow = []byte("SignedBatchesWindow")

	// ParamsStoreKeySignedLogicCallsWindow stores the signed blocks window
	ParamsStoreKeySignedLogicCallsWindow = []byte("SignedLogicCallsWindow")

	// ParamsStoreKeySignedClaimsWindow stores the signed blocks window
	ParamsStoreKeyTargetBatchTimeout = []byte("TargetBatchTimeout")

	// ParamsStoreKeySignedClaimsWindow stores the signed blocks window
	ParamsStoreKeyAverageBlockTime = []byte("AverageBlockTime")

	// ParamsStoreKeySignedClaimsWindow stores the signed blocks window
	ParamsStoreKeyAverageEthereumBlockTime = []byte("AverageEthereumBlockTime")

	// ParamsStoreSlashFractionValset stores the slash fraction valset
	ParamsStoreSlashFractionValset = []byte("SlashFractionValset")

	// ParamsStoreSlashFractionBatch stores the slash fraction Batch
	ParamsStoreSlashFractionBatch = []byte("SlashFractionBatch")

	// ParamStoreUnbondSlashingValsetsWindow stores unbond slashing valset window
	ParamStoreUnbondSlashingValsetsWindow = []byte("UnbondSlashingValsetsWindow")

	// ParamStoreSlashFractionBadEthSignature stores the amount by which a validator making a fraudulent eth signature will be slashed
	ParamStoreSlashFractionBadEthSignature = []byte("SlashFractionBadEthSignature")

	// ValsetRewardAmount the amount of the coin, both denom and amount to issue
	// to a relayer when they relay a valset
	ParamStoreValsetRewardAmount = []byte("ValsetReward")

	// ResetBridgeState boolean indicates the oracle events of the bridge history should be reset
	ParamStoreResetBridgeState = []byte("ResetBridgeState")

	// ResetBridgeHeight stores the nonce after which oracle events should be discarded when resetting the bridge
	ParamStoreResetBridgeNonce = []byte("ResetBridgeNonce")

	// ParamBridgeActive allows governance to temporarily halt the bridge via vote, in this context halting
	// means no more batches will be created and no oracle events executed. Valset creation will continue
	// to be allowed as it must continue to ensure bridge continuity.
	ParamStoreBridgeActive = []byte("BridgeActive")

	// ParamStoreEthereumBlacklist allows storage of blocked Ethereum addresses blocked for use with the bridge
	// this could be for technical reasons (zero address) or non-technical reasons, these apply across all ERC20 tokens
	ParamStoreEthereumBlacklist = []byte("EthereumBlacklist")

	// ParamStoreMinChainFeeBasisPoints allows governance to set the minimum SendToEth `ChainFee` in terms of basis points
	// or hundredths of a percent, e.g. 10% fee = 1000 and 0.02% fee = 2. If this is set > 0 and a MsgSendToEth is
	// submitted with too low of a ChainFee value, it will be rejected in the AnteHandler
	ParamStoreMinChainFeeBasisPoints = []byte("MinChainFeeBasisPoints")

	// Ensure that params implements the proper interface
	_ paramtypes.ParamSet = &Params{
		GravityId:                    "",
		ContractSourceHash:           "",
		BridgeEthereumAddress:        "",
		BridgeChainId:                0,
		SignedValsetsWindow:          0,
		SignedBatchesWindow:          0,
		SignedLogicCallsWindow:       0,
		TargetBatchTimeout:           0,
		AverageBlockTime:             0,
		AverageEthereumBlockTime:     0,
		SlashFractionValset:          sdk.Dec{},
		SlashFractionBatch:           sdk.Dec{},
		SlashFractionLogicCall:       sdk.Dec{},
		UnbondSlashingValsetsWindow:  0,
		SlashFractionBadEthSignature: sdk.Dec{},
		ValsetReward: sdk.Coin{
			Denom:  "",
			Amount: sdk.Int{},
		},
		BridgeActive:           true,
		EthereumBlacklist:      []string{},
		MinChainFeeBasisPoints: 0,
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
		Params:                 DefaultParams(),
		GravityNonces:          GravityNonces{},
		Valsets:                []Valset{},
		ValsetConfirms:         []MsgValsetConfirm{},
		Batches:                []OutgoingTxBatch{},
		BatchConfirms:          []MsgConfirmBatch{},
		LogicCalls:             []OutgoingLogicCall{},
		LogicCallConfirms:      []MsgConfirmLogicCall{},
		Attestations:           []Attestation{},
		DelegateKeys:           []MsgSetOrchestratorAddress{},
		Erc20ToDenoms:          []ERC20ToDenom{},
		UnbatchedTransfers:     []OutgoingTransferTx{},
		PendingIbcAutoForwards: []PendingIbcAutoForward{},
	}
}

// DefaultParams returns a copy of the default params
func DefaultParams() *Params {
	return &Params{
		GravityId:                    "defaultgravityid",
		ContractSourceHash:           "",
		BridgeEthereumAddress:        "0x0000000000000000000000000000000000000000",
		BridgeChainId:                0,
		SignedValsetsWindow:          10000,
		SignedBatchesWindow:          10000,
		SignedLogicCallsWindow:       10000,
		TargetBatchTimeout:           43200000,
		AverageBlockTime:             5000,
		AverageEthereumBlockTime:     15000,
		SlashFractionValset:          sdk.NewDec(1).Quo(sdk.NewDec(1000)),
		SlashFractionBatch:           sdk.NewDec(1).Quo(sdk.NewDec(1000)),
		SlashFractionLogicCall:       sdk.NewDec(1).Quo(sdk.NewDec(1000)),
		UnbondSlashingValsetsWindow:  10000,
		SlashFractionBadEthSignature: sdk.NewDec(1).Quo(sdk.NewDec(1000)),
		ValsetReward:                 sdk.Coin{Denom: "", Amount: sdk.ZeroInt()},
		BridgeActive:                 true,
		EthereumBlacklist:            []string{},
		MinChainFeeBasisPoints:       2,
	}
}

// ValidateBasic checks that the parameters have valid values.
func (p Params) ValidateBasic() error {
	if err := validateGravityID(p.GravityId); err != nil {
		return sdkerrors.Wrap(err, "gravity id")
	}
	if err := validateContractHash(p.ContractSourceHash); err != nil {
		return sdkerrors.Wrap(err, "contract hash")
	}
	if err := validateBridgeContractAddress(p.BridgeEthereumAddress); err != nil {
		return sdkerrors.Wrap(err, "bridge contract address")
	}
	if err := validateBridgeChainID(p.BridgeChainId); err != nil {
		return sdkerrors.Wrap(err, "bridge chain id")
	}
	if err := validateTargetBatchTimeout(p.TargetBatchTimeout); err != nil {
		return sdkerrors.Wrap(err, "Batch timeout")
	}
	if err := validateAverageBlockTime(p.AverageBlockTime); err != nil {
		return sdkerrors.Wrap(err, "Block time")
	}
	if err := validateAverageEthereumBlockTime(p.AverageEthereumBlockTime); err != nil {
		return sdkerrors.Wrap(err, "Ethereum block time")
	}
	if err := validateSignedValsetsWindow(p.SignedValsetsWindow); err != nil {
		return sdkerrors.Wrap(err, "signed blocks window valsets")
	}
	if err := validateSignedBatchesWindow(p.SignedBatchesWindow); err != nil {
		return sdkerrors.Wrap(err, "signed blocks window batches")
	}
	if err := validateSignedLogicCallsWindow(p.SignedLogicCallsWindow); err != nil {
		return sdkerrors.Wrap(err, "signed blocks window logic calls")
	}
	if err := validateSlashFractionValset(p.SlashFractionValset); err != nil {
		return sdkerrors.Wrap(err, "slash fraction valset")
	}
	if err := validateSlashFractionBatch(p.SlashFractionBatch); err != nil {
		return sdkerrors.Wrap(err, "slash fraction batch")
	}
	if err := validateSlashFractionLogicCall(p.SlashFractionLogicCall); err != nil {
		return sdkerrors.Wrap(err, "slash fraction logic call")
	}
	if err := validateSlashFractionBadEthSignature(p.SlashFractionBadEthSignature); err != nil {
		return sdkerrors.Wrap(err, "slash fraction BadEthSignature")
	}
	if err := validateUnbondSlashingValsetsWindow(p.UnbondSlashingValsetsWindow); err != nil {
		return sdkerrors.Wrap(err, "unbond Slashing valset window")
	}
	if err := validateValsetRewardAmount(p.ValsetReward); err != nil {
		return sdkerrors.Wrap(err, "ValsetReward amount")
	}
	if err := validateBridgeActive(p.BridgeActive); err != nil {
		return sdkerrors.Wrap(err, "bridge active parameter")
	}
	if err := validateEthereumBlacklistAddresses(p.EthereumBlacklist); err != nil {
		return sdkerrors.Wrap(err, "ethereum blacklist parameter")
	}
	if err := validateMinChainFeeBasisPoints(p.MinChainFeeBasisPoints); err != nil {
		return sdkerrors.Wrap(err, "min chain fee basis points parameter")
	}
	return nil
}

// ParamKeyTable for auth module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{
		GravityId:                    "",
		ContractSourceHash:           "",
		BridgeEthereumAddress:        "",
		BridgeChainId:                0,
		SignedValsetsWindow:          0,
		SignedBatchesWindow:          0,
		SignedLogicCallsWindow:       0,
		TargetBatchTimeout:           0,
		AverageBlockTime:             0,
		AverageEthereumBlockTime:     0,
		SlashFractionValset:          sdk.Dec{},
		SlashFractionBatch:           sdk.Dec{},
		SlashFractionLogicCall:       sdk.Dec{},
		UnbondSlashingValsetsWindow:  0,
		SlashFractionBadEthSignature: sdk.Dec{},
		ValsetReward:                 sdk.Coin{Denom: "", Amount: sdk.Int{}},
		BridgeActive:                 false,
		EthereumBlacklist:            []string{},
		MinChainFeeBasisPoints:       0,
	})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamsStoreKeyGravityID, &p.GravityId, validateGravityID),
		paramtypes.NewParamSetPair(ParamsStoreKeyContractHash, &p.ContractSourceHash, validateContractHash),
		paramtypes.NewParamSetPair(ParamsStoreKeyBridgeEthereumAddress, &p.BridgeEthereumAddress, validateBridgeContractAddress),
		paramtypes.NewParamSetPair(ParamsStoreKeyBridgeContractChainID, &p.BridgeChainId, validateBridgeChainID),
		paramtypes.NewParamSetPair(ParamsStoreKeySignedValsetsWindow, &p.SignedValsetsWindow, validateSignedValsetsWindow),
		paramtypes.NewParamSetPair(ParamsStoreKeySignedBatchesWindow, &p.SignedBatchesWindow, validateSignedBatchesWindow),
		paramtypes.NewParamSetPair(ParamsStoreKeySignedLogicCallsWindow, &p.SignedLogicCallsWindow, validateSignedLogicCallsWindow),
		paramtypes.NewParamSetPair(ParamsStoreKeyTargetBatchTimeout, &p.TargetBatchTimeout, validateTargetBatchTimeout),
		paramtypes.NewParamSetPair(ParamsStoreKeyAverageBlockTime, &p.AverageBlockTime, validateAverageBlockTime),
		paramtypes.NewParamSetPair(ParamsStoreKeyAverageEthereumBlockTime, &p.AverageEthereumBlockTime, validateAverageEthereumBlockTime),
		paramtypes.NewParamSetPair(ParamsStoreSlashFractionValset, &p.SlashFractionValset, validateSlashFractionValset),
		paramtypes.NewParamSetPair(ParamsStoreSlashFractionBatch, &p.SlashFractionBatch, validateSlashFractionBatch),
		paramtypes.NewParamSetPair(ParamStoreUnbondSlashingValsetsWindow, &p.UnbondSlashingValsetsWindow, validateUnbondSlashingValsetsWindow),
		paramtypes.NewParamSetPair(ParamStoreSlashFractionBadEthSignature, &p.SlashFractionBadEthSignature, validateSlashFractionBadEthSignature),
		paramtypes.NewParamSetPair(ParamStoreValsetRewardAmount, &p.ValsetReward, validateValsetRewardAmount),
		paramtypes.NewParamSetPair(ParamStoreBridgeActive, &p.BridgeActive, validateBridgeActive),
		paramtypes.NewParamSetPair(ParamStoreEthereumBlacklist, &p.EthereumBlacklist, validateEthereumBlacklistAddresses),
		paramtypes.NewParamSetPair(ParamStoreMinChainFeeBasisPoints, &p.MinChainFeeBasisPoints, validateMinChainFeeBasisPoints),
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

func validateGravityID(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if _, err := strToFixByteArray(v); err != nil {
		return err
	}
	return nil
}

func validateContractHash(i interface{}) error {
	// TODO: should we validate that the input here is a properly formatted
	// SHA256 (or other) hash?
	if _, ok := i.(string); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateBridgeChainID(i interface{}) error {
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateTargetBatchTimeout(i interface{}) error {
	val, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	} else if val < 60000 {
		return fmt.Errorf("invalid target batch timeout, less than 60 seconds is too short")
	}
	return nil
}

func validateAverageBlockTime(i interface{}) error {
	val, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	} else if val < 100 {
		return fmt.Errorf("invalid average Cosmos block time, too short for latency limitations")
	}
	return nil
}

func validateAverageEthereumBlockTime(i interface{}) error {
	val, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	} else if val < 100 {
		return fmt.Errorf("invalid average Ethereum block time, too short for latency limitations")
	}
	return nil
}

func validateBridgeContractAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := ValidateEthAddress(v); err != nil {
		// TODO: ensure that empty addresses are valid in params
		if !strings.Contains(err.Error(), "empty") {
			return err
		}
	}
	return nil
}

func validateSignedValsetsWindow(i interface{}) error {
	// TODO: do we want to set some bounds on this value?
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateUnbondSlashingValsetsWindow(i interface{}) error {
	// TODO: do we want to set some bounds on this value?
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSlashFractionValset(i interface{}) error {
	// TODO: do we want to set some bounds on this value?
	if _, ok := i.(sdk.Dec); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSignedBatchesWindow(i interface{}) error {
	// TODO: do we want to set some bounds on this value?
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSignedLogicCallsWindow(i interface{}) error {
	// TODO: do we want to set some bounds on this value?
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSlashFractionBatch(i interface{}) error {
	// TODO: do we want to set some bounds on this value?
	if _, ok := i.(sdk.Dec); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSlashFractionLogicCall(i interface{}) error {
	// TODO: do we want to set some bounds on this value?
	if _, ok := i.(sdk.Dec); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSlashFractionBadEthSignature(i interface{}) error {
	// TODO: do we want to set some bounds on this value?
	if _, ok := i.(sdk.Dec); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateValsetRewardAmount(i interface{}) error {
	if _, ok := i.(sdk.Coin); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateBridgeActive(i interface{}) error {
	if _, ok := i.(bool); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateEthereumBlacklistAddresses(i interface{}) error {
	strArr, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for index, value := range strArr {
		if err := ValidateEthAddress(value); err != nil {

			if !strings.Contains(err.Error(), "empty, index is"+strconv.Itoa(index)) {
				return err
			}
		}
	}
	return nil
}

func validateMinChainFeeBasisPoints(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v >= 10000 {
		return fmt.Errorf("MinChainFeeBasisPoints is set to 10000 or more, this is an unreasonable fee amount")
	}
	return nil
}

func strToFixByteArray(s string) ([32]byte, error) {
	var out [32]byte
	if len([]byte(s)) > 32 {
		return out, fmt.Errorf("string too long")
	}
	copy(out[:], s)
	return out, nil
}
