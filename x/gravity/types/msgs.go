package types

import (
	"encoding/hex"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errorsmod "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/palomachain/paloma/util/libmeta"
)

// nolint: exhaustruct
var (
	_ sdk.Msg = &MsgSendToEth{}
	_ sdk.Msg = &MsgCancelSendToEth{}
	_ sdk.Msg = &MsgConfirmBatch{}
	_ sdk.Msg = &MsgSendToPalomaClaim{}
	_ sdk.Msg = &MsgBatchSendToEthClaim{}
	_ sdk.Msg = &MsgSubmitBadSignatureEvidence{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgSendToEth returns a new msgSendToEth
func NewMsgSendToEth(sender sdk.AccAddress, destAddress EthAddress, send sdk.Coin, chainReferenceID string) *MsgSendToEth {
	return &MsgSendToEth{
		Sender:           sender.String(),
		EthDest:          destAddress.GetAddress().Hex(),
		Amount:           send,
		ChainReferenceId: chainReferenceID,
	}
}

// Route should return the name of the module
func (msg MsgSendToEth) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSendToEth) Type() string { return "send_to_eth" }

// ValidateBasic runs stateless checks on the message
// Checks if the Eth address is valid
func (msg MsgSendToEth) ValidateBasic() error {
	if err := libmeta.ValidateBasic(&msg); err != nil {
		return err
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return sdkerrors.Wrap(errorsmod.ErrInvalidCoins, "amount")
	}

	if err := ValidateEthAddress(msg.EthDest); err != nil {
		return sdkerrors.Wrap(err, "ethereum address")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSendToEth) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSendToEth) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&msg)
}

// Route should return the name of the module
func (msg MsgConfirmBatch) Route() string { return RouterKey }

// Type should return the action
func (msg MsgConfirmBatch) Type() string { return "confirm_batch" }

// ValidateBasic performs stateless checks
func (msg MsgConfirmBatch) ValidateBasic() error {
	if err := libmeta.ValidateBasic(&msg); err != nil {
		return err
	}

	if err := ValidateEthAddress(msg.EthSigner); err != nil {
		return sdkerrors.Wrap(err, "eth signer")
	}
	if err := ValidateEthAddress(msg.TokenContract); err != nil {
		return sdkerrors.Wrap(err, "token contract")
	}
	_, err := hex.DecodeString(msg.Signature)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidClaim, "could not decode hex string \"%s\": %v", msg.Signature, err)
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgConfirmBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgConfirmBatch) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&msg)
}

// EthereumClaim represents a claim on ethereum state
type EthereumClaim interface {
	// All Ethereum claims that we relay from the Gravity contract and into the module
	// have a nonce that is strictly increasing and unique, since this nonce is
	// issued by the Ethereum contract it is immutable and must be agreed on by all validators
	// any disagreement on what claim goes to what nonce means someone is lying.
	GetEventNonce() uint64
	// The block height that the claimed event occurred on. This EventNonce provides sufficient
	// ordering for the execution of all claims. The block height is used only for batchTimeouts
	// when we go to create a new batch we set the timeout some number of batches out from the last
	// known height plus projected block progress since then.
	GetEthBlockHeight() uint64
	// the delegate address of the claimer, for MsgDepositClaim and MsgWithdrawClaim
	// this is sent in as the sdk.AccAddress of the delegated key. it is up to the user
	// to disambiguate this into a sdk.ValAddress
	GetClaimer() sdk.AccAddress
	// Which type of claim this is
	GetType() ClaimType
	ValidateBasic() error
	// The claim hash of this claim. This is used to store these claims and also used to check if two different
	// validators claims agree. Therefore it's extremely important that this include all elements of the claim
	// with the exception of the orchestrator who sent it in, which will be used as a different part of the index
	ClaimHash() ([]byte, error)
	// Sets the orchestrator value on the claim
	SetOrchestrator(sdk.AccAddress)
}

// nolint: exhaustruct
var (
	_ EthereumClaim = &MsgSendToPalomaClaim{}
	_ EthereumClaim = &MsgBatchSendToEthClaim{}
)

func (msg *MsgSendToPalomaClaim) SetOrchestrator(orchestrator sdk.AccAddress) {
	msg.Orchestrator = orchestrator.String()
}

// GetType returns the type of the claim
func (msg *MsgSendToPalomaClaim) GetType() ClaimType {
	return CLAIM_TYPE_SEND_TO_PALOMA
}

// ValidateBasic performs stateless checks
func (msg *MsgSendToPalomaClaim) ValidateBasic() error {
	if err := libmeta.ValidateBasic(msg); err != nil {
		return err
	}
	if err := ValidateEthAddress(msg.EthereumSender); err != nil {
		return sdkerrors.Wrap(err, "eth sender")
	}
	if err := ValidateEthAddress(msg.TokenContract); err != nil {
		return sdkerrors.Wrap(err, "erc20 token")
	}
	// note the destination address is intentionally not validated here, since
	// MsgSendToEth has it's destination as a string many invalid inputs are possible
	// the orchestrator will convert these invalid deposits to simply the string invalid'
	// this is done because the oracle requires an event be processed on Cosmos for each event
	// nonce on the Ethereum side, otherwise (A) the oracle will never proceed and (B) the funds
	// sent with the invalid deposit will forever be lost, with no representation minted anywhere
	// on cosmos. The attestation handler deals with this by managing invalid deposits and placing
	// them into the community pool
	if msg.EventNonce == 0 {
		return fmt.Errorf("nonce == 0")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSendToPalomaClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgSendToPalomaClaim) GetClaimer() sdk.AccAddress {
	err := msg.ValidateBasic()
	if err != nil {
		panic("MsgSendToPalomaClaim failed ValidateBasic! Should have been handled earlier")
	}

	val, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}

	return val
}

