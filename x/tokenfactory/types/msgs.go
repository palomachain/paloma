package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypeerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/palomachain/paloma/v2/util/libmeta"
	valsettypes "github.com/palomachain/paloma/v2/x/valset/types"
)

const (
	TypeMsgCreateDenom      = "create_denom"
	TypeMsgMint             = "tf_mint"
	TypeMsgBurn             = "tf_burn"
	TypeMsgChangeAdmin      = "change_admin"
	TypeMsgSetDenomMetadata = "set_denom_metadata"
	TypeMsgUpdateParams     = "update_params"
)

var (
	_ sdk.Msg = &MsgCreateDenom{}
	_ sdk.Msg = &MsgMint{}
	_ sdk.Msg = &MsgBurn{}
	_ sdk.Msg = &MsgChangeAdmin{}
	_ sdk.Msg = &MsgSetDenomMetadata{}
	_ sdk.Msg = &MsgUpdateParams{}
)

func NewMsgCreateDenom(sender, subdenom string) *MsgCreateDenom {
	return &MsgCreateDenom{
		Subdenom: subdenom,
		Metadata: valsettypes.MsgMetadata{
			Creator: sender,
			Signers: []string{sender},
		},
	}
}

func (m MsgCreateDenom) Route() string { return RouterKey }
func (m MsgCreateDenom) Type() string  { return TypeMsgCreateDenom }
func (m MsgCreateDenom) ValidateBasic() error {
	if err := libmeta.ValidateBasic(&m); err != nil {
		return err
	}

	_, err := GetTokenDenom(m.Metadata.Creator, m.Subdenom)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidDenom, err.Error())
	}

	return nil
}

func (m MsgCreateDenom) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgCreateDenom) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&m)
}

func NewMsgMint(sender string, amount sdk.Coin) *MsgMint {
	return &MsgMint{
		Amount: amount,
		Metadata: valsettypes.MsgMetadata{
			Creator: sender,
			Signers: []string{sender},
		},
	}
}

func (m MsgMint) Route() string { return RouterKey }
func (m MsgMint) Type() string  { return TypeMsgMint }
func (m MsgMint) ValidateBasic() error {
	if err := libmeta.ValidateBasic(&m); err != nil {
		return err
	}

	if !m.Amount.IsValid() || m.Amount.Amount.Equal(sdkmath.ZeroInt()) {
		return sdkerrors.Wrap(sdktypeerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

func (m MsgMint) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgMint) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&m)
}

func NewMsgBurn(sender string, amount sdk.Coin) *MsgBurn {
	return &MsgBurn{
		Amount: amount,
		Metadata: valsettypes.MsgMetadata{
			Creator: sender,
			Signers: []string{sender},
		},
	}
}

func (m MsgBurn) Route() string { return RouterKey }
func (m MsgBurn) Type() string  { return TypeMsgBurn }
func (m MsgBurn) ValidateBasic() error {
	if err := libmeta.ValidateBasic(&m); err != nil {
		return err
	}

	if !m.Amount.IsValid() || m.Amount.Amount.Equal(sdkmath.ZeroInt()) {
		return sdkerrors.Wrap(sdktypeerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

func (m MsgBurn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgBurn) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&m)
}

func NewMsgChangeAdmin(sender, denom, newAdmin string) *MsgChangeAdmin {
	return &MsgChangeAdmin{
		Denom:    denom,
		NewAdmin: newAdmin,
		Metadata: valsettypes.MsgMetadata{
			Creator: sender,
			Signers: []string{sender},
		},
	}
}

func (m MsgChangeAdmin) Route() string { return RouterKey }
func (m MsgChangeAdmin) Type() string  { return TypeMsgChangeAdmin }
func (m MsgChangeAdmin) ValidateBasic() error {
	if err := libmeta.ValidateBasic(&m); err != nil {
		return err
	}

	_, _, err := DeconstructDenom(m.Denom)
	if err != nil {
		return err
	}

	return nil
}

func (m MsgChangeAdmin) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgChangeAdmin) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&m)
}

func NewMsgSetDenomMetadata(sender string, metadata banktypes.Metadata) *MsgSetDenomMetadata {
	return &MsgSetDenomMetadata{
		DenomMetadata: metadata,
		Metadata: valsettypes.MsgMetadata{
			Creator: sender,
			Signers: []string{sender},
		},
	}
}

func (m MsgSetDenomMetadata) Route() string { return RouterKey }
func (m MsgSetDenomMetadata) Type() string  { return TypeMsgSetDenomMetadata }
func (m MsgSetDenomMetadata) ValidateBasic() error {
	if err := libmeta.ValidateBasic(&m); err != nil {
		return err
	}

	err := m.DenomMetadata.Validate()
	if err != nil {
		return err
	}

	_, _, err = DeconstructDenom(m.DenomMetadata.Base)
	if err != nil {
		return err
	}

	return nil
}

func (m MsgSetDenomMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgSetDenomMetadata) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&m)
}

func (m MsgUpdateParams) Route() string { return RouterKey }
func (m MsgUpdateParams) Type() string  { return TypeMsgUpdateParams }
func (m MsgUpdateParams) ValidateBasic() error {
	if err := libmeta.ValidateBasic(&m); err != nil {
		return err
	}

	if m.Authority != m.Metadata.Creator {
		return ErrUnauthorized.Wrapf("authority mismatch: %s", m.Authority)
	}

	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.Wrapf(sdktypeerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	return m.Params.Validate()
}

func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(&m)
}
