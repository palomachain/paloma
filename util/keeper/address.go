package keeper

import (
	"errors"

	"cosmossdk.io/core/address"
	"github.com/palomachain/paloma/util/liberr"
)

const ErrFailedToParseValAddress = liberr.Error("failed to parse validator address from bech32: %s")

func ValAddressFromBech32(addressCodec address.Codec, valAddr string) ([]byte, error) {
	bz, err := addressCodec.StringToBytes(valAddr)
	if err != nil {
		return nil, errors.Join(ErrFailedToParseValAddress, err)
	}
	return bz, nil
}
