package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/palomachain/paloma/x/gravity/types"
)

func (k Keeper) HandleCommunityPoolEthereumSpendProposal(ctx sdk.Context, p *types.CommunityPoolEthereumSpendProposal) error {
	feePool := k.DistributionKeeper.GetFeePool(ctx)

	// NOTE the community pool isn't a module account, however its coins
	// are held in the distribution module account. Thus the community pool
	// must be reduced separately from the createSendToEthereum calls
	totalToSpend := p.Amount.Add(p.BridgeFee)
	newPool, negative := feePool.CommunityPool.SafeSub(sdk.NewDecCoinsFromCoins(totalToSpend))
	if negative {
		return distributiontypes.ErrBadDistribution
	}

	feePool.CommunityPool = newPool
	sender := authtypes.NewModuleAddress(distributiontypes.ModuleName)

	txID, err := k.createSendToEthereum(ctx, sender, p.Recipient, p.Amount, p.BridgeFee)
	if err != nil {
		return err
	}

	k.DistributionKeeper.SetFeePool(ctx, feePool)
	k.Logger(ctx).Info("transfer from the community pool created as unbatched send to Ethereum", "tx ID", txID, "amount", p.Amount.String(), "recipient", p.Recipient)

	return nil
}
