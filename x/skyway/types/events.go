package types

const (
	EventTypeObservation                 = "observation"
	EventTypeOutgoingBatch               = "outgoing_batch"
	EventTypeMultisigUpdateRequest       = "multisig_update_request"
	EventTypeOutgoingBatchCanceled       = "outgoing_batch_canceled"
	EventTypeBridgeWithdrawalReceived    = "withdrawal_received"
	EventTypeBridgeDepositReceived       = "deposit_received"
	EventTypeBridgeWithdrawCanceled      = "withdraw_canceled"
	EventTypeInvalidSendToPalomaReceiver = "invalid_send_to_paloma_receiver"

	AttributeKeyAttestationID          = "attestation_id"
	AttributeKeyBatchConfirmKey        = "batch_confirm_key"
	AttributeKeyMultisigID             = "multisig_id"
	AttributeKeyOutgoingBatchID        = "batch_id"
	AttributeKeyOutgoingTXID           = "outgoing_tx_id"
	AttributeKeyAttestationType        = "attestation_type"
	AttributeKeyContract               = "bridge_contract"
	AttributeKeyNonce                  = "nonce"
	AttributeKeyBatchNonce             = "batch_nonce"
	AttributeKeyBridgeChainID          = "bridge_chain_id"
	AttributeKeySetOperatorAddr        = "set_operator_address"
	AttributeKeyBadEthSignature        = "bad_eth_signature"
	AttributeKeyBadEthSignatureSubject = "bad_eth_signature_subject"

	AttributeKeySendToPalomaAmount = "msg_send_to_cosmsos_amount"
	AttributeKeySendToPalomaNonce  = "msg_send_to_cosmsos_nonce"
	AttributeKeySendToPalomaToken  = "msg_send_to_cosmsos_token"
	AttributeKeySendToPalomaSender = "msg_send_to_cosmsos_sender"

	AttributeKeyBatchSignatureSlashing = "batch_signature_slashing"
)
