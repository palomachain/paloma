package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/libmeta"
)

var _ sdk.Msg = &MsgSetRelayerFee{}

// GetSigners implements types.Msg.
func (m *MsgSetRelayerFee) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(m)
}

// ValidateBasic implements types.Msg.
func (m *MsgSetRelayerFee) ValidateBasic() error {
	return libmeta.ValidateBasic(m)
}
