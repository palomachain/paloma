package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValAddressConversion(t *testing.T) {
	codec := address.Bech32Codec{
		Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
	}

	validators := testutil.GenValidators(1, 25000)
	bech32, err := ValAddressFromBech32(codec, validators[0].GetOperator())
	assert.NoError(t, err)
	assert.NotNil(t, bech32)
}
