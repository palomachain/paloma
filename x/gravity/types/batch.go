package types

import (
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func (o OutgoingTransferTx) ToInternal() (*InternalOutgoingTransferTx, error) {
	return NewInternalOutgoingTransferTx(o.Id, o.Sender, o.DestAddress, o.Erc20Token, o.Erc20Fee)
}

// InternalOutgoingTransferTx is an internal duplicate of OutgoingTransferTx with validation
type InternalOutgoingTransferTx struct {
	Id          uint64
	Sender      sdk.AccAddress
	DestAddress *EthAddress
	Erc20Token  *InternalERC20Token
	Erc20Fee    *InternalERC20Token
}

func NewInternalOutgoingTransferTx(
	id uint64,
	sender string,
	destAddress string,
	erc20Token ERC20Token,
	erc20Fee ERC20Token,
) (*InternalOutgoingTransferTx, error) {
	send, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid sender")
	}
	dest, err := NewEthAddress(destAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid eth destination")
	}
	token, err := erc20Token.ToInternal()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid Erc20Token")
	}
	fee, err := erc20Fee.ToInternal()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid Erc20Fee")
	}

	return &InternalOutgoingTransferTx{
		Id:          id,
		Sender:      send,
		DestAddress: dest,
		Erc20Token:  token,
		Erc20Fee:    fee,
	}, nil
}

func (i InternalOutgoingTransferTx) ToExternal() OutgoingTransferTx {
	return OutgoingTransferTx{
		Id:          i.Id,
		Sender:      i.Sender.String(),
		DestAddress: i.DestAddress.GetAddress().Hex(),
		Erc20Token:  i.Erc20Token.ToExternal(),
		Erc20Fee:    i.Erc20Fee.ToExternal(),
	}
}

func (i InternalOutgoingTransferTx) ValidateBasic() error {
	//TODO: Validate id?
	//TODO: Validate cosmos sender?
	err := i.DestAddress.ValidateBasic()
	if err != nil {
		return sdkerrors.Wrap(err, "invalid DestAddress")
	}
	err = i.Erc20Token.ValidateBasic()
	if err != nil {
		return sdkerrors.Wrap(err, "invalid Erc20Token")
	}
	err = i.Erc20Fee.ValidateBasic()
	if err != nil {
		return sdkerrors.Wrap(err, "invalid Erc20Fee")
	}
	return nil
}

// InternalOutgoingTxBatch is an internal duplicate array of OutgoingTxBatch with validation
type InternalOutgoingTxBatches []InternalOutgoingTxBatch

// InternalOutgoingTxBatch is an internal duplicate of OutgoingTxBatch with validation
type InternalOutgoingTxBatch struct {
	BatchNonce         uint64
	BatchTimeout       uint64
	Transactions       []*InternalOutgoingTransferTx
	TokenContract      EthAddress
	CosmosBlockCreated uint64
}

func NewInternalOutgingTxBatch(
	nonce uint64,
	timeout uint64,
	transactions []*InternalOutgoingTransferTx,
	contract EthAddress,
	blockCreated uint64) (*InternalOutgoingTxBatch, error) {

	ret := &InternalOutgoingTxBatch{
		BatchNonce:         nonce,
		BatchTimeout:       timeout,
		Transactions:       transactions,
		TokenContract:      contract,
		CosmosBlockCreated: blockCreated,
	}
	if err := ret.ValidateBasic(); err != nil {
		return nil, err
	}
	return ret, nil
}

func NewInternalOutgingTxBatchFromExternalBatch(batch OutgoingTxBatch) (*InternalOutgoingTxBatch, error) {
	contractAddr, err := NewEthAddress(batch.TokenContract)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid eth address")
	}
	txs := make([]*InternalOutgoingTransferTx, len(batch.Transactions))
	for i, tx := range batch.Transactions {
		intTx, err := tx.ToInternal()
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "invalid transaction in batch: %v", tx)
		}
		txs[i] = intTx
	}

	return &InternalOutgoingTxBatch{
		BatchNonce:         batch.BatchNonce,
		BatchTimeout:       batch.BatchTimeout,
		Transactions:       txs,
		TokenContract:      *contractAddr,
		CosmosBlockCreated: batch.CosmosBlockCreated,
	}, nil
}

