package types

import (
	fmt "fmt"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ProposalTypeUnhaltBridge = "UnhaltBridge"
	ProposalTypeAirdrop      = "Airdrop"
	ProposalTypeIBCMetadata  = "IBCMetadata"
)

func (p *UnhaltBridgeProposal) GetTitle() string { return p.Title }

func (p *UnhaltBridgeProposal) GetDescription() string { return p.Description }

func (p *UnhaltBridgeProposal) ProposalRoute() string { return RouterKey }

func (p *UnhaltBridgeProposal) ProposalType() string {
	return ProposalTypeUnhaltBridge
}

func (p *UnhaltBridgeProposal) ValidateBasic() error {
	err := govv1beta1types.ValidateAbstract(p)
	if err != nil {
		return err
	}
	return nil
}

func (p UnhaltBridgeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Unhalt Bridge Proposal:
  Title:          %s
  Description:    %s
  target_nonce:   %d
`, p.Title, p.Description, p.TargetNonce))
	return b.String()
}

func (p *AirdropProposal) GetTitle() string { return p.Title }

func (p *AirdropProposal) GetDescription() string { return p.Description }

func (p *AirdropProposal) ProposalRoute() string { return RouterKey }

func (p *AirdropProposal) ProposalType() string {
	return ProposalTypeAirdrop
}

func (p *AirdropProposal) ValidateBasic() error {
	err := govv1beta1types.ValidateAbstract(p)
	if err != nil {
		return err
	}
	return nil
}

func (p AirdropProposal) String() string {
	var b strings.Builder
	total := uint64(0)
	for _, v := range p.Amounts {
		total += v
	}
	parsedRecipients := make([]sdk.AccAddress, len(p.Recipients)/20)
	for i := 0; i < len(p.Recipients)/20; i++ {
		indexStart := i * 20
		indexEnd := indexStart + 20
		addr := p.Recipients[indexStart:indexEnd]
		parsedRecipients[i] = addr
	}
	recipients := ""
	for i, a := range parsedRecipients {
		recipients += fmt.Sprintf("Account: %s Amount: %d%s", a.String(), p.Amounts[i], p.Denom)
	}

	b.WriteString(fmt.Sprintf(`Airdrop Proposal:
  Title:          %s
  Description:    %s
  Total Amount:   %d%s
  Recipients:     %s
`, p.Title, p.Description, total, p.Denom, recipients))
	return b.String()
}

func (p *IBCMetadataProposal) GetTitle() string { return p.Title }

func (p *IBCMetadataProposal) GetDescription() string { return p.Description }

func (p *IBCMetadataProposal) ProposalRoute() string { return RouterKey }

func (p *IBCMetadataProposal) ProposalType() string {
	return ProposalTypeIBCMetadata
}

func (p *IBCMetadataProposal) ValidateBasic() error {
	err := govv1beta1types.ValidateAbstract(p)
	if err != nil {
		return err
	}
	return nil
}

func (p IBCMetadataProposal) String() string {
	decimals := uint32(0)
	for _, denomUnit := range p.Metadata.DenomUnits {
		if denomUnit.Denom == p.Metadata.Display {
			decimals = denomUnit.Exponent
			break
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`IBC Metadata setting proposal:
  Title:             %s
  Description:       %s
  Token Name:        %s
  Token Symbol:      %s
  Token Display:     %s
  Token Decimals:    %d
  Token Description: %s
`, p.Title, p.Description, p.Metadata.Name, p.Metadata.Symbol, p.Metadata.Display, decimals, p.Metadata.Description))
	return b.String()
}
