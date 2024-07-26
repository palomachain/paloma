package types

import (
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeSetLightNodeClientFeegranterProposal = "PalomaSetLightNodeClientFeegranterProposal"
	ProposalTypeSetLightNodeClientFundersProposal    = "PalomaSetLightNodeClientFundersProposal"
)

var (
	_ govv1beta1types.Content = &SetLightNodeClientFeegranterProposal{}
	_ govv1beta1types.Content = &SetLightNodeClientFundersProposal{}
)

func init() {
	govv1beta1types.RegisterProposalType(ProposalTypeSetLightNodeClientFeegranterProposal)
	govv1beta1types.RegisterProposalType(ProposalTypeSetLightNodeClientFundersProposal)
}

func (a *SetLightNodeClientFeegranterProposal) ProposalRoute() string { return RouterKey }

func (a *SetLightNodeClientFeegranterProposal) ProposalType() string {
	return ProposalTypeSetLightNodeClientFeegranterProposal
}

func (a *SetLightNodeClientFeegranterProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(a); err != nil {
		return err
	}

	return nil
}

func (a *SetLightNodeClientFundersProposal) ProposalRoute() string { return RouterKey }

func (a *SetLightNodeClientFundersProposal) ProposalType() string {
	return ProposalTypeSetLightNodeClientFundersProposal
}

func (a *SetLightNodeClientFundersProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(a); err != nil {
		return err
	}

	return nil
}
