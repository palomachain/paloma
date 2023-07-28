package types

import (
	"bytes"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestValidateMsgSetOrchestratorAddress(t *testing.T) {
	var (
		ethAddress                   = "0xb462864E395d88d6bc7C5dd5F3F5eb4cc2599255"
		cosmosAddress sdk.AccAddress = bytes.Repeat([]byte{0x1}, 20)
		valAddress    sdk.ValAddress = bytes.Repeat([]byte{0x1}, 20)
	)
	specs := map[string]struct {
		srcCosmosAddr sdk.AccAddress
		srcValAddr    sdk.ValAddress
		srcETHAddr    string
		expErr        bool
	}{
		"all good": {
			srcCosmosAddr: cosmosAddress,
			srcValAddr:    valAddress,
			srcETHAddr:    ethAddress,
		},
		"empty validator address": {
			srcETHAddr:    ethAddress,
			srcCosmosAddr: cosmosAddress,
			expErr:        true,
		},
		"short validator address": {
			srcValAddr:    []byte{0x1},
			srcCosmosAddr: cosmosAddress,
			srcETHAddr:    ethAddress,
			expErr:        false,
		},
		"empty cosmos address": {
			srcValAddr: valAddress,
			srcETHAddr: ethAddress,
			expErr:     true,
		},
		"short cosmos address": {
			srcCosmosAddr: []byte{0x1},
			srcValAddr:    valAddress,
			srcETHAddr:    ethAddress,
			expErr:        false,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			println(fmt.Sprintf("Spec is %v", msg))
			ethAddr, err := NewEthAddress(spec.srcETHAddr)
			assert.NoError(t, err)
			msg := NewMsgSetOrchestratorAddress(spec.srcValAddr, spec.srcCosmosAddr, *ethAddr)
			// when
			err = msg.ValidateBasic()
			if spec.expErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}

}

// Gets the ClaimHash() output from every claims member and casts it to a string, panicing on any errors
func getClaimHashStrings(t *testing.T, claims ...EthereumClaim) (hashes []string) {
	for _, claim := range claims {
		hash, e := claim.ClaimHash()
		require.NoError(t, e)
		hashes = append(hashes, string(hash))
	}
	return
}

// Calls SetOrchestrator on every claims member, passing orch as the value
func setOrchestratorOnClaims(orch sdk.AccAddress, claims ...EthereumClaim) (ret []EthereumClaim) {
	for _, claim := range claims {
		clam := claim
		clam.SetOrchestrator(orch)
		ret = append(ret, clam)
	}
	return
}

// Ensures that ClaimHash changes when members of MsgSendToCosmosClaim change
// The only field which MUST NOT affect ClaimHash is Orchestrator
func TestMsgSendToCosmosClaimHash(t *testing.T) {
	base := MsgSendToCosmosClaim{
		EventNonce:     0,
		EthBlockHeight: 0,
		TokenContract:  "",
		Amount:         sdk.Int{},
		EthereumSender: "",
		CosmosReceiver: "",
		Orchestrator:   "",
	}

	// Copy and populate base with values, saving orchestrator for a special check
	orchestrator := NonemptySdkAccAddress()
	mNonce := base
	mNonce.EventNonce = NonzeroUint64()
	mBlock := base
	mBlock.EthBlockHeight = NonzeroUint64()
	mCtr := base
	mCtr.TokenContract = NonemptyEthAddress()
	mAmt := base
	mAmt.Amount = NonzeroSdkInt()
	mSend := base
	mSend.EthereumSender = NonemptyEthAddress()
	mRecv := base
	mRecv.CosmosReceiver = NonemptySdkAccAddress().String()

	hashes := getClaimHashStrings(t, &base, &mNonce, &mBlock, &mCtr, &mAmt, &mSend, &mRecv)
	baseH := hashes[0]
	rest := hashes[1:]
	// Assert that the base claim hash differs from all the rest
	require.False(t, slices.Contains(rest, baseH))

	newClaims := setOrchestratorOnClaims(orchestrator, &base, &mNonce, &mBlock, &mCtr, &mAmt, &mSend, &mRecv)
	newHashes := getClaimHashStrings(t, newClaims...)
	// Assert that the claims with orchestrator set do not change the hashes
	require.Equal(t, hashes, newHashes)
}

// Ensures that ClaimHash changes when members of MsgBatchSendToEth change
// The only field which MUST NOT affect ClaimHash is Orchestrator
func TestMsgBatchSendToEthClaimHash(t *testing.T) {
	base := MsgBatchSendToEthClaim{
		EventNonce:     0,
		EthBlockHeight: 0,
		BatchNonce:     0,
		TokenContract:  "",
		Orchestrator:   "",
	}

	orchestrator := NonemptySdkAccAddress()
	mNonce := base
	mNonce.EventNonce = NonzeroUint64()
	mBlock := base
	mBlock.EthBlockHeight = NonzeroUint64()
	mBatch := base
	mBatch.BatchNonce = NonzeroUint64()
	mCtr := base
	mCtr.TokenContract = NonemptyEthAddress()

	hashes := getClaimHashStrings(t, &base, &mNonce, &mBlock, &mBatch, &mCtr)
	baseH := hashes[0]
	rest := hashes[1:]
	// Assert that the base claim hash differs from all the rest
	require.False(t, slices.Contains(rest, baseH))

	newClaims := setOrchestratorOnClaims(orchestrator, &base, &mNonce, &mBlock, &mBatch, &mCtr)
	newHashes := getClaimHashStrings(t, newClaims...)
	// Assert that the claims with orchestrator set do not change the hashes
	require.Equal(t, hashes, newHashes)
}

// Ensures that ClaimHash changes when members of MsgERC20DeployedClaim change
// The only field which MUST NOT affect ClaimHash is Orchestrator
func TestMsgERC20DeployedClaimHash(t *testing.T) {
	base := MsgERC20DeployedClaim{
		EventNonce:     0,
		EthBlockHeight: 0,
		CosmosDenom:    "",
		TokenContract:  "",
		Name:           "",
		Symbol:         "",
		Decimals:       0,
		Orchestrator:   "",
	}

	orchestrator := NonemptySdkAccAddress()
	mNonce := base
	mNonce.EventNonce = NonzeroUint64()
	mBlock := base
	mBlock.EthBlockHeight = NonzeroUint64()
	mDenom := base
	mDenom.CosmosDenom = NonemptyEthAddress()
	mCtr := base
	mCtr.TokenContract = NonemptyEthAddress()
	mName := base
	mName.Name = NonemptyEthAddress()
	mSymb := base
	mSymb.Symbol = NonemptyEthAddress()
	mDecim := base
	mDecim.Decimals = NonzeroUint64()

	hashes := getClaimHashStrings(t, &base, &mNonce, &mBlock, &mDenom, &mName, &mSymb, &mDecim)
	baseH := hashes[0]
	rest := hashes[1:]
	// Assert that the base claim hash differs from all the rest
	require.False(t, slices.Contains(rest, baseH))

	newClaims := setOrchestratorOnClaims(orchestrator, &base, &mNonce, &mBlock, &mDenom, &mName, &mSymb, &mDecim)
	newHashes := getClaimHashStrings(t, newClaims...)
	// Assert that the claims with orchestrator set do not change the hashes
	require.Equal(t, hashes, newHashes)
}

// Ensures that ClaimHash changes when members of MsgLogicCallExecutedClaim change
// The only field which MUST NOT affect ClaimHash is Orchestrator
func TestMsgLogicCallExecutedClaimHash(t *testing.T) {
	base := MsgLogicCallExecutedClaim{
		EventNonce:        0,
		EthBlockHeight:    0,
		InvalidationId:    []byte{},
		InvalidationNonce: 0,
		Orchestrator:      "",
	}

	orchestrator := NonemptySdkAccAddress()
	mNonce := base
	mNonce.EventNonce = NonzeroUint64()
	mBlock := base
	mBlock.EthBlockHeight = NonzeroUint64()
	mInvId := base
	mInvId.InvalidationId = NonemptySdkAccAddress().Bytes()
	mInvNo := base
	mInvNo.InvalidationNonce = NonzeroUint64()

	hashes := getClaimHashStrings(t, &base, &mNonce, &mBlock, &mInvId, &mInvNo)
	baseH := hashes[0]
	rest := hashes[1:]
	// Assert that the base claim hash differs from all the rest
	require.False(t, slices.Contains(rest, baseH))

	newClaims := setOrchestratorOnClaims(orchestrator, &base, &mNonce, &mBlock, &mInvId, &mInvNo)
	newHashes := getClaimHashStrings(t, newClaims...)
	// Assert that the claims with orchestrator set do not change the hashes
	require.Equal(t, hashes, newHashes)
}
