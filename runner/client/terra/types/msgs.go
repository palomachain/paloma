package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/market module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/staking and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

const (
	EnforcedMaxContractMsgSize = uint64(20 * 1024) // 10KB
)

// GetSigners implements sdk.Msg
func (msg MsgExecuteContract) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{sender}
}

// ValidateBasic implements sdk.Msg
func (msg MsgExecuteContract) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid contract address (%s)", err)
	}

	if !msg.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Coins.String())
	}

	if uint64(len(msg.ExecuteMsg)) > EnforcedMaxContractMsgSize {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "wasm msg byte size is too huge")
	}

	if !json.Valid(msg.ExecuteMsg) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "wasm msg byte format is invalid json")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgExecuteContract) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
