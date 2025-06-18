package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errtypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/palomachain/paloma/v2/util/libmeta"
)

const TypeMsgUpdateParams = "add_status_update"

var _ sdk.Msg = &MsgUpdateParams{}

func (msg *MsgUpdateParams) Route() string {
	return RouterKey
}

func (msg *MsgUpdateParams) Type() string {
	return TypeMsgUpdateParams
}

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}

func (msg *MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateParams) ValidateBasic() error {
	if err := libmeta.ValidateBasic(msg); err != nil {
		return err
	}

	if msg.Authority != msg.Metadata.Creator {
		return ErrUnauthorized.Wrapf("authority mismatch: %s", msg.Authority)
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrapf(errtypes.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	return msg.Params.Validate()
}
