package keeper

import (
	"cosmossdk.io/core/address"
	"github.com/VolumeFi/whoops"
)

const (
	ErrNotFound = whoops.Errorf("item (%T) not found in store: %s")
)

func ConvertStringToBytes(valAddr string) ([]byte, error) {
	var addressCodec address.Codec
	bz, err := addressCodec.StringToBytes(valAddr)
	return bz, err
}
