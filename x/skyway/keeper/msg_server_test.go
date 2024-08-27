package keeper

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"unicode"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/skyway/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testInitStruct struct {
	privKey    *ecdsa.PrivateKey
	ethAddress string
}

func TestConfirmHandlerCommon(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	defer func() {
		sdk.UnwrapSDKContext(ctx).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	batch := types.OutgoingTxBatch{
		BatchNonce:         0,
		BatchTimeout:       420,
		Transactions:       []types.OutgoingTransferTx{},
		TokenContract:      "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
		PalomaBlockCreated: 0,
		ChainReferenceId:   "test-chain",
	}

	checkpoint, err := batch.GetCheckpoint("test-turnstone-id")
	require.NoError(t, err)

	ethSignature, err := types.NewEthereumSignature(checkpoint, EthPrivKeys[0])
	require.NoError(t, err)

	sv := msgServer{input.SkywayKeeper}
	err = sv.confirmHandlerCommon(input.Context, EthAddrs[0].Hex(), AccAddrs[0], hex.EncodeToString(ethSignature), checkpoint, "test-chain")
	require.NoError(t, err)
}

func confirmHandlerCommonWithAddress(t *testing.T, address string, testVar testInitStruct) error {
	input, ctx := SetupFiveValChain(t)
	defer func() {
		sdk.UnwrapSDKContext(ctx).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	privKey := testVar.privKey

	batch := types.OutgoingTxBatch{
		BatchNonce:         0,
		BatchTimeout:       420,
		Transactions:       []types.OutgoingTransferTx{},
		TokenContract:      "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
		PalomaBlockCreated: 0,
		ChainReferenceId:   "test-chain",
	}

	checkpoint, err := batch.GetCheckpoint("test-turnstone-id")
	require.NoError(t, err)

	ethSignature, err := types.NewEthereumSignature(checkpoint, privKey)
	require.NoError(t, err)

	sv := msgServer{input.SkywayKeeper}

	err = sv.confirmHandlerCommon(input.Context, address, AccAddrs[0], hex.EncodeToString(ethSignature), checkpoint, "test-chain")

	return err
}

func TestConfirmHandlerCommonWithLowercaseAddress(t *testing.T) {
	initVar := testInitStruct{privKey: EthPrivKeys[0], ethAddress: EthAddrs[0].String()}

	ret_err := confirmHandlerCommonWithAddress(t, strings.ToLower(EthAddrs[0].String()), initVar)
	assert.Nil(t, ret_err)
}

func TestConfirmHandlerCommonWithUppercaseAddress(t *testing.T) {
	initVar := testInitStruct{privKey: EthPrivKeys[0], ethAddress: EthAddrs[0].String()}

	ret_err := confirmHandlerCommonWithAddress(t, strings.ToUpper(EthAddrs[0].String()), initVar)
	assert.Nil(t, ret_err)
}

func TestConfirmHandlerCommonWithMixedCaseAddress(t *testing.T) {
	initVar := testInitStruct{privKey: EthPrivKeys[0], ethAddress: EthAddrs[0].String()}

	mixedCase := []rune(EthAddrs[0].Hex())
	for i := range mixedCase {
		if rand.Float64() > 0.5 {
			mixedCase[i] = unicode.ToLower(mixedCase[i])
		} else {
			mixedCase[i] = unicode.ToUpper(mixedCase[i])
		}
	}

	ret_err := confirmHandlerCommonWithAddress(t, string(mixedCase), initVar)
	assert.Nil(t, ret_err)
}

func TestOverrideNonceProposal(t *testing.T) {
	input, ctx := SetupFiveValChain(t)
	defer func() {
		sdk.UnwrapSDKContext(ctx).Logger().Info("Asserting invariants at test end")
		input.AssertInvariants()
	}()

	proposal := types.MsgNonceOverrideProposal{
		Metadata: valsettypes.MsgMetadata{
			Creator: input.SkywayKeeper.authority,
			Signers: []string{input.SkywayKeeper.authority},
		},
		ChainReferenceId: "test-chain",
		Nonce:            3,
	}

	err := input.SkywayKeeper.setLastObservedSkywayNonce(ctx, "test-chain", 15)
	require.NoError(t, err)

	lastObserved, err := input.SkywayKeeper.GetLastObservedSkywayNonce(ctx, "test-chain")
	require.NoError(t, err)
	require.Equal(t, uint64(15), lastObserved)

	for k, v := range map[int]uint64{1: 15, 2: 14, 3: 13} {
		err = input.SkywayKeeper.SetLastSkywayNonceByValidator(ctx, sdk.ValAddress(fmt.Sprintf("validator-%d", k)), "test-chain", v)
		require.NoError(t, err)
		lastObserved, err = input.SkywayKeeper.GetLastSkywayNonceByValidator(ctx, sdk.ValAddress(fmt.Sprintf("validator-%d", k)), "test-chain")
		require.NoError(t, err)
		require.Equal(t, v, lastObserved)
	}

	sv := msgServer{input.SkywayKeeper}
	_, err = sv.OverrideNonceProposal(ctx, &proposal)
	require.NoError(t, err)

	lastObserved, err = input.SkywayKeeper.GetLastObservedSkywayNonce(ctx, "test-chain")
	require.NoError(t, err)
	require.Equal(t, proposal.Nonce, lastObserved)

	for k := range map[int]uint64{1: 15, 2: 14, 3: 13} {
		lastObserved, err = input.SkywayKeeper.GetLastSkywayNonceByValidator(ctx, sdk.ValAddress(fmt.Sprintf("validator-%d", k)), "test-chain")
		require.NoError(t, err)
		require.Equal(t, uint64(0), lastObserved, "failed with validator %d", k)
	}
}
