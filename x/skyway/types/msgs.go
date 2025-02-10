package types

import (
	"encoding/hex"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errorsmod "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/palomachain/paloma/v2/util/libmeta"
	tokenfactorytypes "github.com/palomachain/paloma/v2/x/tokenfactory/types"
	"github.com/palomachain/paloma/v2/x/valset/types"
)

// nolint: exhaustruct
var (
	_ sdk.Msg = &MsgSendToRemote{}
	_ sdk.Msg = &MsgCancelSendToRemote{}
	_ sdk.Msg = &MsgConfirmBatch{}
	_ sdk.Msg = &MsgSendToPalomaClaim{}
	_ sdk.Msg = &MsgBatchSendToRemoteClaim{}
	_ sdk.Msg = &MsgSubmitBadSignatureEvidence{}
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgLightNodeSaleClaim{}
	_ sdk.Msg = &MsgSetERC20ToTokenDenom{}
	_ sdk.Msg = &MsgReplenishLostGrainsProposal{}
)

// NewMsgSendToRemote returns a new msgSendToRemote
func NewMsgSendToRemote(sender sdk.AccAddress, destAddress EthAddress, send sdk.Coin, chainReferenceID string) *MsgSendToRemote {
	return &MsgSendToRemote{
		EthDest:          destAddress.GetAddress().Hex(),
		Amount:           send,
		ChainReferenceId: chainReferenceID,
		Metadata: types.MsgMetadata{
			Creator: sender.String(),
		},
	}
}

// Route should return the name of the module
func (msg MsgSendToRemote) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSendToRemote) Type() string { return "send_to_eth" }

// ValidateBasic runs stateless checks on the message
// Checks if the Eth address is valid
func (msg MsgSendToRemote) ValidateBasic() error {
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
func (msg MsgSendToRemote) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSendToRemote) GetSigners() []sdk.AccAddress {
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

// Route should return the name of the module
func (msg MsgEstimateBatchGas) Route() string { return RouterKey }

// Type should return the action
func (msg MsgEstimateBatchGas) Type() string { return "estimate_batch_gas" }

// ValidateBasic performs stateless checks
func (msg MsgEstimateBatchGas) ValidateBasic() error {
	if err := libmeta.ValidateBasic(&msg); err != nil {
		return err
	}

	if err := ValidateEthAddress(msg.EthSigner); err != nil {
		return sdkerrors.Wrap(err, "eth signer")
	}

	if msg.Estimate < 1 {
		return sdkerrors.Wrap(ErrInvalidClaim, "gas estimate must be positive")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgEstimateBatchGas) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgEstimateBatchGas) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&msg)
}

// EthereumClaim represents a claim on ethereum state
type EthereumClaim interface {
	// All Ethereum claims that we relay from the Skyway contract and into the module
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
	// Returns the reference ID of the remote chain on which this claim was observed.
	GetChainReferenceId() string
	// Returns the reference ID of the remote chain on which this claim was observed.
	GetSkywayNonce() uint64
	// Returns the compass ID that sent the original message
	GetCompassID() string
}

// nolint: exhaustruct
var (
	_ EthereumClaim = &MsgSendToPalomaClaim{}
	_ EthereumClaim = &MsgBatchSendToRemoteClaim{}
	_ EthereumClaim = &MsgLightNodeSaleClaim{}
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
	// MsgSendToRemote has it's destination as a string many invalid inputs are possible
	// the orchestrator will convert these invalid deposits to simply the string invalid'
	// this is done because the oracle requires an event be processed on Cosmos for each event
	// nonce on the Ethereum side, otherwise (A) the oracle will never proceed and (B) the funds
	// sent with the invalid deposit will forever be lost, with no representation minted anywhere
	// on cosmos. The attestation handler deals with this by managing invalid deposits and placing
	// them into the community pool
	if msg.EventNonce == 0 {
		return fmt.Errorf("event_nonce must be positive")
	}
	if msg.SkywayNonce == 0 {
		return fmt.Errorf("skyway_nonce must be positive")
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
	path := fmt.Sprintf("%d/%d/%s/%s/%s/%s/%s", msg.SkywayNonce, msg.EthBlockHeight, msg.TokenContract, msg.Amount.String(), msg.EthereumSender, msg.PalomaReceiver, msg.CompassId)
	return tmhash.Sum([]byte(path)), nil
}

func (msg *MsgSendToPalomaClaim) GetCompassID() string {
	return msg.CompassId
}

func (msg *MsgBatchSendToRemoteClaim) SetOrchestrator(orchestrator sdk.AccAddress) {
	msg.Orchestrator = orchestrator.String()
}

// GetType returns the claim type
func (msg *MsgBatchSendToRemoteClaim) GetType() ClaimType {
	return CLAIM_TYPE_BATCH_SEND_TO_ETH
}

// ValidateBasic performs stateless checks
func (e *MsgBatchSendToRemoteClaim) ValidateBasic() error {
	if err := libmeta.ValidateBasic(e); err != nil {
		return err
	}
	if e.EventNonce == 0 {
		return fmt.Errorf("event_nonce == 0")
	}
	if e.SkywayNonce == 0 {
		return fmt.Errorf("skyway_nonce == 0")
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
func (msg *MsgBatchSendToRemoteClaim) ClaimHash() ([]byte, error) {
	path := fmt.Sprintf("%d/%d/%d/%s/%s", msg.SkywayNonce, msg.EthBlockHeight, msg.BatchNonce, msg.TokenContract, msg.CompassId)
	return tmhash.Sum([]byte(path)), nil
}

func (msg *MsgBatchSendToRemoteClaim) GetCompassID() string {
	return msg.CompassId
}

func (msg MsgBatchSendToRemoteClaim) GetClaimer() sdk.AccAddress {
	err := msg.ValidateBasic()
	if err != nil {
		panic(fmt.Errorf("MsgBatchSendToRemoteClaim failed ValidateBasic! Should have been handled earlier: %v", err))
	}
	val, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(fmt.Errorf("Invalid orchestrator: %v", err))
	}
	return val
}

// GetSigners defines whose signature is required
func (msg MsgBatchSendToRemoteClaim) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&msg)
}

