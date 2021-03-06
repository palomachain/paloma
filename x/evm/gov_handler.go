package evm

import (
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
			return k.AddSupportForNewChain(
				ctx,
				c.GetChainReferenceID(),
				c.GetChainID(),
				c.GetBlockHeight(),
				c.GetBlockHashAtHeight(),
			)
		case *types.RemoveChainProposal:
			return k.RemoveSupportForChain(ctx, c)
		case *types.DeployNewSmartContractProposal:
			_, err := k.SaveNewSmartContract(ctx, c.GetAbiJSON(), c.Bytecode())
			return err
		}
		return sdkerrors.ErrUnknownRequest
	}
}