// GetSigners defines whose signature is required
func (msg MsgSendToPalomaClaim) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&msg)
}

// Type should return the action
func (msg MsgSendToPalomaClaim) Type() string { return "send_to_paloma_claim" }

// Route should return the name of the module
func (msg MsgSendToPalomaClaim) Route() string { return RouterKey }

const (
	TypeMsgSendToPalomaClaim = "send_to_paloma_claim"
)

// Hash implements BridgeDeposit.Hash
// modify this with care as it is security sensitive. If an element of the claim is not in this hash a single hostile validator
// could engineer a hash collision and execute a version of the claim with any unhashed data changed to benefit them.
// note that the Orchestrator is the only field excluded from this hash, this is because that value is used higher up in the store
// structure for who has made what claim and is verified by the msg ante-handler for signatures
func (msg *MsgSendToPalomaClaim) ClaimHash() ([]byte, error) {
	path := fmt.Sprintf("%d/%d/%s/%s/%s/%s", msg.EventNonce, msg.EthBlockHeight, msg.TokenContract, msg.Amount.String(), msg.EthereumSender, msg.PalomaReceiver)
	return tmhash.Sum([]byte(path)), nil
}

func (msg *MsgBatchSendToEthClaim) SetOrchestrator(orchestrator sdk.AccAddress) {
	msg.Orchestrator = orchestrator.String()
}

// GetType returns the claim type
func (msg *MsgBatchSendToEthClaim) GetType() ClaimType {
	return CLAIM_TYPE_BATCH_SEND_TO_ETH
}

// ValidateBasic performs stateless checks
func (e *MsgBatchSendToEthClaim) ValidateBasic() error {
	if err := libmeta.ValidateBasic(e); err != nil {
		return err
	}
	if e.EventNonce == 0 {
		return fmt.Errorf("event_nonce == 0")
	}
	if e.BatchNonce == 0 {
		return fmt.Errorf("batch_nonce == 0")
	}
	if err := ValidateEthAddress(e.TokenContract); err != nil {
		return sdkerrors.Wrap(err, "erc20 token")
	}
	return nil
}

// Hash implements WithdrawBatch.Hash
func (msg *MsgBatchSendToEthClaim) ClaimHash() ([]byte, error) {
	path := fmt.Sprintf("%d/%d/%d/%s", msg.EventNonce, msg.EthBlockHeight, msg.BatchNonce, msg.TokenContract)
	return tmhash.Sum([]byte(path)), nil
}

// GetSignBytes encodes the message for signing
func (msg MsgBatchSendToEthClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgBatchSendToEthClaim) GetClaimer() sdk.AccAddress {
	err := msg.ValidateBasic()
	if err != nil {
		panic(fmt.Errorf("MsgBatchSendToEthClaim failed ValidateBasic! Should have been handled earlier: %v", err))
	}
	val, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(fmt.Errorf("Invalid orchestrator: %v", err))
	}
	return val
}

// GetSigners defines whose signature is required
func (msg MsgBatchSendToEthClaim) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&msg)
}

// Route should return the name of the module
func (msg MsgBatchSendToEthClaim) Route() string { return RouterKey }

// Type should return the action
func (msg MsgBatchSendToEthClaim) Type() string { return "batch_send_to_eth_claim" }

const (
	TypeMsgBatchSendToEthClaim = "batch_send_to_eth_claim"
)

// NewMsgCancelSendToEth returns a new msgSetOrchestratorAddress
func NewMsgCancelSendToEth(user sdk.AccAddress, id uint64) *MsgCancelSendToEth {
	return &MsgCancelSendToEth{
		Sender:        user.String(),
		TransactionId: id,
	}
}

// Route should return the name of the module
func (msg *MsgCancelSendToEth) Route() string { return RouterKey }

// Type should return the action
func (msg *MsgCancelSendToEth) Type() string { return "cancel_send_to_eth" }

// ValidateBasic performs stateless checks
func (msg *MsgCancelSendToEth) ValidateBasic() (err error) {
	return libmeta.ValidateBasic(msg)
}

// GetSignBytes encodes the message for signing
func (msg *MsgCancelSendToEth) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg *MsgCancelSendToEth) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

// MsgSubmitBadSignatureEvidence
// ======================================================

// ValidateBasic performs stateless checks
func (e *MsgSubmitBadSignatureEvidence) ValidateBasic() (err error) {
	return libmeta.ValidateBasic(e)
}

// GetSignBytes encodes the message for signing
func (msg MsgSubmitBadSignatureEvidence) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSubmitBadSignatureEvidence) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic("Invalid signer for MsgSubmitBadSignatureEvidence")
	}
	return []sdk.AccAddress{acc}
}

// Type should return the action
func (msg MsgSubmitBadSignatureEvidence) Type() string { return "Submit_Bad_Signature_Evidence" }

// Route should return the name of the module
func (msg MsgSubmitBadSignatureEvidence) Route() string { return RouterKey }

// MsgUpdateParams

// ValidateBasic performs stateless checks
func (msg *MsgUpdateParams) ValidateBasic() (err error) {
	return libmeta.ValidateBasic(msg)
}

// GetSignBytes encodes the message for signing
func (msg MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

// Type should return the action
func (msg MsgUpdateParams) Type() string { return "update_params" }

// Route should return the name of the module
func (msg MsgUpdateParams) Route() string { return RouterKey }
