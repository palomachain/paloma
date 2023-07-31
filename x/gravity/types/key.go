package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// ModuleName is the name of the module
	ModuleName = "gravity"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the module name router key
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName
)

var (
	// EthAddressByValidatorKey indexes cosmos validator account addresses
	// i.e. gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm
	// [0x1248a4405201cc3a00ab515ce9c4dd47]
	EthAddressByValidatorKey = HashString("EthAddressValidatorKey")

	// ValidatorByEthAddressKey indexes ethereum addresses
	// i.e. 0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B
	// [0xbfe41763f372108317ed982a4cd1b4a8]
	ValidatorByEthAddressKey = HashString("ValidatorByEthAddressKey")

	// ValsetRequestKey indexes valset requests by nonce
	// [0xa318dca6c74494c2fac1841c9e2454fe]
	ValsetRequestKey = HashString("ValsetRequestKey")

	// ValsetConfirmKey indexes valset confirmations by nonce and the validator account address
	// i.e gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm
	// [0x2f522701d7f28af84632f8228fbe1750]
	ValsetConfirmKey = HashString("ValsetConfirmKey")

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

	// KeyOutgoingLogicCall indexes the outgoing logic calls
	// [0x98dfff23346c13a1747fbbed5b23d248]
	KeyOutgoingLogicCall = HashString("KeyOutgoingLogicCall")

	// KeyOutgoingLogicConfirm indexes the outgoing logic confirms
	// [0xd244ded2fa29e95a7719ec40696221e4]
	KeyOutgoingLogicConfirm = HashString("KeyOutgoingLogicConfirm")

	// LastObservedEthereumBlockHeightKey indexes the latest Ethereum block height
	// [0x83a283a6c3390f1526250df45e9ef8c6]
	LastObservedEthereumBlockHeightKey = HashString("LastObservedEthereumBlockHeightKey")

	// DenomToERC20Key prefixes the index of Cosmos originated asset denoms to ERC20s
	// [0x19fb4f512868744eea13f3eac3672c12]
	DenomToERC20Key = HashString("DenomToERC20Key")

	// ERC20ToDenomKey prefixes the index of Cosmos originated assets ERC20s to denoms
	// [0x877b20a916c830ad4db23d22f7b2cde0]
	ERC20ToDenomKey = HashString("ERC20ToDenomKey")

	// LastSlashedValsetNonce indexes the latest slashed valset nonce
	// [0x3adee74534260faaea6ac8e31826b09e]
	LastSlashedValsetNonce = HashString("LastSlashedValsetNonce")

	// LatestValsetNonce indexes the latest valset nonce
	// [0xba0fa05da166611b656bac7739a6e7d3]
	LatestValsetNonce = HashString("LatestValsetNonce")

	// LastSlashedBatchBlock indexes the latest slashed batch block height
	// [0xcbaedf5dd1e068d9e2223281f693358c]
	LastSlashedBatchBlock = HashString("LastSlashedBatchBlock")

	// LastSlashedLogicCallBlock indexes the latest slashed logic call block height
	// [0x3df72087ae3f58d49c6d0b1737c8da0c]
	LastSlashedLogicCallBlock = HashString("LastSlashedLogicCallBlock")

	// LastUnBondingBlockHeight indexes the last validator unbonding block height
	// [0x06a6b30651341e80276e0d2e19449250]
	LastUnBondingBlockHeight = HashString("LastUnBondingBlockHeight")

	// PastEthSignatureCheckpointKey indexes eth signature checkpoints that have existed
	// [0x1cbe0be407a979331b98e599eeedd09f]
	PastEthSignatureCheckpointKey = HashString("PastEthSignatureCheckpointKey")

	// LastObservedValsetKey indexes the latest observed valset nonce
	// HERE THERE BE DRAGONS, do not use this value as an up to date validator set
	// on Ethereum it will always lag significantly and may be totally wrong at some
	// times.
	// [0x2418e9d990ce88e9b844b0bb723d4d7a]
	LastObservedValsetKey = HashString("LastObservedValsetKey")

	// PendingIbcAutoForwards indexes pending SendToCosmos sends via IBC, queued by event nonce
	// [0x5b89a7c5dc9abd2a7abc2560d6eb42ea]
	PendingIbcAutoForwards = HashString("IbcAutoForwardQueue")
)

// GetOrchestratorAddressKey returns the following key format
// prefix 				orchestrator address
// [0x0][gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm]
func GetOrchestratorAddressKey(orc sdk.AccAddress) []byte {
	if err := sdk.VerifyAddressFormat(orc); err != nil {
		panic(sdkerrors.Wrap(err, "invalid orchestrator address"))
	}
	return AppendBytes(KeyOrchestratorAddress, orc.Bytes())
}

