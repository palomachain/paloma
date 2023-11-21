package types

import (
	"math/big"

	"github.com/VolumeFi/whoops"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ethereum/go-ethereum/common"
)

const (
	ProposalTypeAddChain                = "EVMAddChainProposal"
	ProposalTypeRemoveChain             = "EVMRemoveChainProposal"
	ProposalDeployNewSmartContract      = "EVMDeployNewSmartContract"
	ProposalTypeChangeMinOnChainBalance = "EVMProposalChangeMinOnChainBalance"
	ProposalTypeRelayWeights            = "EVMProposalRelayWeights"
)

var (
	_ govv1beta1types.Content = &AddChainProposal{}
	_ govv1beta1types.Content = &RemoveChainProposal{}
	_ govv1beta1types.Content = &DeployNewSmartContractProposal{}
	_ govv1beta1types.Content = &RelayWeightsProposal{}
)

func init() {
	govv1beta1types.RegisterProposalType(ProposalTypeAddChain)
	govv1beta1types.RegisterProposalType(ProposalTypeRemoveChain)
	govv1beta1types.RegisterProposalType(ProposalDeployNewSmartContract)
	govv1beta1types.RegisterProposalType(ProposalTypeChangeMinOnChainBalance)
	govv1beta1types.RegisterProposalType(ProposalTypeRelayWeights)
}

func (a *AddChainProposal) ProposalRoute() string { return RouterKey }
func (a *AddChainProposal) ProposalType() string  { return ProposalTypeAddChain }
func (a *AddChainProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(a); err != nil {
		return err
	}

	return nil
}

func (a *RemoveChainProposal) ProposalRoute() string { return RouterKey }
func (a *RemoveChainProposal) ProposalType() string  { return ProposalTypeRemoveChain }
func (a *RemoveChainProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(a); err != nil {
		return err
	}

	return nil
}

func (a *DeployNewSmartContractProposal) Bytecode() []byte {
	return common.FromHex(a.GetBytecodeHex())
}
func (a *DeployNewSmartContractProposal) ProposalRoute() string { return RouterKey }
func (a *DeployNewSmartContractProposal) ProposalType() string  { return ProposalDeployNewSmartContract }
func (a *DeployNewSmartContractProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(a); err != nil {
		return err
	}

	return nil
}

func (a *ChangeMinOnChainBalanceProposal) ProposalRoute() string { return RouterKey }
func (a *ChangeMinOnChainBalanceProposal) ProposalType() string {
	return ProposalTypeChangeMinOnChainBalance
}

func (a *ChangeMinOnChainBalanceProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(a); err != nil {
		return err
	}
	_, ok := new(big.Int).SetString(a.GetMinOnChainBalance(), 10)
	if !ok {
		return whoops.String("invalid balance")
	}

	return nil
}

func (a *RelayWeightsProposal) ProposalRoute() string { return RouterKey }
func (a *RelayWeightsProposal) ProposalType() string  { return ProposalTypeRelayWeights }
func (a *RelayWeightsProposal) ValidateBasic() error {
	if err := govv1beta1types.ValidateAbstract(a); err != nil {
		return err
	}

	return nil
}
