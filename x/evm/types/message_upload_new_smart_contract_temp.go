package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUploadNewSmartContractTemp = "upload_new_smart_contract_temp"

var _ sdk.Msg = &MsgUploadNewSmartContractTemp{}

func NewMsgUploadNewSmartContractTemp(creator string, abi string, bytecode string, constructorInput string, chainReferenceID string) *MsgUploadNewSmartContractTemp {
	return &MsgUploadNewSmartContractTemp{
		Creator:          creator,
		Abi:              abi,
		Bytecode:         bytecode,
		ConstructorInput: constructorInput,
		// ChainReferenceID:          chainReferenceID,
	}
}

func (msg *MsgUploadNewSmartContractTemp) Route() string {
	return RouterKey
}

func (msg *MsgUploadNewSmartContractTemp) Type() string {
	return TypeMsgUploadNewSmartContractTemp
}

func (msg *MsgUploadNewSmartContractTemp) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUploadNewSmartContractTemp) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUploadNewSmartContractTemp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
