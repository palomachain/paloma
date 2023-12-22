package types

import (
	"strings"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "gravity2"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the module name router key
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName
)

var (
	// EthAddressByValidatorKey indexes cosmos validator account addresses
	// i.e. paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk
	// [0x1248a4405201cc3a00ab515ce9c4dd47]
	EthAddressByValidatorKey = HashString("EthAddressValidatorKey")

	// ValidatorByEthAddressKey indexes ethereum addresses
	// i.e. 0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B
	// [0xbfe41763f372108317ed982a4cd1b4a8]
	ValidatorByEthAddressKey = HashString("ValidatorByEthAddressKey")

	// LEGACYOracleClaimKey Claim details by nonce and validator address
	// Note: This is a LEGACY key, i.e. it is no longer in use!
	// ** DO NOT USE THIS OUTSIDE OF MIGRATION TESTING! **
	//
	// i.e. gravityvaloper1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm
	// A claim is named more intuitively than an Attestation, it is literally
	// a validator making a claim to have seen something happen. Claims are
	// attached to attestations which can be thought of as 'the event' that
	// will eventually be executed.
	// [0x15968a4f1cb06e26ab155d6e59eccc85]
	LEGACYOracleClaimKey = HashString("OracleClaimKey")

	// OracleAttestationKey attestation details by nonce and validator address
	// i.e. gravityvaloper1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm
	// An attestation can be thought of as the 'event to be executed' while
	// the Claims are an individual validator saying that they saw an event
	// occur the Attestation is 'the event' that multiple claims vote on and
	// eventually executes
	// [0x0bfa165ff4ef558b3d0b62ea4d4a46c5]
	OracleAttestationKey = HashString("OracleAttestationKey")

	// OutgoingTXPoolKey indexes the last nonce for the outgoing tx pool
	// [0x44f7816ec23d990ee39d9ed4609bbd4d]
	OutgoingTXPoolKey = HashString("OutgoingTXPoolKey")

	// OutgoingTXBatchKey indexes outgoing tx batches under a nonce and token address
	// [0x54e2db44755d8865b1ff4227402e204f]
	OutgoingTXBatchKey = HashString("OutgoingTXBatchKey")

	// BatchConfirmKey indexes validator confirmations by token contract address
	// [0x75b935a854d50880236724b9c4822daf]
	BatchConfirmKey = HashString("BatchConfirmKey")

	// LastEventNonceByValidatorKey indexes lateset event nonce by validator
	// [0xeefcb999cc3d7b80b052b55106a6ba5e]
	LastEventNonceByValidatorKey = HashString("LastEventNonceByValidatorKey")

	// LastObservedEventNonceKey indexes the latest event nonce
	// [0xa34e56ab6fab9ee91e82ba216bfeb759]
	LastObservedEventNonceKey = HashString("LastObservedEventNonceKey")

	// LEGACYSequenceKeyPrefix indexes different txids
	// Note: This is a LEGACY key, i.e. it is no longer in use!
	// ** DO NOT USE THIS OUTSIDE OF MIGRATION TESTING! **
	//
	// [0x33887862fa4fba9c592d6fb84d8dd755]
	LEGACYSequenceKeyPrefix = HashString("SequenceKeyPrefix")

	// KeyLastTXPoolID indexes the lastTxPoolID
	// [0xfd87a616141bfbd26fd2938d6e1cf099]
	KeyLastTXPoolID = HashString("SequenceKeyPrefix" + "lastTxPoolId")

	// KeyLastOutgoingBatchID indexes the lastBatchID
	// [0x4f9c42e30316353cb1e0ed74200abbbb]
	KeyLastOutgoingBatchID = HashString("SequenceKeyPrefix" + "lastBatchId")

	// KeyOrchestratorAddress indexes the validator keys for an orchestrator
	// [0x391e8708521fb085676169e8fb232cda]
	KeyOrchestratorAddress = HashString("KeyOrchestratorAddress")

	// LastObservedEthereumBlockHeightKey indexes the latest Ethereum block height
	// [0x83a283a6c3390f1526250df45e9ef8c6]
	LastObservedEthereumBlockHeightKey = HashString("LastObservedEthereumBlockHeightKey")

	// DenomToERC20Key prefixes the index of Cosmos originated asset denoms to ERC20s
	// [0x19fb4f512868744eea13f3eac3672c12]
	DenomToERC20Key = HashString("DenomToERC20Key")

	// ERC20ToDenomKey prefixes the index of Cosmos originated assets ERC20s to denoms
	// [0x877b20a916c830ad4db23d22f7b2cde0]
	ERC20ToDenomKey = HashString("ERC20ToDenomKey")

	// LastSlashedBatchBlock indexes the latest slashed batch block height
	// [0xcbaedf5dd1e068d9e2223281f693358c]
	LastSlashedBatchBlock = HashString("LastSlashedBatchBlock")

	// LastUnBondingBlockHeight indexes the last validator unbonding block height
	// [0x06a6b30651341e80276e0d2e19449250]
	LastUnBondingBlockHeight = HashString("LastUnBondingBlockHeight")

	// PastEthSignatureCheckpointKey indexes eth signature checkpoints that have existed
	// [0x1cbe0be407a979331b98e599eeedd09f]
	PastEthSignatureCheckpointKey = HashString("PastEthSignatureCheckpointKey")
	ParamsKey                     = []byte{0x01}
)

// GetOrchestratorAddressKey returns the following key format
// prefix 				orchestrator address
// [0x0][paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk]
func GetOrchestratorAddressKey(orc sdk.AccAddress) ([]byte, error) {
	if err := sdk.VerifyAddressFormat(orc); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid orchestrator address")
	}
	return AppendBytes(KeyOrchestratorAddress, orc.Bytes()), nil
}

