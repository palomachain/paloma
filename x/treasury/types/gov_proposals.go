package types

import (
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeCommunityFundFee = "CommunityFundFeeProposal"
	ProposalTypeSecurityFee      = "SecurityFeeProposal"
)

var (
	_ govv1beta1types.Content = &CommunityFundFeeProposal{}
	_ govv1beta1types.Content = &SecurityFeeProposal{}
)

func init() {
	govv1beta1types.RegisterProposalType(ProposalTypeCommunityFundFee)
	govv1beta1types.RegisterProposalType(ProposalTypeSecurityFee)
}

func (p *CommunityFundFeeProposal) ProposalRoute() string { return RouterKey }
func (p *CommunityFundFeeProposal) ProposalType() string  { return ProposalTypeCommunityFundFee }
func (p *CommunityFundFeeProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(p); err != nil {
		return err
	}

	return nil
}

func (p *SecurityFeeProposal) ProposalRoute() string { return RouterKey }
func (p *SecurityFeeProposal) ProposalType() string  { return ProposalTypeSecurityFee }
func (p *SecurityFeeProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(p); err != nil {
		return err
	}

	return nil
}
