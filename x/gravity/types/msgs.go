package types

import (
	"encoding/hex"
	"fmt"

	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// nolint: exhaustruct
var (
	_ sdk.Msg = &MsgSetOrchestratorAddress{}
	_ sdk.Msg = &MsgValsetConfirm{}
	_ sdk.Msg = &MsgSendToEth{}
	_ sdk.Msg = &MsgCancelSendToEth{}
	_ sdk.Msg = &MsgRequestBatch{}
	_ sdk.Msg = &MsgConfirmBatch{}
	_ sdk.Msg = &MsgERC20DeployedClaim{}
	_ sdk.Msg = &MsgConfirmLogicCall{}
	_ sdk.Msg = &MsgLogicCallExecutedClaim{}
	_ sdk.Msg = &MsgSendToCosmosClaim{}
	_ sdk.Msg = &MsgExecuteIbcAutoForwards{}
	_ sdk.Msg = &MsgBatchSendToEthClaim{}
	_ sdk.Msg = &MsgValsetUpdatedClaim{}
	_ sdk.Msg = &MsgSubmitBadSignatureEvidence{}
)

// NewMsgSetOrchestratorAddress returns a new msgSetOrchestratorAddress
func NewMsgSetOrchestratorAddress(val sdk.ValAddress, oper sdk.AccAddress, eth EthAddress) *MsgSetOrchestratorAddress {
	return &MsgSetOrchestratorAddress{
		Validator:    val.String(),
		Orchestrator: oper.String(),
		EthAddress:   eth.GetAddress().Hex(),
	}
}

// Route should return the name of the module
func (msg *MsgSetOrchestratorAddress) Route() string { return RouterKey }

// Type should return the action
func (msg *MsgSetOrchestratorAddress) Type() string { return "set_operator_address" }

// ValidateBasic performs stateless checks
func (msg *MsgSetOrchestratorAddress) ValidateBasic() (err error) {
	if _, err = sdk.ValAddressFromBech32(msg.Validator); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Validator)
	}
	if _, err = sdk.AccAddressFromBech32(msg.Orchestrator); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Orchestrator)
	}
	if err := ValidateEthAddress(msg.EthAddress); err != nil {
		return sdkerrors.Wrap(err, "ethereum address")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgSetOrchestratorAddress) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg *MsgSetOrchestratorAddress) GetSigners() []sdk.AccAddress {
	acc, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sdk.AccAddress(acc)}
}

// NewMsgValsetConfirm returns a new msgValsetConfirm
func NewMsgValsetConfirm(
	nonce uint64,
	ethAddress EthAddress,
	validator sdk.AccAddress,
	signature string,
) *MsgValsetConfirm {
	return &MsgValsetConfirm{
		Nonce:        nonce,
		Orchestrator: validator.String(),
		EthAddress:   ethAddress.GetAddress().Hex(),
		Signature:    signature,
	}
}

// Route should return the name of the module
func (msg *MsgValsetConfirm) Route() string { return RouterKey }

// Type should return the action
func (msg *MsgValsetConfirm) Type() string { return "valset_confirm" }

// ValidateBasic performs stateless checks
func (msg *MsgValsetConfirm) ValidateBasic() (err error) {
	if _, err = sdk.AccAddressFromBech32(msg.Orchestrator); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Orchestrator)
	}
	if err := ValidateEthAddress(msg.EthAddress); err != nil {
		return sdkerrors.Wrap(err, "ethereum address")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgValsetConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg *MsgValsetConfirm) GetSigners() []sdk.AccAddress {
	// TODO: figure out how to convert between AccAddress and ValAddress properly
	acc, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{acc}
}

// NewMsgSendToEth returns a new msgSendToEth
func NewMsgSendToEth(sender sdk.AccAddress, destAddress EthAddress, send sdk.Coin, bridgeFee sdk.Coin, chainFee sdk.Coin) *MsgSendToEth {
	return &MsgSendToEth{
		Sender:    sender.String(),
		EthDest:   destAddress.GetAddress().Hex(),
		Amount:    send,
		BridgeFee: bridgeFee, // This is paid to the Ethereum Relayer
		ChainFee:  chainFee,  // This is paid to Cosmos stakers
	}
}

// Route should return the name of the module
func (msg MsgSendToEth) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSendToEth) Type() string { return "send_to_eth" }

