package cli

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"
)

func TestParseCommunityPoolEthereumSpendProposal(t *testing.T) {
	encodingConfig := moduletestutil.MakeTestEncodingConfig()

	okJSON := testutil.WriteToNewTempFile(t, `
{
"title": "Community Pool Ethereum Spend",
"description": "Bridge me some tokens to Ethereum!",
"recipient": "0x0000000000000000000000000000000000000000",
"amount": "20000ugrain",
"bridge_fee": "1000ugrain",
"deposit": "1000ugrain"
}
`)

	proposal, err := ParseCommunityPoolEthereumSpendProposal(encodingConfig.Codec, okJSON.Name())
	require.NoError(t, err)

	require.Equal(t, "Community Pool Ethereum Spend", proposal.Title)
	require.Equal(t, "Bridge me some tokens to Ethereum!", proposal.Description)
	require.Equal(t, "0x0000000000000000000000000000000000000000", proposal.Recipient)
	require.Equal(t, "20000ugrain", proposal.Amount)
	require.Equal(t, "1000ugrain", proposal.BridgeFee)
	require.Equal(t, "1000ugrain", proposal.Deposit)
}