// GetEthAddressByValidatorKey returns the following key format
// prefix              cosmos-validator
// [0x0][gravityvaloper1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm]
func GetEthAddressByValidatorKey(validator sdk.ValAddress) []byte {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		panic(sdkerrors.Wrap(err, "invalid validator address"))
	}
	return AppendBytes(EthAddressByValidatorKey, validator.Bytes())
}

// GetValidatorByEthAddressKey returns the following key format
// prefix              cosmos-validator
// [0x0][0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B]
func GetValidatorByEthAddressKey(ethAddress EthAddress) []byte {
	return AppendBytes(ValidatorByEthAddressKey, ethAddress.GetAddress().Bytes())
}

// GetValsetKey returns the following key format
// prefix    nonce
// [0x0][0 0 0 0 0 0 0 1]
func GetValsetKey(nonce uint64) []byte {
	return AppendBytes(ValsetRequestKey, UInt64Bytes(nonce))
}

// GetValsetConfirmNoncePrefix returns the following format
// prefix   nonce
// [0x0][0 0 0 0 0 0 0 1]
func GetValsetConfirmNoncePrefix(nonce uint64) []byte {
	return AppendBytes(ValsetConfirmKey, UInt64Bytes(nonce))
}

// GetValsetConfirmKey returns the following key format
// prefix   nonce                    validator-address
// [0x0][0 0 0 0 0 0 0 1][gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm]
// MARK finish-batches: this is where the key is created in the old (presumed working) code
func GetValsetConfirmKey(nonce uint64, validator sdk.AccAddress) []byte {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		panic(sdkerrors.Wrap(err, "invalid validator address"))
	}
	return AppendBytes(GetValsetConfirmNoncePrefix(nonce), validator.Bytes())
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
// prefix			feeContract
// [0x0][0xc783df8a850f42e7F7e57013759C285caa701eB6]
// This prefix is used for iterating over unbatched transactions for a given contract
func GetOutgoingTxPoolContractPrefix(contractAddress EthAddress) []byte {
	return AppendBytes(OutgoingTXPoolKey, contractAddress.GetAddress().Bytes())
}

// GetOutgoingTxPoolKey returns the following key format
// prefix				feeContract					 feeAmount			id
// [0x0][0xc783df8a850f42e7F7e57013759C285caa701eB6][1000000000][0 0 0 0 0 0 0 1]
func GetOutgoingTxPoolKey(fee InternalERC20Token, id uint64) []byte {
	amount := make([]byte, 32)
	amount = fee.Amount.BigInt().FillBytes(amount)
	return AppendBytes(OutgoingTXPoolKey, fee.Contract.GetAddress().Bytes(), amount, UInt64Bytes(id))
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
// TODO this should be a sdk.ValAddress
func GetBatchConfirmKey(tokenContract EthAddress, batchNonce uint64, validator sdk.AccAddress) []byte {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		panic(sdkerrors.Wrap(err, "invalid validator address"))
	}
	return AppendBytes(GetBatchConfirmNonceContractPrefix(tokenContract, batchNonce), validator.Bytes())
}

// GetLastEventNonceByValidatorKey indexes latest event nonce by validator
// GetLastEventNonceByValidatorKey returns the following key format
// prefix              cosmos-validator
// [0x0][gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm]
func GetLastEventNonceByValidatorKey(validator sdk.ValAddress) []byte {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		panic(sdkerrors.Wrap(err, "invalid validator address"))
	}
	return AppendBytes(LastEventNonceByValidatorKey, validator.Bytes())
}

func GetDenomToERC20Key(denom string) []byte {
	return AppendBytes(DenomToERC20Key, []byte(denom))
}

func GetERC20ToDenomKey(erc20 EthAddress) []byte {
	return AppendBytes(ERC20ToDenomKey, erc20.GetAddress().Bytes())
}

func GetOutgoingLogicCallKey(invalidationId []byte, invalidationNonce uint64) []byte {
	return AppendBytes(KeyOutgoingLogicCall, invalidationId, UInt64Bytes(invalidationNonce))
}

func GetLogicConfirmNonceInvalidationIdPrefix(invalidationId []byte, invalidationNonce uint64) []byte {
	return AppendBytes(KeyOutgoingLogicConfirm, invalidationId, UInt64Bytes(invalidationNonce))
}

func GetLogicConfirmKey(invalidationId []byte, invalidationNonce uint64, validator sdk.AccAddress) []byte {
	if err := sdk.VerifyAddressFormat(validator); err != nil {
		panic(sdkerrors.Wrap(err, "invalid validator address"))
	}
	return AppendBytes(GetLogicConfirmNonceInvalidationIdPrefix(invalidationId, invalidationNonce), validator.Bytes())
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

// GetPendingIbcAutoForwardKey returns the following key format
// prefix		EventNonce
// [0x0][0 0 0 0 0 0 0 1]
func GetPendingIbcAutoForwardKey(eventNonce uint64) []byte {
	return AppendBytes(PendingIbcAutoForwards, UInt64Bytes(eventNonce))
}