func (o *OutgoingTxBatch) ToInternal() (*InternalOutgoingTxBatch, error) {
	return NewInternalOutgingTxBatchFromExternalBatch(*o)
}

func (i *InternalOutgoingTxBatch) ToExternal() OutgoingTxBatch {
	txs := make([]OutgoingTransferTx, len(i.Transactions))
	for i, tx := range i.Transactions {
		txs[i] = tx.ToExternal()
	}
	return OutgoingTxBatch{
		BatchNonce:         i.BatchNonce,
		BatchTimeout:       i.BatchTimeout,
		Transactions:       txs,
		TokenContract:      i.TokenContract.GetAddress().Hex(),
		CosmosBlockCreated: i.CosmosBlockCreated,
	}
}

func (i *InternalOutgoingTxBatches) ToExternalArray() []OutgoingTxBatch {
	var arr []OutgoingTxBatch

	for index, val := range *i {
		txs := make([]OutgoingTransferTx, len((*i)[index].Transactions))
		for i, tx := range (*i)[index].Transactions {
			txs[i] = tx.ToExternal()
		}

		arr = append(arr, OutgoingTxBatch{
			BatchNonce:         val.BatchNonce,
			BatchTimeout:       val.BatchTimeout,
			Transactions:       txs,
			TokenContract:      val.TokenContract.GetAddress().Hex(),
			CosmosBlockCreated: val.CosmosBlockCreated,
		})
	}

	return arr
}

func (i *InternalOutgoingTxBatch) ValidateBasic() error {
	if err := i.TokenContract.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "invalid eth address")
	}

	for i, tx := range i.Transactions {
		if err := tx.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "transaction %d is invalid", i)
		}
	}
	return nil
}

// Required for EthereumSigned interface
func (o OutgoingTxBatch) GetCheckpoint(gravityIDstring string) []byte {
	i, err := o.ToInternal()
	if err != nil {
		panic(sdkerrors.Wrap(err, "invalid OutgoingTxBatch"))
	}
	return i.GetCheckpoint(gravityIDstring)
}

// GetCheckpoint gets the checkpoint signature from the given outgoing tx batch
func (i InternalOutgoingTxBatch) GetCheckpoint(gravityIDstring string) []byte {

	abi, err := abi.JSON(strings.NewReader(OutgoingBatchTxCheckpointABIJSON))
	if err != nil {
		panic("Bad ABI constant!")
	}

	// the contract argument is not a arbitrary length array but a fixed length 32 byte
	// array, therefore we have to utf8 encode the string (the default in this case) and
	// then copy the variable length encoded data into a fixed length array. This function
	// will panic if gravityId is too long to fit in 32 bytes
	gravityID, err := strToFixByteArray(gravityIDstring)
	if err != nil {
		panic(err)
	}

	// Create the methodName argument which salts the signature
	methodNameBytes := []uint8("transactionBatch")
	var batchMethodName [32]uint8
	copy(batchMethodName[:], methodNameBytes)

	// Run through the elements of the batch and serialize them
	txAmounts := make([]*big.Int, len(i.Transactions))
	txDestinations := make([]gethcommon.Address, len(i.Transactions))
	txFees := make([]*big.Int, len(i.Transactions))
	for j, tx := range i.Transactions {
		txAmounts[j] = tx.Erc20Token.Amount.BigInt()
		txDestinations[j] = tx.DestAddress.GetAddress()
		txFees[j] = tx.Erc20Fee.Amount.BigInt()
	}

	// the methodName needs to be the same as the 'name' above in the checkpointAbiJson
	// but other than that it's a constant that has no impact on the output. This is because
	// it gets encoded as a function name which we must then discard.
	abiEncodedBatch, err := abi.Pack("submitBatch",
		gravityID,
		batchMethodName,
		txAmounts,
		txDestinations,
		txFees,
		big.NewInt(int64(i.BatchNonce)),
		i.TokenContract.GetAddress(),
		big.NewInt(int64(i.BatchTimeout)),
	)

	// this should never happen outside of test since any case that could crash on encoding
	// should be filtered above.
	if err != nil {
		panic(fmt.Sprintf("Error packing checkpoint! %s/n", err))
	}

	// we hash the resulting encoded bytes discarding the first 4 bytes these 4 bytes are the constant
	// method name 'checkpoint'. If you were to replace the checkpoint constant in this code you would
	// then need to adjust how many bytes you truncate off the front to get the output of abi.encode()
	return crypto.Keccak256Hash(abiEncodedBatch[4:]).Bytes()
}

