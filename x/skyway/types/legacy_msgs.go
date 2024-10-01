package types

import (
	fmt "fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/libmeta"
)

var _ sdk.Msg = &MsgBatchSendToEthClaim{}

var _ EthereumClaim = &MsgBatchSendToEthClaim{}

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
func (msg *MsgBatchSendToEthClaim) ClaimHash() ([]byte, error) {
	path := fmt.Sprintf("%d/%d/%d/%s", msg.SkywayNonce, msg.EthBlockHeight, msg.BatchNonce, msg.TokenContract)
	return tmhash.Sum([]byte(path)), nil
}

func (msg *MsgBatchSendToEthClaim) GetCompassID() string { return "" }

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
