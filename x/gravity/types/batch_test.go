package types

import (
	"encoding/hex"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutgoingTxBatchCheckpointGold1 tests an outgoing tx batch checkpoint
// nolint: exhaustruct
func TestOutgoingTxBatchCheckpoint(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("paloma", "pub")
	config.SetBech32PrefixForValidator("palomavaloper", "valoperpub")

	src := OutgoingTxBatch{
		BatchNonce:   10,
		BatchTimeout: 1693598120,
		Transactions: []OutgoingTransferTx{
			{
				Id:          4,
				Sender:      "paloma1rxdhpk85wju9z9kqf6m0wq0rkty7gpjhey4wd2",
				DestAddress: "0xE3cD54d29CBf35648EDcf53D6a344bd4B88DA059",
				Erc20Token: ERC20Token{
					Amount:           math.NewInt(10000000),
					Contract:         "0x28E9e9bfedEd29747FCc33ccA25b4B75f05E434B",
					ChainReferenceId: "bnb-main",
				},
			},
		},
		TokenContract: "0x28E9e9bfedEd29747FCc33ccA25b4B75f05E434B",
	}

	actualHash, err := src.GetCheckpoint("5270")
	require.NoError(t, err)

	actualHashHex := hex.EncodeToString(actualHash)
	// hash from bridge contract
	expectedHash := "0xcab79faf47d1556c5b273afb063ab1889461e35bc6403e410af240cefdb0fd8e"[2:]

	assert.Equal(t, expectedHash, actualHashHex)
}