// GetEthAddressByValidatorKey returns the following key format
// prefix              cosmos-validator
// [0x0][gravityvaloper1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm]
func GetEthAddressByValidatorKey(validator sdk.ValAddress) ([]byte, error) {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid validator address")
	}
	return AppendBytes(EthAddressByValidatorKey, validator.Bytes()), nil
}

// GetValidatorByEthAddressKey returns the following key format
// prefix              cosmos-validator
// [0x0][0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B]
func GetValidatorByEthAddressKey(ethAddress EthAddress) []byte {
	return AppendBytes(ValidatorByEthAddressKey, ethAddress.GetAddress().Bytes())
}

// GetAttestationKey returns the following key format
// prefix     nonce                             claim-details-hash
// [0x0][0 0 0 0 0 0 0 1][fd1af8cec6c67fcf156f1b61fdf91ebc04d05484d007436e75342fc05bbff35a]
// An attestation is an event multiple people are voting on, this function needs the claim
// details because each Attestation is aggregating all claims of a specific event, lets say
// validator X and validator y were making different claims about the same event nonce
// Note that the claim hash does NOT include the claimer address and only identifies an event
func GetAttestationKey(eventNonce uint64, claimHash []byte) []byte {
	return AppendBytes(OracleAttestationKey, UInt64Bytes(eventNonce), claimHash)
}

// GetOutgoingTxPoolContractPrefix returns
// [0x0][0xc783df8a850f42e7F7e57013759C285caa701eB6]
// This prefix is used for iterating over unbatched transactions for a given contract
func GetOutgoingTxPoolContractPrefix(contractAddress EthAddress) []byte {
	return AppendBytes(OutgoingTXPoolKey, contractAddress.GetAddress().Bytes())
}

// GetOutgoingTxPoolKey returns the following key format
// prefix				token					 amount			id
// [0x0][0xc783df8a850f42e7F7e57013759C285caa701eB6][1000000000][0 0 0 0 0 0 0 1]
func GetOutgoingTxPoolKey(token InternalERC20Token, id uint64) []byte {
	amount := make([]byte, 32)
	amount = token.Amount.BigInt().FillBytes(amount)
	return AppendBytes(OutgoingTXPoolKey, token.Contract.GetAddress().Bytes(), amount, UInt64Bytes(id))
}

// GetOutgoingTxBatchContractPrefix returns the following format
// prefix     eth-contract-address
// [0x0][0xc783df8a850f42e7F7e57013759C285caa701eB6]
func GetOutgoingTxBatchContractPrefix(tokenContract EthAddress) []byte {
	return AppendBytes(OutgoingTXBatchKey, tokenContract.GetAddress().Bytes())
}

// GetOutgoingTxBatchKey returns the following key format
// prefix     eth-contract-address                     nonce
// [0x0][0xc783df8a850f42e7F7e57013759C285caa701eB6][0 0 0 0 0 0 0 1]
func GetOutgoingTxBatchKey(tokenContract EthAddress, nonce uint64) []byte {
	return AppendBytes(GetOutgoingTxBatchContractPrefix(tokenContract), UInt64Bytes(nonce))
}

// GetBatchConfirmNonceContractPrefix returns
// prefix           eth-contract-address                BatchNonce
// [0x0][0xc783df8a850f42e7F7e57013759C285caa701eB6][0 0 0 0 0 0 0 1]
func GetBatchConfirmNonceContractPrefix(tokenContract EthAddress, batchNonce uint64) []byte {
	return AppendBytes(BatchConfirmKey, tokenContract.GetAddress().Bytes(), UInt64Bytes(batchNonce))
}

// GetBatchConfirmKey returns the following key format
// prefix           eth-contract-address                BatchNonce                       Validator-address
// [0x0		][0xc783df8a850f42e7F7e57013759C285caa701eB6][0 0 0 0 0 0 0 1][gravityvaloper1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm]
func GetBatchConfirmKey(tokenContract EthAddress, batchNonce uint64, validator sdk.AccAddress) ([]byte, error) {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid validator address")
	}
	return AppendBytes(GetBatchConfirmNonceContractPrefix(tokenContract, batchNonce), validator.Bytes()), nil
}

// GetLastEventNonceByValidatorKey indexes latest event nonce by validator
// GetLastEventNonceByValidatorKey returns the following key format
// prefix              cosmos-validator
// [0x0][paloma1ahx7f8wyertuus9r20284ej0asrs085c945jyk]
func GetLastEventNonceByValidatorKey(validator sdk.ValAddress) ([]byte, error) {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid validator address")
	}
	return AppendBytes(LastEventNonceByValidatorKey, validator.Bytes()), nil
}

func GetDenomToERC20Key(chainReferenceId, denom string) []byte {
	return AppendBytes(DenomToERC20Key, []byte(chainReferenceId), []byte(denom))
}

func GetERC20ToDenomKey(chainReferenceId string, erc20 EthAddress) []byte {
	return AppendBytes(ERC20ToDenomKey, []byte(chainReferenceId), erc20.GetAddress().Bytes())
}

// GetPastEthSignatureCheckpointKey returns the following key format
// prefix    checkpoint
// [0x0][ checkpoint bytes ]
func GetPastEthSignatureCheckpointKey(checkpoint []byte) []byte {
	return AppendBytes(PastEthSignatureCheckpointKey, []byte(convertByteArrToString(checkpoint)))
}

// This function is broken and it should not be used in other places except in GetPastEthSignatureCheckpointKey
func convertByteArrToString(value []byte) string {
	var ret strings.Builder
	for i := 0; i < len(value); i++ {
		ret.WriteString(string(value[i]))
	}
	return ret.String()
}
