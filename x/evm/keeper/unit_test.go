package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestGeneratingNewUniqueID(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{
		Height: 55,
	}, false, nil)

	res := generateSmartContractID(ctx)
	var expected [32]byte

	expected[0] = '5'
	expected[1] = '5'
	require.Equal(t, expected, res)
}
