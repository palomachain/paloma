package keeper

import (
	"context"
	"fmt"
	"strconv"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/gravity/types"
)

// Check that distKeeper implements the expected type
var _ types.DistributionKeeper = (*distrkeeper.Keeper)(nil)

// AttestationHandler processes `observed` Attestations
type AttestationHandler struct {
	// NOTE: If you add anything to this struct, add a nil check to ValidateMembers below!
	keeper *Keeper
}

// Check for nil members
func (a AttestationHandler) ValidateMembers() {
	if a.keeper == nil {
		panic("Nil keeper!")
	}
}

// Handle is the entry point for Attestation processing, only attestations with sufficient validator submissions
// should be processed through this function, solidifying their effect in chain state
func (a AttestationHandler) Handle(ctx context.Context, att types.Attestation, claim types.EthereumClaim) error {
	switch claim := claim.(type) {

	case *types.MsgSendToPalomaClaim:
		return a.handleSendToPaloma(ctx, *claim)

	case *types.MsgBatchSendToEthClaim:
		return a.handleBatchSendToEth(ctx, *claim)

	default:
		return fmt.Errorf("invalid event type for attestations %s", claim.GetType())
	}
}

// Upon acceptance of sufficient validator SendToPaloma claims: transfer tokens to the appropriate paloma account
// The paloma receiver must be a native account (e.g. paloma1abc...)
// Bank module handles the transfer
func (a AttestationHandler) handleSendToPaloma(ctx context.Context, claim types.MsgSendToPalomaClaim) error {
	invalidAddress := false
	// Validate the receiver as a valid bech32 address
	receiverAddress, addressErr := types.IBCAddressFromBech32(claim.PalomaReceiver)

	if addressErr != nil {
		invalidAddress = true
		hash, er := claim.ClaimHash()
		if er != nil {
			return sdkerrors.Wrapf(er, "Unable to log error %v, could not compute ClaimHash for claim %v: %v", addressErr, claim, er)
		}

		liblog.FromSDKLogger(a.keeper.Logger(ctx)).WithFields(
			"address", receiverAddress,
			"cause", addressErr.Error(),
			"claim type", claim.GetType(),
			"id", types.GetAttestationKey(claim.GetEventNonce(), hash),
			"nonce", fmt.Sprint(claim.GetEventNonce())).Error("Invalid SendToPaloma receiver")
	}
	tokenAddress, errTokenAddress := types.NewEthAddress(claim.TokenContract)
	_, errEthereumSender := types.NewEthAddress(claim.EthereumSender)
	// nil address is not possible unless the validators get together and submit
	// a bogus event, this would create lost tokens stuck in the bridge
	// and not accessible to anyone
	if errTokenAddress != nil {
		hash, er := claim.ClaimHash()
		if er != nil {
			return sdkerrors.Wrapf(er, "Unable to log error %v, could not compute ClaimHash for claim %v: %v", errTokenAddress, claim, er)
		}
		liblog.FromSDKLogger(a.keeper.Logger(ctx)).WithFields(
			"cause", errTokenAddress.Error(),
			"claim type", claim.GetType(),
			"id", types.GetAttestationKey(claim.GetEventNonce(), hash),
			"nonce", fmt.Sprint(claim.GetEventNonce())).Error("Invalid token contract")
		return sdkerrors.Wrap(errTokenAddress, "invalid token contract on claim")
	}
	// likewise nil sender would have to be caused by a bogus event
	if errEthereumSender != nil {
		hash, er := claim.ClaimHash()
		if er != nil {
			return sdkerrors.Wrapf(er, "Unable to log error %v, could not compute ClaimHash for claim %v: %v", errEthereumSender, claim, er)
		}
		liblog.FromSDKLogger(a.keeper.Logger(ctx)).WithFields(
			"cause", errEthereumSender.Error(),
			"claim type", claim.GetType(),
			"id", types.GetAttestationKey(claim.GetEventNonce(), hash),
			"nonce", fmt.Sprint(claim.GetEventNonce())).Error("Invalid ethereum sender")
		return sdkerrors.Wrap(errTokenAddress, "invalid ethereum sender on claim")
	}

	denom, err := a.keeper.GetDenomOfERC20(ctx, claim.GetChainReferenceId(), *tokenAddress)
	if err != nil {
		return err
	}
	coin := sdk.NewCoin(denom, claim.Amount)
	coins := sdk.Coins{coin}

	moduleAddr := a.keeper.accountKeeper.GetModuleAddress(types.ModuleName)

	if !invalidAddress { // address appears valid, attempt to send minted/locked coins to receiver
		preSendBalance := a.keeper.bankKeeper.GetBalance(ctx, moduleAddr, denom)

		err := a.sendCoinToLocalAddress(ctx, claim, receiverAddress, coin)

		// Perform module balance assertions
		if err != nil { // errors should not send tokens to anyone
			err = a.assertNothingSent(ctx, moduleAddr, preSendBalance, denom)
			if err != nil {
				return err
			}
		} else { // No error, local send -> assert send had right amount
			err = a.assertSentAmount(ctx, moduleAddr, preSendBalance, denom, claim.Amount)
			if err != nil {
				return err
			}
		}

		if err != nil { // trigger send to community pool
			invalidAddress = true
		}
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// for whatever reason above, invalid string, etc this deposit is not valid
	// we can't send the tokens back on the Ethereum side, and if we don't put them somewhere on
	// the paloma side they will be lost an inaccessible even though they are locked in the bridge.
	// so we deposit the tokens into the community pool for later use via governance vote
	if invalidAddress {
		if err := a.keeper.SendToCommunityPool(ctx, coins); err != nil {
			hash, er := claim.ClaimHash()
			if er != nil {
				return sdkerrors.Wrapf(er, "Unable to log error %v, could not compute ClaimHash for claim %v: %v", err, claim, er)
			}
			liblog.FromSDKLogger(a.keeper.Logger(ctx)).WithFields(
				"cause", err.Error(),
				"claim type", claim.GetType(),
				"id", types.GetAttestationKey(claim.GetEventNonce(), hash),
				"nonce", fmt.Sprint(claim.GetEventNonce())).Error("Failed community pool send")
			return sdkerrors.Wrap(err, "failed to send to Community pool")
		}

		if err := sdkCtx.EventManager().EmitTypedEvent(
			&types.EventInvalidSendToPalomaReceiver{
				Amount: claim.Amount.String(),
				Nonce:  strconv.Itoa(int(claim.GetEventNonce())),
				Token:  tokenAddress.GetAddress().Hex(),
				Sender: claim.EthereumSender,
			},
		); err != nil {
			return err
		}

	} else {
		if err := sdkCtx.EventManager().EmitTypedEvent(
			&types.EventSendToPaloma{
				Amount: claim.Amount.String(),
				Nonce:  strconv.Itoa(int(claim.GetEventNonce())),
				Token:  tokenAddress.GetAddress().Hex(),
			},
		); err != nil {
			return err
		}
	}

	return nil
}

// Upon acceptance of sufficient validator BatchSendToEth claims: burn ethereum originated vouchers, invalidate pending
// batches with lower claim.BatchNonce, and clean up state
// Note: Previously SendToEth was referred to as a bridge "Withdrawal", as tokens are withdrawn from the gravity contract
func (a AttestationHandler) handleBatchSendToEth(ctx context.Context, claim types.MsgBatchSendToEthClaim) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	contract, err := types.NewEthAddress(claim.TokenContract)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid token contract on batch")
	}
	err = a.keeper.OutgoingTxBatchExecuted(ctx, *contract, claim)
	if err != nil {
		return err
	}
	err = sdkCtx.EventManager().EmitTypedEvent(
		&types.EventBatchSendToEthClaim{
			Nonce: strconv.Itoa(int(claim.BatchNonce)),
		},
	)

	return err
}

