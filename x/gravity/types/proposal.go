package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// ProposalTypeCommunityPoolEthereumSpend defines the type for a CommunityPoolEthereumSpendProposal
	ProposalTypeCommunityPoolEthereumSpend = "CommunityPoolEthereumSpend"
)

// Assert CommunityPoolEthereumSpendProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &CommunityPoolEthereumSpendProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeCommunityPoolEthereumSpend)
}

// NewCommunityPoolEthereumSpendProposal creates a new community pool spend proposal.
//
//nolint:interfacer
func NewCommunityPoolEthereumSpendProposal(title, description string, recipient string, amount sdk.Coin, bridgeFee sdk.Coin) *CommunityPoolEthereumSpendProposal {
	return &CommunityPoolEthereumSpendProposal{title, description, recipient, amount, bridgeFee}
}

// GetTitle returns the title of a community pool Ethereum spend proposal.
func (csp *CommunityPoolEthereumSpendProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool Ethereum spend proposal.
func (csp *CommunityPoolEthereumSpendProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool Ethereum spend proposal.
func (csp *CommunityPoolEthereumSpendProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool Ethereum spend proposal.
func (csp *CommunityPoolEthereumSpendProposal) ProposalType() string {
	return ProposalTypeCommunityPoolEthereumSpend
}

// ValidateBasic runs basic stateless validity checks
func (csp *CommunityPoolEthereumSpendProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(csp)
	if err != nil {
		return err
	}

	if !common.IsHexAddress(csp.Recipient) {
		return ErrInvalidEthereumProposalRecipient
	}

	if !csp.Amount.IsValid() || csp.Amount.IsZero() {
		return ErrInvalidEthereumProposalAmount
	}

	if !csp.BridgeFee.IsValid() {
		return ErrInvalidEthereumProposalBridgeFee
	}

	if csp.Amount.Denom != csp.BridgeFee.Denom {
		return ErrEthereumProposalDenomMismatch
	}

	return nil
}

// String implements the Stringer interface.
func (csp CommunityPoolEthereumSpendProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Spend Proposal:
  Title:       %s
  Description: %s
  Recipient:   %s
  Amount:      %s
  Bridge Fee:  %s
`, csp.Title, csp.Description, csp.Recipient, csp.Amount, csp.BridgeFee))
	return b.String()
}
