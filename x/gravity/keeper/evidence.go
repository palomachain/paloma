package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/x/gravity/types"
)

func (k Keeper) CheckBadSignatureEvidence(
	ctx context.Context,
	msg *types.MsgSubmitBadSignatureEvidence,
	chainReferenceID string,
) error {
	var subject types.EthereumSigned

	err := k.cdc.UnpackAny(msg.Subject, &subject)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalid, fmt.Sprintf("invalid Any encoded evidence %s", err))
	}

	switch subject := subject.(type) {
	case *types.OutgoingTxBatch:
		subject.ChainReferenceId = chainReferenceID
		return k.checkBadSignatureEvidenceInternal(ctx, subject, msg.Signature)

	default:
		return sdkerrors.Wrap(types.ErrInvalid, fmt.Sprintf("bad signature must be over a batch. got %s", subject))
	}
}

func (k Keeper) checkBadSignatureEvidenceInternal(ctx context.Context, subject types.EthereumSigned, signature string) error {
	// Get checkpoint of the supposed bad signature (fake batch submitted to eth)

	ci, err := k.evmKeeper.GetChainInfo(ctx, subject.GetChainReferenceID())
	if err != nil {
		return sdkerrors.Wrap(err, "unable to create batch")
	}
	turnstoneID := string(ci.SmartContractUniqueID)
	checkpoint, err := subject.GetCheckpoint(turnstoneID)
	if err != nil {
		return err
	}
	// Try to find the checkpoint in the archives. If it exists, we don't slash because
	// this is not a bad signature
	if k.GetPastEthSignatureCheckpoint(ctx, checkpoint) {
		return sdkerrors.Wrap(types.ErrInvalid, "Checkpoint exists, cannot slash")
	}

	// Decode Eth signature to bytes

	// strip 0x prefix if needed
	if signature[:2] == "0x" {
		signature = signature[2:]
	}
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalid, fmt.Sprintf("signature decoding %s", signature))
	}

	// Get eth address of the offending validator using the checkpoint and the signature
	ethAddress, err := types.EthAddressFromSignature(checkpoint, sigBytes)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalid, fmt.Sprintf("signature to eth address failed with checkpoint %s and signature %s", hex.EncodeToString(checkpoint), signature))
	}

	// Find the offending validator by eth address
	val, found, err := k.GetValidatorByEthAddress(ctx, *ethAddress, subject.GetChainReferenceID())
	if err != nil {
		return err
	}
	if !found {
		return sdkerrors.Wrap(types.ErrInvalid, fmt.Sprintf("Did not find validator for eth address %s from signature %s with checkpoint %s and TurnstoneID %s", ethAddress.GetAddress().Hex(), signature, hex.EncodeToString(checkpoint), turnstoneID))
	}

	// Slash the offending validator
	cons, err := val.GetConsAddr()
	if err != nil {
		return sdkerrors.Wrap(err, "Could not get consensus key address for validator")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(ctx)
	if !val.IsJailed() {
		err := k.StakingKeeper.Jail(ctx, cons)
		if err != nil {
			return err
		}
		_, err = k.StakingKeeper.Slash(ctx, cons, sdkCtx.BlockHeight(), val.ConsensusPower(sdk.DefaultPowerReduction), params.SlashFractionBadEthSignature)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetPastEthSignatureCheckpoint puts the checkpoint of a batch into a set
// in order to prove later that it existed at one point.
func (k Keeper) SetPastEthSignatureCheckpoint(ctx context.Context, checkpoint []byte) {
	store := k.GetStore(ctx)
	store.Set(types.GetPastEthSignatureCheckpointKey(checkpoint), []byte{0x1})
}

// GetPastEthSignatureCheckpoint tells you whether a given checkpoint has ever existed
func (k Keeper) GetPastEthSignatureCheckpoint(ctx context.Context, checkpoint []byte) (found bool) {
	store := k.GetStore(ctx)
	if bytes.Equal(store.Get(types.GetPastEthSignatureCheckpointKey(checkpoint)), []byte{0x1}) {
		return true
	} else {
		return false
	}
}

func (k Keeper) IteratePastEthSignatureCheckpoints(ctx context.Context, cb func(key []byte, value []byte) (stop bool)) error {
	prefixStore := prefix.NewStore(k.GetStore(ctx), types.PastEthSignatureCheckpointKey)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		val := iter.Value()
		if !bytes.Equal(val, []byte{0x1}) {
			return fmt.Errorf("Invalid stored past eth signature checkpoint key=%v: value %v", key, val)
		}

		if cb(key, val) {
			break
		}
	}
	return nil
}
