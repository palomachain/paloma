package types

import (
	"math/big"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	proto "github.com/gogo/protobuf/proto"
	"github.com/vizualni/whoops"
)

const (
	ProposalTypeAddChain                = "EVMAddChainProposal"
	ProposalTypeRemoveChain             = "EVMRemoveChainProposal"
	ProposalDeployNewSmartContract      = "EVMDeployNewSmartContract"
	ProposalTypeChangeMinOnChainBalance = "EVMProposalChangeMinOnChainBalance"
)

var _ govtypes.Content = &AddChainProposal{}
var _ govtypes.Content = &RemoveChainProposal{}
var _ govtypes.Content = &DeployNewSmartContractProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddChain)
	govtypes.RegisterProposalTypeCodec(&AddChainProposal{}, proto.MessageName(&AddChainProposal{}))

	govtypes.RegisterProposalType(ProposalTypeRemoveChain)
	govtypes.RegisterProposalTypeCodec(&RemoveChainProposal{}, proto.MessageName(&RemoveChainProposal{}))

	govtypes.RegisterProposalType(ProposalDeployNewSmartContract)
	govtypes.RegisterProposalTypeCodec(&DeployNewSmartContractProposal{}, proto.MessageName(&DeployNewSmartContractProposal{}))

	govtypes.RegisterProposalType(ProposalTypeChangeMinOnChainBalance)
	govtypes.RegisterProposalTypeCodec(&ChangeMinOnChainBalanceProposal{}, proto.MessageName(&ChangeMinOnChainBalanceProposal{}))
}

func (a *AddChainProposal) ProposalRoute() string { return RouterKey }
func (a *AddChainProposal) ProposalType() string  { return ProposalTypeAddChain }
func (a *AddChainProposal) ValidateBasic() error {
	if err := govtypes.ValidateAbstract(a); err != nil {
		return err
	}

	return nil
}

func (a *RemoveChainProposal) ProposalRoute() string { return RouterKey }
func (a *RemoveChainProposal) ProposalType() string  { return ProposalTypeRemoveChain }
func (a *RemoveChainProposal) ValidateBasic() error {
	if err := govtypes.ValidateAbstract(a); err != nil {
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
	if err := govtypes.ValidateAbstract(a); err != nil {
		return err
	}

	return nil
}

func (a *ChangeMinOnChainBalanceProposal) ProposalRoute() string { return RouterKey }
func (a *ChangeMinOnChainBalanceProposal) ProposalType() string {
	return ProposalTypeChangeMinOnChainBalance
}
func (a *ChangeMinOnChainBalanceProposal) ValidateBasic() error {
	if err := govtypes.ValidateAbstract(a); err != nil {
		return err
	}
	_, ok := new(big.Int).SetString(a.GetMinOnChainBalance(), 10)
	if !ok {
		return whoops.String("invalid balance")
	}

	return nil
}