// Route should return the name of the module
func (msg MsgBatchSendToRemoteClaim) Route() string { return RouterKey }

// Type should return the action
func (msg MsgBatchSendToRemoteClaim) Type() string { return "batch_send_to_eth_claim" }

const (
	TypeMsgBatchSendToRemoteClaim = "batch_send_to_eth_claim"
)

// NewMsgCancelSendToRemote returns a new msgSetOrchestratorAddress
func NewMsgCancelSendToRemote(user sdk.AccAddress, id uint64) *MsgCancelSendToRemote {
	return &MsgCancelSendToRemote{
		Metadata: types.MsgMetadata{
			Creator: user.String(),
		},
		TransactionId: id,
	}
}

// Route should return the name of the module
func (msg *MsgCancelSendToRemote) Route() string { return RouterKey }

// Type should return the action
func (msg *MsgCancelSendToRemote) Type() string { return "cancel_send_to_eth" }

// GetSigners defines whose signature is required
func (msg *MsgCancelSendToRemote) GetSigners() []sdk.AccAddress {
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

// GetSigners defines whose signature is required
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

// Type should return the action
func (msg MsgUpdateParams) Type() string { return "update_params" }

// Route should return the name of the module
func (msg MsgUpdateParams) Route() string { return RouterKey }

// MsgLightNodeSaleClaim
// ======================================================

// ValidateBasic performs stateless checks
func (msg *MsgLightNodeSaleClaim) ValidateBasic() error {
	if err := libmeta.ValidateBasic(msg); err != nil {
		return err
	}

	if msg.EventNonce == 0 {
		return fmt.Errorf("event_nonce must be positive")
	}

	if msg.SkywayNonce == 0 {
		return fmt.Errorf("skyway_nonce must be positive")
	}

	return nil
}

func (msg *MsgLightNodeSaleClaim) ClaimHash() ([]byte, error) {
	path := fmt.Sprintf("%d/%d/%s/%s/%s", msg.SkywayNonce, msg.EthBlockHeight,
		msg.ClientAddress, msg.Amount.String(), msg.CompassId)
	return tmhash.Sum([]byte(path)), nil
}

func (msg *MsgLightNodeSaleClaim) GetCompassID() string {
	return msg.CompassId
}

func (msg MsgLightNodeSaleClaim) GetClaimer() sdk.AccAddress {
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

// GetType returns the type of the claim
func (msg *MsgLightNodeSaleClaim) GetType() ClaimType {
	return CLAIM_TYPE_LIGHT_NODE_SALE
}

func (msg *MsgLightNodeSaleClaim) SetOrchestrator(orchestrator sdk.AccAddress) {
	msg.Orchestrator = orchestrator.String()
}

func NewMsgSetERC20ToTokenDenom(sender sdk.AccAddress, erc20 EthAddress, chainReferenceID string, denom string) *MsgSetERC20ToTokenDenom {
	return &MsgSetERC20ToTokenDenom{
		Metadata:         types.MsgMetadata{Creator: sender.String(), Signers: []string{sender.String()}},
		Denom:            denom,
		ChainReferenceId: chainReferenceID,
		Erc20:            erc20.address.Hex(),
	}
}

func (msg *MsgSetERC20ToTokenDenom) ValidateBasic() error {
	_, _, err := tokenfactorytypes.DeconstructDenom(msg.Denom)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalid, "no tokenfactory denom")
	}

	if err := ValidateEthAddress(msg.Erc20); err != nil {
		return sdkerrors.Wrap(err, "invalid erc20 address")
	}

	if msg.ChainReferenceId == "" {
		return sdkerrors.Wrap(ErrInvalid, "chain reference id cannot be empty")
	}

	return libmeta.ValidateBasic(msg)
}

func (msg *MsgSetERC20ToTokenDenom) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg MsgSetERC20ToTokenDenom) Type() string  { return "set-erc20-to-denom" }
func (msg MsgSetERC20ToTokenDenom) Route() string { return RouterKey }

func (msg MsgReplenishLostGrainsProposal) Route() string { return RouterKey }
func (msg MsgReplenishLostGrainsProposal) Type() string  { return "replenish_lost_grains_proposal" }
func (msg MsgReplenishLostGrainsProposal) ValidateBasic() error {
	return libmeta.ValidateBasic(&msg)
}

func (msg MsgReplenishLostGrainsProposal) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgReplenishLostGrainsProposal) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&msg)
}

func (msg MsgSetERC20MappingProposal) Route() string { return RouterKey }
func (msg MsgSetERC20MappingProposal) Type() string  { return "set_erc20_to_denom_proposal" }
func (msg MsgSetERC20MappingProposal) ValidateBasic() error {
	return libmeta.ValidateBasic(&msg)
}

func (msg MsgSetERC20MappingProposal) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgSetERC20MappingProposal) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&msg)
}
