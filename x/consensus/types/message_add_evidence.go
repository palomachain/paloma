package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAddEvidence = "add_evidence"

var _ sdk.Msg = &MsgAddEvidence{}

func NewMsgAddEvidence(creator string, messageID string, signature string, publicKey string, evidence string) *MsgAddEvidence {
	return &MsgAddEvidence{
		Creator:   creator,
		MessageID: messageID,
		Signature: signature,
		PublicKey: publicKey,
		Evidence:  evidence,
	}
}

func (msg *MsgAddEvidence) Route() string {
	return RouterKey
}

func (msg *MsgAddEvidence) Type() string {
	return TypeMsgAddEvidence
}

func (msg *MsgAddEvidence) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddEvidence) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddEvidence) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
