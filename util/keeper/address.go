package keeper

import (
	"errors"

	"cosmossdk.io/core/address"
)

type err string

const ErrFailedToParseValAddress err = "failed to parse validator address from bech32"

func (e err) Error() string {
	return string(e)
}
func ValAddressFromBech32(addressCodec address.Codec, valAddr string) ([]byte, error) {
	bz, err := addressCodec.StringToBytes(valAddr)
	if err != nil {
		return nil, errors.Join(ErrFailedToParseValAddress, err)
	}
	return bz, nil
}
