package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errorsmod "github.com/cosmos/cosmos-sdk/types/errors"
	gethcommon "github.com/ethereum/go-ethereum/common"
)

const (
	// ZeroAddress is an EthAddress containing the zero ethereum address
	ZeroAddressString = "0x0000000000000000000000000000000000000000"
)

// Regular EthAddress
type EthAddress struct {
	address gethcommon.Address
}

// Returns the contained address as a string
func (ea EthAddress) GetAddress() gethcommon.Address {
	return ea.address
}

// Sets the contained address, performing validation before updating the value
func (ea *EthAddress) SetAddress(address string) error {
	if err := ValidateEthAddress(address); err != nil {
		return err
	}
	ea.address = gethcommon.HexToAddress(address)
	return nil
}

func NewEthAddressFromBytes(address []byte) (*EthAddress, error) {
	if err := ValidateEthAddress(hex.EncodeToString(address)); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid input address")
	}

	addr := EthAddress{gethcommon.BytesToAddress(address)}
	return &addr, nil
}

// Creates a new EthAddress from a string, performing validation and returning any validation errors
func NewEthAddress(address string) (*EthAddress, error) {
	if err := ValidateEthAddress(address); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid input address")
	}

	addr := EthAddress{gethcommon.HexToAddress(address)}
	return &addr, nil
}

// Returns a new EthAddress with 0x0000000000000000000000000000000000000000 as the wrapped address
func ZeroAddress() EthAddress {
	return EthAddress{gethcommon.HexToAddress(ZeroAddressString)}
}

// Validates the input string as an Ethereum Address
// Addresses must not be empty, have 42 character length, start with 0x and have 40 remaining characters in [0-9a-fA-F]
func ValidateEthAddress(address string) error {
	if address == "" {
		return fmt.Errorf("empty")
	}

	// An auditor recommended we should check the error of hex.DecodeString, given that geth's HexToAddress ignores it

	if has0xPrefix(address) {
		address = address[2:]
	}

	if _, err := hex.DecodeString(address); err != nil {
		return fmt.Errorf("invalid hex with error: %s", err)
	}

	if !gethcommon.IsHexAddress(address) {
		return fmt.Errorf("address(%s) doesn't pass format validation", address)
	}

	return nil
}

// Performs validation on the wrapped string
func (ea EthAddress) ValidateBasic() error {
	return ValidateEthAddress(ea.address.Hex())
}

// EthAddrLessThan migrates the Ethereum address less than function
func EthAddrLessThan(e EthAddress, o EthAddress) bool {
	return bytes.Compare([]byte(e.GetAddress().Hex()), []byte(o.GetAddress().Hex())) == -1
}

/////////////////////////
// ERC20Token      //
/////////////////////////

// NewERC20Token returns a new instance of an ERC20
func NewERC20Token(amount uint64, contract string, chainReferenceId string) ERC20Token {
	return NewSDKIntERC20Token(math.NewIntFromUint64(amount), contract, chainReferenceId)
}

// NewSDKIntERC20Token returns a new instance of an ERC20, accepting a sdk.Int
func NewSDKIntERC20Token(amount math.Int, contract string, chainReferenceId string) ERC20Token {
	return ERC20Token{
		Amount:           amount,
		Contract:         contract,
		ChainReferenceId: chainReferenceId,
	}
}

// ToInternal converts an ERC20Token to the internal type InternalERC20Token
func (e ERC20Token) ToInternal() (*InternalERC20Token, error) {
	return NewInternalERC20Token(e.Amount, e.Contract, e.ChainReferenceId)
}

// InternalERC20Token contains validated fields, used for all internal computation
type InternalERC20Token struct {
	Amount           math.Int
	Contract         EthAddress
	ChainReferenceID string
}

// NewInternalERC20Token creates an InternalERC20Token, performing validation and returning any errors
func NewInternalERC20Token(amount math.Int, contract string, chainReferenceId string) (*InternalERC20Token, error) {
	ethAddress, err := NewEthAddress(contract)
	if err != nil { // ethAddress could be nil, must return here
		return nil, sdkerrors.Wrap(err, "invalid contract")
	}
	ret := &InternalERC20Token{
		Amount:           amount,
		Contract:         *ethAddress,
		ChainReferenceID: chainReferenceId,
	}
	if err := ret.ValidateBasic(); err != nil {
		return nil, err
	}

	return ret, nil
}

// ValidateBasic performs validation on all fields of an InternalERC20Token
func (i *InternalERC20Token) ValidateBasic() error {
	if i.Amount.IsNegative() {
		return sdkerrors.Wrap(errorsmod.ErrInvalidCoins, "coins must not be negative")
	}
	err := i.Contract.ValidateBasic()
	if err != nil {
		return sdkerrors.Wrap(err, "invalid contract")
	}
	return nil
}

// ToExternal converts an InternalERC20Token to the external type ERC20Token
func (i *InternalERC20Token) ToExternal() ERC20Token {
	return ERC20Token{
		Amount:           i.Amount,
		Contract:         i.Contract.GetAddress().Hex(),
		ChainReferenceId: i.ChainReferenceID,
	}
}

// ValidateBasic performs stateless validation
func (e *ERC20Token) ValidateBasic() error {
	if err := ValidateEthAddress(e.Contract); err != nil {
		return sdkerrors.Wrap(err, "ethereum address")
	}
	return nil
}

// Add adds one ERC20 to another
func (i *InternalERC20Token) Add(o *InternalERC20Token) (*InternalERC20Token, error) {
	if i.Contract.GetAddress() != o.Contract.GetAddress() {
		return nil, sdkerrors.Wrap(ErrMismatched, "cannot add two different tokens")
	}
	sum := i.Amount.Add(o.Amount) // validation happens in NewInternalERC20Token()
	return NewInternalERC20Token(sum, i.Contract.GetAddress().Hex(), i.ChainReferenceID)
}

func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

func (m ERC20ToDenom) ValidateBasic() error {
	trimDenom := strings.TrimSpace(m.Denom)
	if trimDenom == "" || trimDenom != m.Denom {
		return sdkerrors.Wrap(ErrInvalid, "invalid erc20todenom: denom must be properly formatted")
	}
	trimErc20 := strings.TrimSpace(m.Erc20)
	if trimErc20 == "" || trimErc20 != m.Erc20 {
		return sdkerrors.Wrap(ErrInvalid, "invalid erc20todenom: erc20 must be properly formatted")
	}
	addr, err := NewEthAddress(m.Erc20)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalid, "invalid erc20todenom: erc20 must be a valid ethereum address: %v", err)
	}
	if err = addr.ValidateBasic(); err != nil {
		return sdkerrors.Wrapf(ErrInvalid, "invalid erc20todenom: erc20 address failed validate basic: %v", err)
	}
	if err = sdk.ValidateDenom(m.Denom); err != nil {
		return sdkerrors.Wrapf(ErrInvalid, "invalid erc20todenom: denom is invalid: %v", err)
	}

	return nil
}
