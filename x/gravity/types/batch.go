package types

import (
	"fmt"
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func (o OutgoingTransferTx) ToInternal() (*InternalOutgoingTransferTx, error) {
	return NewInternalOutgoingTransferTx(o.Id, o.Sender, o.DestAddress, o.Erc20Token)
}

// InternalOutgoingTransferTx is an internal duplicate of OutgoingTransferTx with validation
type InternalOutgoingTransferTx struct {
	Id          uint64
	Sender      sdk.AccAddress
	DestAddress *EthAddress
	Erc20Token  *InternalERC20Token
}

func NewInternalOutgoingTransferTx(
	id uint64,
	sender string,
	destAddress string,
	erc20Token ERC20Token,
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

	return &InternalOutgoingTransferTx{
		Id:          id,
		Sender:      send,
		DestAddress: dest,
		Erc20Token:  token,
	}, nil
}

func (i InternalOutgoingTransferTx) ToExternal() OutgoingTransferTx {
	return OutgoingTransferTx{
		Id:          i.Id,
		Sender:      i.Sender.String(),
		DestAddress: i.DestAddress.GetAddress().Hex(),
		Erc20Token:  i.Erc20Token.ToExternal(),
	}
}

func (i InternalOutgoingTransferTx) ValidateBasic() error {
	err := i.DestAddress.ValidateBasic()
	if err != nil {
		return sdkerrors.Wrap(err, "invalid DestAddress")
	}
	err = i.Erc20Token.ValidateBasic()
	if err != nil {
		return sdkerrors.Wrap(err, "invalid Erc20Token")
	}
	return nil
}

// InternalOutgoingTxBatches is an internal duplicate array of OutgoingTxBatch with validation
type InternalOutgoingTxBatches []InternalOutgoingTxBatch

// InternalOutgoingTxBatch is an internal duplicate of OutgoingTxBatch with validation
type InternalOutgoingTxBatch struct {
	BatchNonce         uint64
	BatchTimeout       uint64
	Transactions       []*InternalOutgoingTransferTx
	TokenContract      EthAddress
	PalomaBlockCreated uint64
	ChainReferenceID   string
	BytesToSign        []byte
	Assignee           string
}

func NewInternalOutgingTxBatch(
	nonce uint64,
	timeout uint64,
	transactions []*InternalOutgoingTransferTx,
	contract EthAddress,
	blockCreated uint64,
	chainReferenceID string,
	turnstoneID string,
	assignee string,
) (*InternalOutgoingTxBatch, error) {
	ret := &InternalOutgoingTxBatch{
		BatchNonce:         nonce,
		BatchTimeout:       timeout,
		Transactions:       transactions,
		TokenContract:      contract,
		PalomaBlockCreated: blockCreated,
		ChainReferenceID:   chainReferenceID,
		Assignee:           assignee,
	}
	bytesToSign, err := ret.GetCheckpoint(turnstoneID)
	if err != nil {
		return nil, err
	}
	ret.BytesToSign = bytesToSign

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

	intBatch := InternalOutgoingTxBatch{
		BatchNonce:         batch.BatchNonce,
		BatchTimeout:       batch.BatchTimeout,
		Transactions:       txs,
		TokenContract:      *contractAddr,
		PalomaBlockCreated: batch.PalomaBlockCreated,
		ChainReferenceID:   batch.ChainReferenceId,
		BytesToSign:        batch.BytesToSign,
		Assignee:           batch.Assignee,
	}
	return &intBatch, nil
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
		PalomaBlockCreated: i.PalomaBlockCreated,
		ChainReferenceId:   i.ChainReferenceID,
		BytesToSign:        i.BytesToSign,
		Assignee:           i.Assignee,
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
			PalomaBlockCreated: val.PalomaBlockCreated,
			ChainReferenceId:   val.ChainReferenceID,
			BytesToSign:        val.BytesToSign,
			Assignee:           val.Assignee,
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

// GetChainReferenceID is required for EthereumSigned interface
func (o OutgoingTxBatch) GetChainReferenceID() string {
	return o.ChainReferenceId
}

// GetCheckpoint is required for EthereumSigned interface
func (o OutgoingTxBatch) GetCheckpoint(turnstoneID string) ([]byte, error) {
	i, err := o.ToInternal()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid OutgoingTxBatch")
	}
	return i.GetCheckpoint(turnstoneID)
}

func (i InternalOutgoingTxBatch) GetChainReferenceID() string {
	return i.ChainReferenceID
}

// GetCheckpoint gets the checkpoint signature from the given outgoing tx batch
func (i InternalOutgoingTxBatch) GetCheckpoint(turnstoneID string) ([]byte, error) {
	arguments := abi.Arguments{
		// token
		{Type: whoops.Must(abi.NewType("address", "", nil))},
		// args
		{Type: whoops.Must(abi.NewType("tuple", "", []abi.ArgumentMarshaling{
			{Name: "receiver", Type: "address[]"},
			{Name: "amount", Type: "uint256[]"},
		}))},
		// batch Nonce
		{Type: whoops.Must(abi.NewType("uint256", "", nil))},
		// turnstone id
		{Type: whoops.Must(abi.NewType("bytes32", "", nil))},
		// deadline
		{Type: whoops.Must(abi.NewType("uint256", "", nil))},
	}

	var turnstoneBytes32 [32]byte
	copy(turnstoneBytes32[:], turnstoneID)

	method := abi.NewMethod("batch_call", "batch_call", abi.Function, "", false, false, arguments, abi.Arguments{})

	methodNameBytes := []uint8("batch_call")
	var batchMethodName [32]uint8
	copy(batchMethodName[:], methodNameBytes)

	// Run through the elements of the batch and serialize them
	txAmounts := make([]*big.Int, len(i.Transactions))
	txDestinations := make([]gethcommon.Address, len(i.Transactions))
	for j, tx := range i.Transactions {
		txAmounts[j] = tx.Erc20Token.Amount.BigInt()
		txDestinations[j] = tx.DestAddress.GetAddress()
	}

	args := struct {
		Receiver []gethcommon.Address
		Amount   []*big.Int
	}{
		Receiver: txDestinations,
		Amount:   txAmounts,
	}

	abiEncodedBatch, err := arguments.Pack(
		i.TokenContract.GetAddress(),
		args,
		big.NewInt(int64(i.BatchNonce)),
		turnstoneBytes32,
		big.NewInt(int64(i.BatchTimeout)),
	)
	// this should never happen outside of test since any case that could crash on encoding
	// should be filtered above.
	if err != nil {
		return nil, sdkerrors.Wrap(err, fmt.Sprintf("Error packing checkpoint! %s/n", err))
	}

	abiEncodedBatch = append(method.ID[:], abiEncodedBatch...)

	return crypto.Keccak256(abiEncodedBatch), nil
}
