package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/testutil"
	"github.com/stretchr/testify/assert"
)

func TestValAddressConversion(t *testing.T) {
	codec := address.Bech32Codec{
		Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
	}

	validators := testutil.GenValidators(2, 25000)
	bech32, err := ValAddressFromBech32(codec, validators[0].GetOperator())
	actual, _ := codec.StringToBytes(validators[0].GetOperator())
	different, _ := ValAddressFromBech32(codec, validators[1].GetOperator())
	assert.NoError(t, err)
	assert.NotNil(t, bech32)
	assert.Equal(t, bech32, actual)
	assert.NotEqual(t, bech32, different)
}