// ValidateBasic runs stateless checks on the message
// Checks if the Eth address is valid
func (msg MsgSendToEth) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Sender)
	}

	// bridge fee (paid to relayers) and send must be of the same denom
	if msg.Amount.Denom != msg.BridgeFee.Denom {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins,
			fmt.Sprintf("bridge fee (paid to relayers) and amount must be the same type %s != %s", msg.Amount.Denom, msg.BridgeFee.Denom))
	}
	// chain fee (paid to cosmos stakers) and send must be of the same denom
	if msg.Amount.Denom != msg.ChainFee.Denom {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins,
			fmt.Sprintf("chain fee (paid to stakers) and amount must be the same type %s != %s", msg.Amount.Denom, msg.ChainFee.Denom))
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "amount")
	}
	if !msg.BridgeFee.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "bridge fee")
	}
	if !msg.ChainFee.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "chain fee")
	}
	if err := ValidateEthAddress(msg.EthDest); err != nil {
		return sdkerrors.Wrap(err, "ethereum address")
	}
	// TODO validate fee is sufficient, fixed fee to start
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSendToEth) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSendToEth) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{acc}
}

// NewMsgRequestBatch returns a new msgRequestBatch
func NewMsgRequestBatch(orchestrator sdk.AccAddress, chainReferenceId string) *MsgRequestBatch {
	return &MsgRequestBatch{
		Sender:           orchestrator.String(),
		ChainReferenceId: chainReferenceId,
	}
}

// Route should return the name of the module
func (msg MsgRequestBatch) Route() string { return RouterKey }

// Type should return the action
func (msg MsgRequestBatch) Type() string { return "request_batch" }

// ValidateBasic performs stateless checks
func (msg MsgRequestBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Sender)
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgRequestBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgRequestBatch) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{acc}
}

// Route should return the name of the module
func (msg MsgConfirmBatch) Route() string { return RouterKey }

// Type should return the action
func (msg MsgConfirmBatch) Type() string { return "confirm_batch" }

// ValidateBasic performs stateless checks
func (msg MsgConfirmBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Orchestrator); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Orchestrator)
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
	acc, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{acc}
}

// Route should return the name of the module
func (msg MsgConfirmLogicCall) Route() string { return RouterKey }

// Type should return the action
func (msg MsgConfirmLogicCall) Type() string { return "confirm_logic" }

// ValidateBasic performs stateless checks
func (msg MsgConfirmLogicCall) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Orchestrator); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Orchestrator)
	}
	if err := ValidateEthAddress(msg.EthSigner); err != nil {
		return sdkerrors.Wrap(err, "eth signer")
	}
	_, err := hex.DecodeString(msg.Signature)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "Could not decode hex string %s", msg.Signature)
	}
	_, err = hex.DecodeString(msg.InvalidationId)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "Could not decode hex string %s", msg.InvalidationId)
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgConfirmLogicCall) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgConfirmLogicCall) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{acc}
}

// EthereumClaim represents a claim on ethereum state
type EthereumClaim interface {
	// All Ethereum claims that we relay from the Gravity contract and into the module
	// have a nonce that is strictly increasing and unique, since this nonce is
	// issued by the Ethereum contract it is immutable and must be agreed on by all validators
	// any disagreement on what claim goes to what nonce means someone is lying.
	GetEventNonce() uint64
	// The block height that the claimed event occurred on. This EventNonce provides sufficient
	// ordering for the execution of all claims. The block height is used only for batchTimeouts + logicTimeouts
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
	_ EthereumClaim = &MsgSendToCosmosClaim{}
	_ EthereumClaim = &MsgBatchSendToEthClaim{}
	_ EthereumClaim = &MsgERC20DeployedClaim{}
	_ EthereumClaim = &MsgLogicCallExecutedClaim{}
)

func (msg *MsgSendToCosmosClaim) SetOrchestrator(orchestrator sdk.AccAddress) {
	msg.Orchestrator = orchestrator.String()
}

// GetType returns the type of the claim
func (msg *MsgSendToCosmosClaim) GetType() ClaimType {
	return CLAIM_TYPE_SEND_TO_COSMOS
}

// ValidateBasic performs stateless checks
func (msg *MsgSendToCosmosClaim) ValidateBasic() error {
	if err := ValidateEthAddress(msg.EthereumSender); err != nil {
		return sdkerrors.Wrap(err, "eth sender")
	}
	if err := ValidateEthAddress(msg.TokenContract); err != nil {
		return sdkerrors.Wrap(err, "erc20 token")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Orchestrator); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "orchestrator")
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
func (msg MsgSendToCosmosClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgSendToCosmosClaim) GetClaimer() sdk.AccAddress {
	err := msg.ValidateBasic()
	if err != nil {
		panic("MsgSendToCosmosClaim failed ValidateBasic! Should have been handled earlier")
	}

	val, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}

	return val
}

// GetSigners defines whose signature is required
func (msg MsgSendToCosmosClaim) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{acc}
}

// Type should return the action
func (msg MsgSendToCosmosClaim) Type() string { return "send_to_cosmos_claim" }

