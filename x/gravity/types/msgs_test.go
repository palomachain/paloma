package types

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

// Gets the ClaimHash() output from every claims member and casts it to a string
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

// Ensures that ClaimHash changes when members of MsgSendToPalomaClaim change
// The only field which MUST NOT affect ClaimHash is Orchestrator
func TestMsgSendToPalomaClaimHash(t *testing.T) {
	base := MsgSendToPalomaClaim{
		EventNonce:     0,
		EthBlockHeight: 0,
		TokenContract:  "",
		Amount:         math.Int{},
		EthereumSender: "",
		PalomaReceiver: "",
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
	mRecv.PalomaReceiver = NonemptySdkAccAddress().String()

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
