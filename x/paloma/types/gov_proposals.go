package types

import (
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeSetLightNodeClientFeegranterProposal = "PalomaSetLightNodeClientFeegranterProposal"
)

var _ govv1beta1types.Content = &SetLightNodeClientFeegranterProposal{}

func init() {
	govv1beta1types.RegisterProposalType(ProposalTypeSetLightNodeClientFeegranterProposal)
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
