package types

import (
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeSetERC20ToDenomProposal = "SetERC20ToDenomProposal"
	ProposalTypeSetBridgeTaxProposal    = "SetBridgeTaxProposal"
)

var (
	_ govv1beta1types.Content = &SetERC20ToDenomProposal{}
	_ govv1beta1types.Content = &SetBridgeTaxProposal{}
)

func init() {
	govv1beta1types.RegisterProposalType(ProposalTypeSetERC20ToDenomProposal)
	govv1beta1types.RegisterProposalType(ProposalTypeSetBridgeTaxProposal)
}

func (p *SetERC20ToDenomProposal) ProposalRoute() string { return RouterKey }
func (p *SetERC20ToDenomProposal) ProposalType() string  { return ProposalTypeSetERC20ToDenomProposal }
func (p *SetERC20ToDenomProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(p); err != nil {
		return err
	}

	return nil
}

func (p *SetBridgeTaxProposal) ProposalRoute() string { return RouterKey }
func (p *SetBridgeTaxProposal) ProposalType() string  { return ProposalTypeSetBridgeTaxProposal }
func (p *SetBridgeTaxProposal) ValidateBasic() error {
	return govv1beta1types.ValidateAbstract(p)
}
