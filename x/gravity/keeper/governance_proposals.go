package keeper

import (
	"fmt"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/palomachain/paloma/x/gravity/types"
)

// this file contains code related to custom governance proposals

func RegisterProposalTypes() {
	// use of prefix stripping to prevent a typo between the proposal we check
	// and the one we register, any issues with the registration string will prevent
	// the proposal from working. We must check for double registration so that cli commands
	// submitting these proposals work.
	// For some reason the cli code is run during app.go startup, but of course app.go is not
	// run during operation of one off tx commands, so we need to run this 'twice'
	prefix := "gravity/"
	metadata := "gravity/IBCMetadata"
	if !govv1beta1types.IsValidProposalType(strings.TrimPrefix(metadata, prefix)) {
		govv1beta1types.RegisterProposalType(types.ProposalTypeIBCMetadata)
	}
	unhalt := "gravity/UnhaltBridge"
	if !govv1beta1types.IsValidProposalType(strings.TrimPrefix(unhalt, prefix)) {
		govv1beta1types.RegisterProposalType(types.ProposalTypeUnhaltBridge)
	}
	airdrop := "gravity/Airdrop"
	if !govv1beta1types.IsValidProposalType(strings.TrimPrefix(airdrop, prefix)) {
		govv1beta1types.RegisterProposalType(types.ProposalTypeAirdrop)
	}
}

func NewGravityProposalHandler(k Keeper) govv1beta1types.Handler {
	return func(ctx sdk.Context, content govv1beta1types.Content) error {
		switch c := content.(type) {
		case *types.UnhaltBridgeProposal:
			return k.HandleUnhaltBridgeProposal(ctx, c)
		case *types.AirdropProposal:
			return k.HandleAirdropProposal(ctx, c)
		case *types.IBCMetadataProposal:
			return k.HandleIBCMetadataProposal(ctx, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized Gravity proposal content type: %T", c)
		}
	}
}

// Unhalt Bridge specific functions

// In the event the bridge is halted and governance has decided to reset oracle
// history, we roll back oracle history and reset the parameters
func (k Keeper) HandleUnhaltBridgeProposal(ctx sdk.Context, p *types.UnhaltBridgeProposal) error {
	ctx.Logger().Info("Gov vote passed: Resetting oracle history", "nonce", p.TargetNonce)
	pruneAttestationsAfterNonce(ctx, k, p.TargetNonce)
	return nil
}

// Iterate over all attestations currently being voted on in order of nonce
// and prune those that are older than nonceCutoff
func pruneAttestationsAfterNonce(ctx sdk.Context, k Keeper, nonceCutoff uint64) {
	// Decide on the most recent nonce we can actually roll back to
	lastObserved := k.GetLastObservedEventNonce(ctx)
	if nonceCutoff < lastObserved || nonceCutoff == 0 {
		ctx.Logger().Error("Attempted to reset to a nonce before the last \"observed\" event, which is not allowed", "lastObserved", lastObserved, "nonce", nonceCutoff)
		return
	}

	// Get relevant event nonces
	attmap, keys := k.GetAttestationMapping(ctx)

	// Discover all affected validators whose LastEventNonce must be reset to nonceCutoff

	numValidators := len(k.StakingKeeper.GetBondedValidatorsByPower(ctx))
	// void and setMember are necessary for sets to work
	type void struct{}
	var setMember void
	// Initialize a Set of validators
	affectedValidatorsSet := make(map[string]void, numValidators)

	// Delete all reverted attestations, keeping track of the validators who attested to any of them
	for _, nonce := range keys {
		for _, att := range attmap[nonce] {
			// we delete all attestations earlier than the cutoff event nonce
			if nonce > nonceCutoff {
				ctx.Logger().Info(fmt.Sprintf("Deleting attestation at height %v", att.Height))
				for _, vote := range att.Votes {
					if _, ok := affectedValidatorsSet[vote]; !ok { // if set does not contain vote
						affectedValidatorsSet[vote] = setMember // add key to set
					}
				}

				k.DeleteAttestation(ctx, att)
			}
		}
	}

	// Reset the last event nonce for all validators affected by history deletion
	for vote := range affectedValidatorsSet {
		val, err := sdk.ValAddressFromBech32(vote)
		if err != nil {
			panic(sdkerrors.Wrap(err, "invalid validator address affected by bridge reset"))
		}
		valLastNonce := k.GetLastEventNonceByValidator(ctx, val)
		if valLastNonce > nonceCutoff {
			ctx.Logger().Info("Resetting validator's last event nonce due to bridge unhalt", "validator", vote, "lastEventNonce", valLastNonce, "resetNonce", nonceCutoff)
			k.SetLastEventNonceByValidator(ctx, val, nonceCutoff)
		}
	}
}

// Allows governance to deploy an airdrop to a provided list of addresses
func (k Keeper) HandleAirdropProposal(ctx sdk.Context, p *types.AirdropProposal) error {
	ctx.Logger().Info("Gov vote passed: Performing airdrop")
	startingSupply := k.bankKeeper.GetSupply(ctx, p.Denom)

	validateDenom := sdk.ValidateDenom(p.Denom)
	if validateDenom != nil {
		ctx.Logger().Info("Airdrop failed to execute invalid denom!")
		return sdkerrors.Wrap(types.ErrInvalid, "Invalid airdrop denom")
	}

	feePool := k.DistKeeper.GetFeePool(ctx)
	feePoolAmount := feePool.CommunityPool.AmountOf(p.Denom)

	airdropTotal := sdk.NewInt(0)
	for _, v := range p.Amounts {
		airdropTotal = airdropTotal.Add(sdk.NewIntFromUint64(v))
	}

	totalRequiredDecCoin := sdk.NewDecCoinFromCoin(sdk.NewCoin(p.Denom, airdropTotal))

	// check that we have enough tokens in the community pool to actually execute
	// this airdrop with the provided recipients list
	totalRequiredDec := totalRequiredDecCoin.Amount
	if totalRequiredDec.GT(feePoolAmount) {
		ctx.Logger().Info("Airdrop failed to execute insufficient tokens in the community pool!")
		return sdkerrors.Wrap(types.ErrInvalid, "Insufficient tokens in community pool")
	}

	// we're packing addresses as 20 bytes rather than valid bech32 in order to maximize participants
	// so if the recipients list is not a multiple of 20 it must be invalid
	numRecipients := len(p.Recipients) / 20
	if len(p.Recipients)%20 != 0 || numRecipients != len(p.Amounts) {
		ctx.Logger().Info("Airdrop failed to execute invalid recipients")
		return sdkerrors.Wrap(types.ErrInvalid, "Invalid recipients")
	}

	parsedRecipients := make([]sdk.AccAddress, len(p.Recipients)/20)
	for i := 0; i < numRecipients; i++ {
		indexStart := i * 20
		indexEnd := indexStart + 20
		addr := p.Recipients[indexStart:indexEnd]
		parsedRecipients[i] = addr
	}

	// check again, just in case the above modulo math is somehow wrong or spoofed
	if len(parsedRecipients) != len(p.Amounts) {
		ctx.Logger().Info("Airdrop failed to execute invalid recipients")
		return sdkerrors.Wrap(types.ErrInvalid, "Invalid recipients")
	}

	// the total amount actually sent in dec coins
	totalSent := sdk.NewDec(0)
	for i, addr := range parsedRecipients {
		usersAmount := p.Amounts[i]
		usersIntAmount := sdk.NewIntFromUint64(usersAmount)
		usersDecAmount := sdk.NewDecFromInt(usersIntAmount)
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, disttypes.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(p.Denom, usersIntAmount)))
		// if there is no error we add to the total actually sent
		if err == nil {
			totalSent = totalSent.Add(usersDecAmount)
		} else {
			// return an err to prevent execution from finishing, this will prevent the changes we
			// have made so far from taking effect the governance proposal will instead time out
			ctx.Logger().Info("invalid address in airdrop! not executing", "address", addr)
			return err
		}
	}

	if !totalRequiredDecCoin.Amount.Equal(totalSent) {
		ctx.Logger().Info("Airdrop failed to execute Invalid amount sent", "sent", totalRequiredDecCoin.Amount, "expected", totalSent)
		return sdkerrors.Wrap(types.ErrInvalid, "Invalid amount sent")
	}

	newCoins, InvalidModuleBalance := feePool.CommunityPool.SafeSub(sdk.NewDecCoins(totalRequiredDecCoin))
	// this shouldn't ever happen because we check that we have enough before starting
	// but lets be conservative.
	if InvalidModuleBalance {
		return sdkerrors.Wrap(types.ErrInvalid, "internal error!")
	}
	feePool.CommunityPool = newCoins
	k.DistKeeper.SetFeePool(ctx, feePool)

	endingSupply := k.bankKeeper.GetSupply(ctx, p.Denom)
	if !startingSupply.Equal(endingSupply) {
		return sdkerrors.Wrap(types.ErrInvalid, "total chain supply has changed!")
	}

	return nil
}

