package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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
}

func NewGravityProposalHandler(k Keeper) govv1beta1types.Handler {
	return func(ctx sdk.Context, content govv1beta1types.Content) error {
		switch c := content.(type) {
		case *types.IBCMetadataProposal:
			return k.HandleIBCMetadataProposal(ctx, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized Gravity proposal content type: %T", c)
		}
	}
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
