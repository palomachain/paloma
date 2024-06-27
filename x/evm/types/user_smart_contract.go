package types

import (
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func (m *UserSmartContract) Validate() error {
	if m.Title == "" {
		return ErrInvalid.Wrap("smart contract title can't be empty")
	}

	if _, err := abi.JSON(strings.NewReader(m.AbiJson)); err != nil {
		return ErrInvalid.Wrap("invalid ABI")
	}

	if m.Bytecode == "" {
		return ErrInvalid.Wrap("smart contract bytecode can't be empty")
	}

	// Validate bytecode is hex string
	if err := validateHex(m.Bytecode); err != nil {
		return ErrInvalid.Wrap("invalid smart contract bytecode")
	}

	// ConstructorInput may be empty
	if m.ConstructorInput != "" {
		if err := validateHex(m.ConstructorInput); err != nil {
			return ErrInvalid.Wrap("invalid smart contract constructor input")
		}
	}

	return nil
}

func validateHex(s string) error {
	if strings.HasPrefix(s, "0x") {
		_, err := hex.DecodeString(s[2:])
		return err
	}

	_, err := hex.DecodeString(s)
	return err
}
