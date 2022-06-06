package types

import (
	"strings"

	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const TypeMsgSubmitNewJob = "submit_new_job"

var _ sdk.Msg = &MsgSubmitNewJob{}

func NewMsgSubmitNewJob(creator string) *MsgSubmitNewJob {
	return &MsgSubmitNewJob{
		Creator: creator,
	}
}

func (msg *MsgSubmitNewJob) Route() string {
	return RouterKey
}

func (msg *MsgSubmitNewJob) Type() string {
	return TypeMsgSubmitNewJob
}

func (msg *MsgSubmitNewJob) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSubmitNewJob) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitNewJob) ValidateBasic() error {
	if msg == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "MsgSubmitNewJob can't be empty")
	}
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.Method == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "cannot call the constructor method")
	}

	// validate EVM part of the message
	if !common.IsHexAddress(msg.HexSmartContractAddress) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "smart contract address is not valid")
	}

	evmabi, err := abi.JSON(strings.NewReader(msg.Abi))
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid evm JSON (%s)", err)
	}

	bz, err := hex.DecodeString(msg.HexPayload)

	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid hex payload (%s)", err)
	}

	method, err := evmabi.MethodById(bz)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid hex payload. Couldn't find method (%s)", err)
	}

	// first 4 bytes are the method id
	bz = bz[4:]

	_, err = method.Inputs.Unpack(bz)

	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "couldn't unpack payload into method arguments (%s)", err)
	}

	return nil
}