func (c OutgoingLogicCall) ValidateBasic() error {
	for _, t := range c.Transfers {
		_, err := t.ToInternal() // ToInternal calls ValidateBasic for us
		if err != nil {
			return sdkerrors.Wrapf(ErrInvalidLogicCall, "invalid transfer in logic call: %v", err)
		}
	}
	for _, t := range c.Fees {
		_, err := t.ToInternal() // ToInternal calls ValidateBasic for us
		if err != nil {
			return sdkerrors.Wrapf(ErrInvalidLogicCall, "invalid fee in logic call: %v", err)
		}
	}
	_, err := NewEthAddress(c.LogicContractAddress)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidLogicCall, "invalid logic call logic contract address: %v", err)
	}

	return nil
}

// GetCheckpoint gets the checkpoint signature from the given outgoing tx batch
func (c OutgoingLogicCall) GetCheckpoint(gravityIDstring string) []byte {

	abi, err := abi.JSON(strings.NewReader(OutgoingLogicCallABIJSON))
	if err != nil {
		panic("Bad ABI constant!")
	}

	// Create the methodName argument which salts the signature
	methodNameBytes := []uint8("logicCall")
	var logicCallMethodName [32]uint8
	copy(logicCallMethodName[:], methodNameBytes)

	// the contract argument is not a arbitrary length array but a fixed length 32 byte
	// array, therefore we have to utf8 encode the string (the default in this case) and
	// then copy the variable length encoded data into a fixed length array. This function
	// will panic if gravityId is too long to fit in 32 bytes
	gravityID, err := strToFixByteArray(gravityIDstring)
	if err != nil {
		panic(err)
	}

	// Run through the elements of the logic call and serialize them
	transferAmounts := make([]*big.Int, len(c.Transfers))
	transferTokenContracts := make([]gethcommon.Address, len(c.Transfers))
	feeAmounts := make([]*big.Int, len(c.Fees))
	feeTokenContracts := make([]gethcommon.Address, len(c.Fees))
	for i, tx := range c.Transfers {
		transferAmounts[i] = tx.Amount.BigInt()
		transferTokenContracts[i] = gethcommon.HexToAddress(tx.Contract)
	}
	for i, tx := range c.Fees {
		feeAmounts[i] = tx.Amount.BigInt()
		feeTokenContracts[i] = gethcommon.HexToAddress(tx.Contract)
	}
	payload := make([]byte, len(c.Payload))
	copy(payload, c.Payload)
	var invalidationId [32]byte
	copy(invalidationId[:], c.InvalidationId)

	// the methodName needs to be the same as the 'name' above in the checkpointAbiJson
	// but other than that it's a constant that has no impact on the output. This is because
	// it gets encoded as a function name which we must then discard.
	abiEncodedCall, err := abi.Pack("checkpoint",
		gravityID,
		logicCallMethodName,
		transferAmounts,
		transferTokenContracts,
		feeAmounts,
		feeTokenContracts,
		gethcommon.HexToAddress(c.LogicContractAddress),
		payload,
		big.NewInt(int64(c.Timeout)),
		invalidationId,
		big.NewInt(int64(c.InvalidationNonce)),
	)

	// this should never happen outside of test since any case that could crash on encoding
	// should be filtered above.
	if err != nil {
		panic(fmt.Sprintf("Error packing checkpoint! %s/n", err))
	}

	return crypto.Keccak256Hash(abiEncodedCall[4:]).Bytes()
}
