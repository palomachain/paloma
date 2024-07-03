package types

import (
	"bytes"
	fmt "fmt"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

func NewParams(contractSourceHash string, bridgeEthereumAddress string, bridgeChainId uint64, signedBatchesWindow uint64, targetBatchTimeout uint64, averageBlockTime uint64, averageEthereumBlockTime uint64, slashFractionBatch math.LegacyDec, slashFractionBadEthSignature math.LegacyDec) Params {
	return Params{
		ContractSourceHash:           contractSourceHash,
		BridgeEthereumAddress:        bridgeEthereumAddress,
		BridgeChainId:                bridgeChainId,
		SignedBatchesWindow:          signedBatchesWindow,
		TargetBatchTimeout:           targetBatchTimeout,
		AverageBlockTime:             averageBlockTime,
		AverageEthereumBlockTime:     averageEthereumBlockTime,
		SlashFractionBatch:           slashFractionBatch,
		SlashFractionBadEthSignature: slashFractionBadEthSignature,
	}
}

// DefaultParams returns a copy of the default params
func DefaultParams() *Params {
	return &Params{
		ContractSourceHash:           "",
		BridgeEthereumAddress:        "0x0000000000000000000000000000000000000000",
		BridgeChainId:                0,
		SignedBatchesWindow:          10000,
		TargetBatchTimeout:           43200000,
		AverageBlockTime:             5000,
		AverageEthereumBlockTime:     15000,
		SlashFractionBatch:           math.LegacyNewDec(1).Quo(math.LegacyNewDec(1000)),
		SlashFractionBadEthSignature: math.LegacyNewDec(1).Quo(math.LegacyNewDec(1000)),
	}
}

// ValidateBasic checks that the parameters have valid values.
func (p Params) ValidateBasic() error {
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
	if err := validateSignedBatchesWindow(p.SignedBatchesWindow); err != nil {
		return sdkerrors.Wrap(err, "signed blocks window batches")
	}
	if err := validateSlashFractionBatch(p.SlashFractionBatch); err != nil {
		return sdkerrors.Wrap(err, "slash fraction batch")
	}
	if err := validateSlashFractionBadEthSignature(p.SlashFractionBadEthSignature); err != nil {
		return sdkerrors.Wrap(err, "slash fraction BadEthSignature")
	}
	return nil
}

// ParamKeyTable for auth module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{
		ContractSourceHash:           "",
		BridgeEthereumAddress:        "",
		BridgeChainId:                0,
		SignedBatchesWindow:          0,
		TargetBatchTimeout:           0,
		AverageBlockTime:             0,
		AverageEthereumBlockTime:     0,
		SlashFractionBatch:           math.LegacyDec{},
		SlashFractionBadEthSignature: math.LegacyDec{},
	})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamsStoreKeyContractHash, &p.ContractSourceHash, validateContractHash),
		paramtypes.NewParamSetPair(ParamsStoreKeyBridgeEthereumAddress, &p.BridgeEthereumAddress, validateBridgeContractAddress),
		paramtypes.NewParamSetPair(ParamsStoreKeyBridgeContractChainID, &p.BridgeChainId, validateBridgeChainID),
		paramtypes.NewParamSetPair(ParamsStoreKeySignedBatchesWindow, &p.SignedBatchesWindow, validateSignedBatchesWindow),
		paramtypes.NewParamSetPair(ParamsStoreKeyTargetBatchTimeout, &p.TargetBatchTimeout, validateTargetBatchTimeout),
		paramtypes.NewParamSetPair(ParamsStoreKeyAverageBlockTime, &p.AverageBlockTime, validateAverageBlockTime),
		paramtypes.NewParamSetPair(ParamsStoreKeyAverageEthereumBlockTime, &p.AverageEthereumBlockTime, validateAverageEthereumBlockTime),
		paramtypes.NewParamSetPair(ParamsStoreSlashFractionBatch, &p.SlashFractionBatch, validateSlashFractionBatch),
		paramtypes.NewParamSetPair(ParamStoreSlashFractionBadEthSignature, &p.SlashFractionBadEthSignature, validateSlashFractionBadEthSignature),
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
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

func validateSignedBatchesWindow(i interface{}) error {
	// TODO: do we want to set some bounds on this value?
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSlashFractionBatch(i interface{}) error {
	// TODO: do we want to set some bounds on this value?
	if _, ok := i.(math.LegacyDec); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSlashFractionBadEthSignature(i interface{}) error {
	// TODO: do we want to set some bounds on this value?
	if _, ok := i.(math.LegacyDec); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
