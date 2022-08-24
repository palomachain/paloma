package evm

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/palomachain/paloma/x/evm/keeper"
	"github.com/palomachain/paloma/x/evm/types"
)

func NewReferenceChainReferenceIDProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
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
			_, err := k.SaveNewSmartContract(ctx, c.GetAbiJSON(), c.Bytecode())
			return err
		case *types.ChangeMinOnChainBalanceProposal:
			balance, _ := new(big.Int).SetString(c.GetMinOnChainBalance(), 10)
			err := k.ChangeMinOnChainBalance(ctx, c.GetChainReferenceID(), balance)
			return err
		}
		return sdkerrors.ErrUnknownRequest
	}
}
