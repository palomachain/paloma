package types

import (
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeSetPigeonRequirements = "ValsetSetPigeonRequirementsProposal"
)

var _ govv1beta1types.Content = &SetPigeonRequirementsProposal{}

func init() {
	govv1beta1types.RegisterProposalType(ProposalTypeSetPigeonRequirements)
}

func (a *SetPigeonRequirementsProposal) ProposalRoute() string { return RouterKey }

func (a *SetPigeonRequirementsProposal) ProposalType() string {
	return ProposalTypeSetPigeonRequirements
}

func (a *SetPigeonRequirementsProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(a); err != nil {
		return err
	}

	return nil
}