// assertNothingSent performs a runtime assertion that the actual sent amount of `denom` is zero
func (a AttestationHandler) assertNothingSent(ctx context.Context, moduleAddr sdk.AccAddress, preSendBalance sdk.Coin, denom string) error {
	postSendBalance := a.keeper.bankKeeper.GetBalance(ctx, moduleAddr, denom)
	if !preSendBalance.Equal(postSendBalance) {
		return fmt.Errorf(
			"SendToPaloma somehow sent tokens in an error case! Previous balance %v Post-send balance %v",
			preSendBalance.String(), postSendBalance.String(),
		)
	}
	return nil
}

// assertSentAmount performs a runtime assertion that the actual sent amount of `denom` equals the MsgSendToPaloma
// claim's amount to send
func (a AttestationHandler) assertSentAmount(ctx context.Context, moduleAddr sdk.AccAddress, preSendBalance sdk.Coin, denom string, amount math.Int) error {
	postSendBalance := a.keeper.bankKeeper.GetBalance(ctx, moduleAddr, denom)
	if !preSendBalance.Sub(postSendBalance).Amount.Equal(amount) {
		return fmt.Errorf(
			"SendToPaloma somehow sent incorrect amount! Previous balance %v Post-send balance %v claim amount %v",
			preSendBalance.String(), postSendBalance.String(), amount.String(),
		)
	}
	return nil
}

// Send tokens via bank keeper to a native gravity address, re-prefixing receiver to a gravity native address if necessary
// Note: This should only be used as part of SendToPaloma attestation handling and is not a good solution for general use
func (a AttestationHandler) sendCoinToLocalAddress(
	ctx context.Context, claim types.MsgSendToPalomaClaim, receiver sdk.AccAddress, coin sdk.Coin,
) (err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err = a.keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, sdk.NewCoins(coin))
	if err != nil {
		// log and send to Community pool
		hash, er := claim.ClaimHash()
		if er != nil {
			return sdkerrors.Wrapf(er, "Unable to log error %v, could not compute ClaimHash for claim %v: %v", err, claim, er)
		}
		liblog.FromSDKLogger(a.keeper.Logger(ctx)).WithFields(
			"cause", err.Error(),
			"claim type", claim.GetType(),
			"id", types.GetAttestationKey(claim.GetEventNonce(), hash),
			"nonce", fmt.Sprint(claim.GetEventNonce())).Error("Failed deposit")
	} else { // no error
		liblog.FromSDKLogger(a.keeper.Logger(ctx)).WithFields(
			"ethSender", claim.EthereumSender,
			"receiver", receiver,
			"denom", coin.Denom,
			"amount", coin.Amount.String(),
			"nonce", claim.EventNonce,
			"ethContract", claim.TokenContract,
			"ethBlockHeight", claim.EthBlockHeight,
			"palomaBlockHeight", sdkCtx.BlockHeight()).Info("SendToPaloma to local gravity receiver")
		if err := sdkCtx.EventManager().EmitTypedEvent(&types.EventSendToPalomaLocal{
			Nonce:    fmt.Sprint(claim.EventNonce),
			Receiver: receiver.String(),
			Token:    coin.Denom,
			Amount:   coin.Amount.String(),
		}); err != nil {
			return err
		}
	}

	return err // returns nil if no error
}
