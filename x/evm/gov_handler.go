package evm

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/palomachain/paloma/v2/x/evm/keeper"
	"github.com/palomachain/paloma/v2/x/evm/types"
)

func NewReferenceChainReferenceIDProposalHandler(k keeper.Keeper) govv1beta1types.Handler {
	return func(ctx sdk.Context, content govv1beta1types.Content) error {
		switch c := content.(type) {
		case *types.AddChainProposal:
			balance, ok := new(big.Int).SetString(c.GetMinOnChainBalance(), 10)
			if !ok {
				panic("cannot parse balance " + c.GetMinOnChainBalance())
			}

			return k.AddSupportForNewChain(
				ctx,
				c.GetChainReferenceID(),
				c.GetChainID(),
				c.GetBlockHeight(),
				c.GetBlockHashAtHeight(),
				balance,
			)
		case *types.RemoveChainProposal:
			return k.RemoveSupportForChain(ctx, c)

		case *types.DeployNewSmartContractProposal:
			sc, err := k.SaveNewSmartContract(ctx, c.GetAbiJSON(), c.Bytecode())
			if err != nil {
				return err
			}

			err = k.SetAsCompassContract(ctx, sc)
			return err

		case *types.ChangeMinOnChainBalanceProposal:
			balance, _ := new(big.Int).SetString(c.GetMinOnChainBalance(), 10)
			err := k.ChangeMinOnChainBalance(ctx, c.GetChainReferenceID(), balance)
			return err

		case *types.RelayWeightsProposal:
			return k.SetRelayWeights(
				ctx,
				c.GetChainReferenceID(),
				&types.RelayWeights{
					Fee:           c.Fee,
					Uptime:        c.Uptime,
					SuccessRate:   c.SuccessRate,
					ExecutionTime: c.ExecutionTime,
					FeatureSet:    c.FeatureSet,
				},
			)
		case *types.SetFeeManagerAddressProposal:
			return k.SetFeeManagerAddress(ctx, c.ChainReferenceID, c.FeeManagerAddress)
		case *types.SetSmartContractDeployersProposal:
			for _, dep := range c.Deployers {
				err := k.SetSmartContractDeployer(ctx, dep.ChainReferenceID, dep.ContractAddress)
				if err != nil {
					return err
				}
			}

			return nil
		default:
			return sdkerrors.ErrUnknownRequest
		}
	}
}