// Route should return the name of the module
func (msg MsgSendToCosmosClaim) Route() string { return RouterKey }

const (
	TypeMsgSendToCosmosClaim = "send_to_cosmos_claim"
)

// Hash implements BridgeDeposit.Hash
// modify this with care as it is security sensitive. If an element of the claim is not in this hash a single hostile validator
// could engineer a hash collision and execute a version of the claim with any unhashed data changed to benefit them.
// note that the Orchestrator is the only field excluded from this hash, this is because that value is used higher up in the store
// structure for who has made what claim and is verified by the msg ante-handler for signatures
func (msg *MsgSendToCosmosClaim) ClaimHash() ([]byte, error) {
	path := fmt.Sprintf("%d/%d/%s/%s/%s/%s", msg.EventNonce, msg.EthBlockHeight, msg.TokenContract, msg.Amount.String(), msg.EthereumSender, msg.CosmosReceiver)
	return tmhash.Sum([]byte(path)), nil
}

func (msg *MsgExecuteIbcAutoForwards) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Executor); err != nil {
		return sdkerrors.Wrap(err, "Unable to parse executor as a valid bech32 address")
	}
	return nil
}

func (msg *MsgExecuteIbcAutoForwards) GetSigners() []sdk.AccAddress {
	msg.ProtoMessage()
	acc, err := sdk.AccAddressFromBech32(msg.Executor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{acc}
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
	if e.EventNonce == 0 {
		return fmt.Errorf("event_nonce == 0")
	}
	if e.BatchNonce == 0 {
		return fmt.Errorf("batch_nonce == 0")
	}
	if err := ValidateEthAddress(e.TokenContract); err != nil {
		return sdkerrors.Wrap(err, "erc20 token")
	}
	if _, err := sdk.AccAddressFromBech32(e.Orchestrator); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, e.Orchestrator)
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
	acc, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{acc}
}

// Route should return the name of the module
func (msg MsgBatchSendToEthClaim) Route() string { return RouterKey }

// Type should return the action
func (msg MsgBatchSendToEthClaim) Type() string { return "batch_send_to_eth_claim" }

const (
	TypeMsgBatchSendToEthClaim = "batch_send_to_eth_claim"
)

// EthereumClaim implementation for MsgERC20DeployedClaim
// ======================================================

func (msg *MsgERC20DeployedClaim) SetOrchestrator(orchestrator sdk.AccAddress) {
	msg.Orchestrator = orchestrator.String()
}

// GetType returns the type of the claim
func (e *MsgERC20DeployedClaim) GetType() ClaimType {
	return CLAIM_TYPE_ERC20_DEPLOYED
}

// ValidateBasic performs stateless checks
func (e *MsgERC20DeployedClaim) ValidateBasic() error {
	if err := ValidateEthAddress(e.TokenContract); err != nil {
		return sdkerrors.Wrap(err, "erc20 token")
	}
	if _, err := sdk.AccAddressFromBech32(e.Orchestrator); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, e.Orchestrator)
	}
	if e.EventNonce == 0 {
		return fmt.Errorf("nonce == 0")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgERC20DeployedClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgERC20DeployedClaim) GetClaimer() sdk.AccAddress {
	err := msg.ValidateBasic()
	if err != nil {
		panic("MsgERC20DeployedClaim failed ValidateBasic! Should have been handled earlier")
	}

	val, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}
	return val
}

// GetSigners defines whose signature is required
func (msg MsgERC20DeployedClaim) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{acc}
}

// Type should return the action
func (msg MsgERC20DeployedClaim) Type() string { return "ERC20_deployed_claim" }

// Route should return the name of the module
func (msg MsgERC20DeployedClaim) Route() string { return RouterKey }

// Hash implements BridgeDeposit.Hash
// modify this with care as it is security sensitive. If an element of the claim is not in this hash a single hostile validator
// could engineer a hash collision and execute a version of the claim with any unhashed data changed to benefit them.
// note that the Orchestrator is the only field excluded from this hash, this is because that value is used higher up in the store
// structure for who has made what claim and is verified by the msg ante-handler for signatures
func (b *MsgERC20DeployedClaim) ClaimHash() ([]byte, error) {
	path := fmt.Sprintf("%d/%d/%s/%s/%s/%s/%d", b.EventNonce, b.EthBlockHeight, b.CosmosDenom, b.TokenContract, b.Name, b.Symbol, b.Decimals)
	return tmhash.Sum([]byte(path)), nil
}

// EthereumClaim implementation for MsgLogicCallExecutedClaim
// ======================================================

func (msg *MsgLogicCallExecutedClaim) SetOrchestrator(orchestrator sdk.AccAddress) {
	msg.Orchestrator = orchestrator.String()
}