// handles a governance proposal for setting the metadata of an IBC token, this takes the normal
// metadata struct with one key difference, the base unit must be set as the ibc path string in order
// for setting the denom metadata to work.
func (k Keeper) HandleIBCMetadataProposal(ctx sdk.Context, p *types.IBCMetadataProposal) error {
	ctx.Logger().Info("Gov vote passed: Setting IBC Metadata", "denom", p.IbcDenom)

	// checks if the provided token denom is a proper IBC token, not a native token.
	if !strings.HasPrefix(p.IbcDenom, "ibc/") && !strings.HasPrefix(p.IbcDenom, "IBC/") {
		ctx.Logger().Info("invalid denom for metadata proposal", "denom", p.IbcDenom)
		return sdkerrors.Wrap(types.ErrInvalid, "Target denom is not an IBC token")
	}

	// check that our base unit is the IBC token name on this chain. This makes setting/loading denom
	// metadata work out, as SetDenomMetadata uses the base denom as an index
	if p.Metadata.Base != p.IbcDenom {
		ctx.Logger().Info("invalid metadata for metadata proposal must be the same as IBCDenom", "base", p.Metadata.Base)
		return sdkerrors.Wrap(types.ErrInvalid, "Metadata base must be the same as the IBC denom!")
	}

	// outsource validating this to the bank validation function
	metadataErr := p.Metadata.Validate()
	if metadataErr != nil {
		ctx.Logger().Info("invalid metadata for metadata proposal", "validation error", metadataErr)
		return sdkerrors.Wrap(metadataErr, "Invalid metadata")

	}

	// if metadata already exists then changing it is only a good idea if we have not already deployed an ERC20
	// for this denom if we have we can't change it
	_, metadataExists := k.bankKeeper.GetDenomMetaData(ctx, p.IbcDenom)
	_, erc20RepresentationExists := k.GetCosmosOriginatedERC20(ctx, p.IbcDenom)
	if metadataExists && erc20RepresentationExists {
		ctx.Logger().Info("invalid trying to set metadata when ERC20 has already been deployed")
		return sdkerrors.Wrap(types.ErrInvalid, "Metadata can only be changed before ERC20 is created")

	}

	// write out metadata, this will update existing metadata if no erc20 has been deployed
	k.bankKeeper.SetDenomMetaData(ctx, p.Metadata)

	return nil
}
