package libmeta

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const ErrMissingSigners err = "missing signers"

type err string

func (e err) Error() string { return string(e) }

func ValidateBasic[E Metadata, T MsgWithMetadata[E]](msg T) error {
	signers := msg.GetMetadata().GetSigners()
	creator := msg.GetMetadata().GetCreator()

	if len(signers) < 1 {
		return ErrMissingSigners
	}

	for _, v := range append(signers, creator) {
		_, err := sdk.AccAddressFromBech32(v)
		if err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
		}
	}

	return nil
}
