package keeper

import (
	"strings"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/require"
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

func TestGeneratingUniqueIDGeneratesDifferentBytecode(t *testing.T) {
	jsonAbi := `[
    {
        "stateMutability": "nonpayable",
        "type": "constructor",
        "inputs": [
            {
                "name": "turnstone_id",
                "type": "bytes32"
            }
        ],
        "outputs": []
    }
]
`

	contractABI, err := abi.JSON(strings.NewReader(jsonAbi))
	require.NoError(t, err)

	uniqueID1 := generateSmartContractID(sdk.Context{}.WithBlockHeight(123))
	uniqueID2 := generateSmartContractID(sdk.Context{}.WithBlockHeight(456))

	require.NotEqual(t, uniqueID1, uniqueID2)

	input1, err := contractABI.Pack("", uniqueID1)
	require.NoError(t, err)

	input2, err := contractABI.Pack("", uniqueID2)
	require.NoError(t, err)

	require.NotEqual(t, input1, input2)
}
