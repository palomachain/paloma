package client

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// OfflineSignMsgs take the acc num, seq and a slice of messages and it signs them offline.
// Offline here means that it does not connect to the chain to get the acc num and seq, but
// rather uses what was provided.
func (cc *LensClient) OfflineSignMsgs(accNum, seq uint64, msgs ...sdk.Msg) (authsigning.Tx, error) {
	txf := cc.TxFactory().WithAccountNumber(accNum).WithSequence(seq)
	txb, err := cc.prepareMsgsForSigning(txf, msgs...)
	if err != nil {
		return nil, err
	}

	err = func() error {
		done := cc.SetSDKContext()
		defer done()
		if err = tx.Sign(txf, cc.Config.Key, txb, false); err != nil {
			return err
		}
		return nil
	}()

	if err != nil {
		return nil, err
	}

	return txb.GetTx(), nil
}

func (cc *LensClient) prepareMsgsForSigning(
	txf tx.Factory,
	msgs ...sdk.Msg,
) (client.TxBuilder, error) {

	txf, err := cc.PrepareFactory(txf)
	if err != nil {
		return nil, err
	}
	txb, err := tx.BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return nil, err
	}

	for _, msg := range msgs {
		_, err := cc.Codec.Marshaler.MarshalJSON(msg)
		if err != nil {
			return nil, err
		}
	}

	return txb, nil
}

func (cc *LensClient) VerifySignature(
	ctx context.Context,
	tx sdk.Tx,
) error {
	return verifySignature(
		ctx,
		cc.Config.ChainID,
		cc,
		cc.Codec.TxConfig.SignModeHandler(),
		tx)
}

type accountQueryier interface {
	QueryAccount(ctx context.Context, address sdk.AccAddress) (authtypes.AccountI, error)
}

func verifySignature(
	ctx context.Context,
	chainID string,
	accq accountQueryier,
	smh authsigning.SignModeHandler,
	tx sdk.Tx,
) error {

	sigTx := tx.(authsigning.SigVerifiableTx)
	signers := sigTx.GetSigners()

	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return err
	}

	if len(sigs) != len(signers) {
		return ErrTooLittleOrTooManySignaturesProvided
	}

	for i, sig := range sigs {
		var (
			pubKey  = sig.PubKey
			sigAddr = sdk.AccAddress(pubKey.Address())
		)

		if !sigAddr.Equals(signers[i]) {
			return ErrSignatureDoesNotMatchItsRespectiveSigner
		}

		// validate the actual signature over the transaction bytes since we can
		// reach out to a full node to query accounts.
		acci, err := accq.QueryAccount(ctx, sigAddr)
		if err != nil {
			return err
		}
		accNum, accSeq := acci.GetAccountNumber(), acci.GetSequence()
		signingData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNum,
			Sequence:      accSeq,
		}
		err = authsigning.VerifySignature(pubKey, signingData, sig.Data, smh, sigTx)
		if err != nil {
			return ErrSignatureVerificationFailed
		}

	}

	return nil
}
