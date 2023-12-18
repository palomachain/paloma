package keeper

import "cosmossdk.io/core/address"

func ValAddressFromBech32(addressCodec address.Codec, valAddr string) ([]byte, error) {
	bz, err := addressCodec.StringToBytes(valAddr)
	return bz, err
}
