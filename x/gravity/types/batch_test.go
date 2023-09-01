package types

//// TODO: making these test work (with a new hash from compass) will prove useful
//
//import (
//	"encoding/hex"
//	"testing"
//
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//)
//
//// TODO : Come back and fix this test once I have checkpointing done
//// TestOutgoingTxBatchCheckpointGold1 tests an outgoing tx batch checkpoint
//// nolint: exhaustruct
//func TestOutgoingTxBatchCheckpoint(t *testing.T) {
//	senderAddr, err := sdk.AccAddressFromHexUnsafe("527FBEE652609AB150F0AEE9D61A2F76CFC4A73E")
//	require.NoError(t, err)
//	var (
//		erc20Addr = "0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"
//	)
//	erc20Address, err := NewEthAddress(erc20Addr)
//	require.NoError(t, err)
//	destAddress, err := NewEthAddress("0x9FC9C2DfBA3b6cF204C37a5F690619772b926e39")
//	require.NoError(t, err)
//	src := OutgoingTxBatch{
//		BatchNonce:   1,
//		BatchTimeout: 2111,
//		Transactions: []OutgoingTransferTx{
//			{
//				Id:          0x1,
//				Sender:      senderAddr.String(),
//				DestAddress: destAddress.GetAddress().Hex(),
//				Erc20Token: ERC20Token{
//					Amount:           sdk.NewInt(0x1),
//					Contract:         erc20Address.GetAddress().Hex(),
//					ChainReferenceId: "test-chain",
//				},
//			},
//		},
//		TokenContract: erc20Address.GetAddress().Hex(),
//	}
//
//	ourHash := src.GetCheckpoint("foo")
//
//	// hash from bridge contract
//	goldHash := "0xa3a7ee0a363b8ad2514e7ee8f110d7449c0d88f3b0913c28c1751e6e0079a9b2"[2:]
//	// The function used to compute the "gold hash" above is in /solidity/test/updateValsetAndSubmitBatch.ts
//	// Be aware that every time that you run the above .ts file, it will use a different tokenContractAddress and thus compute
//	// a different hash.
//	assert.Equal(t, goldHash, hex.EncodeToString(ourHash))
//}
