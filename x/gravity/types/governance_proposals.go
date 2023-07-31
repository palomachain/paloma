package types

import (
	"fmt"
	"strings"

	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeIBCMetadata = "IBCMetadata"
)

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