// GetType returns the type of the claim
func (e *MsgLogicCallExecutedClaim) GetType() ClaimType {
	return CLAIM_TYPE_LOGIC_CALL_EXECUTED
}

// ValidateBasic performs stateless checks
func (e *MsgLogicCallExecutedClaim) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(e.Orchestrator); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, e.Orchestrator)
	}
	if e.EventNonce == 0 {
		return fmt.Errorf("nonce == 0")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgLogicCallExecutedClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgLogicCallExecutedClaim) GetClaimer() sdk.AccAddress {
	err := msg.ValidateBasic()
	if err != nil {
		panic("MsgERC20DeployedClaim failed ValidateBasic! Should have been handled earlier")
	}

	val, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}
	return val
}

// GetSigners defines whose signature is required
func (msg MsgLogicCallExecutedClaim) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{acc}
}

// Type should return the action
func (msg MsgLogicCallExecutedClaim) Type() string { return "Logic_Call_Executed_Claim" }

// Route should return the name of the module
func (msg MsgLogicCallExecutedClaim) Route() string { return RouterKey }

// Hash implements BridgeDeposit.Hash
// modify this with care as it is security sensitive. If an element of the claim is not in this hash a single hostile validator
// could engineer a hash collision and execute a version of the claim with any unhashed data changed to benefit them.
// note that the Orchestrator is the only field excluded from this hash, this is because that value is used higher up in the store
// structure for who has made what claim and is verified by the msg ante-handler for signatures
func (b *MsgLogicCallExecutedClaim) ClaimHash() ([]byte, error) {
	path := fmt.Sprintf("%d/%d/%s/%d", b.EventNonce, b.EthBlockHeight, b.InvalidationId, b.InvalidationNonce)
	return tmhash.Sum([]byte(path)), nil
}

// EthereumClaim implementation for MsgValsetUpdatedClaim
// ======================================================
func (e *MsgValsetUpdatedClaim) SetOrchestrator(orchestrator sdk.AccAddress) {
	e.Orchestrator = orchestrator.String()
}

// GetType returns the type of the claim
func (e *MsgValsetUpdatedClaim) GetType() ClaimType {
	return CLAIM_TYPE_VALSET_UPDATED
}

// ValidateBasic performs stateless checks
func (e *MsgValsetUpdatedClaim) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(e.Orchestrator); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, e.Orchestrator)
	}
	if e.EventNonce == 0 {
		return fmt.Errorf("nonce == 0")
	}

	err := ValidateEthAddress(e.RewardToken)
	if err != nil {
		return err
	}

	for _, member := range e.Members {
		err := ValidateEthAddress(member.EthereumAddress)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgValsetUpdatedClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgValsetUpdatedClaim) GetClaimer() sdk.AccAddress {
	err := msg.ValidateBasic()
	if err != nil {
		panic(fmt.Errorf("MsgERC20DeployedClaim failed ValidateBasic! Should have been handled earlier: %v", err))
	}

	val, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(fmt.Errorf("Orchestrator is invalid: %v", err))
	}
	return val
}

// GetSigners defines whose signature is required
func (msg MsgValsetUpdatedClaim) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Orchestrator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{acc}
}

// Type should return the action
func (msg MsgValsetUpdatedClaim) Type() string { return "Valset_Updated_Claim" }

// Route should return the name of the module
func (msg MsgValsetUpdatedClaim) Route() string { return RouterKey }

// Hash implements BridgeDeposit.Hash
// modify this with care as it is security sensitive. If an element of the claim is not in this hash a single hostile validator
// could engineer a hash collision and execute a version of the claim with any unhashed data changed to benefit them.
// note that the Orchestrator is the only field excluded from this hash, this is because that value is used higher up in the store
// structure for who has made what claim and is verified by the msg ante-handler for signatures
func (b *MsgValsetUpdatedClaim) ClaimHash() ([]byte, error) {
	var members BridgeValidators = b.Members
	internalMembers, err := members.ToInternal()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid members")
	}
	internalMembers.Sort()
	path := fmt.Sprintf("%d/%d/%d/%x/%s/%s", b.EventNonce, b.ValsetNonce, b.EthBlockHeight, internalMembers.ToExternal(), b.RewardAmount.String(), b.RewardToken)
	return tmhash.Sum([]byte(path)), nil
}

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
	_, err = sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgCancelSendToEth) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg *MsgCancelSendToEth) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{acc}
}

// MsgSubmitBadSignatureEvidence
// ======================================================

// ValidateBasic performs stateless checks
func (e *MsgSubmitBadSignatureEvidence) ValidateBasic() (err error) {
	_, err = sdk.AccAddressFromBech32(e.Sender)
	if err != nil {
		return err
	}
	return nil
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
