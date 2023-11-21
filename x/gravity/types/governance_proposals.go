package types

import (
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeSetERC20ToDenomProposal = "SetERC20ToDenomProposal"
)

var _ govv1beta1types.Content = &SetERC20ToDenomProposal{}

func init() {
	govv1beta1types.RegisterProposalType(ProposalTypeSetERC20ToDenomProposal)
}

func (p *SetERC20ToDenomProposal) ProposalRoute() string { return RouterKey }
func (p *SetERC20ToDenomProposal) ProposalType() string  { return ProposalTypeSetERC20ToDenomProposal }
func (p *SetERC20ToDenomProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(p); err != nil {
		return err
	}

	return nil
}
