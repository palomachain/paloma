package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
)

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options ante.HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

// SigVerificationDecorator wraps the default SDK SigVerificationDecorator
// AnteHandler decorator and implements custom Paloma logic based on execution
// context. Specifically, the decorator will be bypassed during CheckTx if a tx
// contains certain messages.
type SigVerificationDecorator struct {
	ante.SigVerificationDecorator
}

func NewSigVerificationDecorator(ak ante.AccountKeeper, signModeHandler authsigning.SignModeHandler) SigVerificationDecorator {
	return SigVerificationDecorator{
		SigVerificationDecorator: ante.NewSigVerificationDecorator(ak, signModeHandler),
	}
}

func (svd SigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if BypassSigVerificationDecorator(ctx, tx) {
		return next(ctx, tx, simulate)
	}

	// In cases where we do not need to bypass or the execution context is
	// PrepareProposal, ProcessProposal, or DeliverTx, we execute the default
	// decorator.
	return svd.SigVerificationDecorator.AnteHandle(ctx, tx, simulate, next)
}

// BypassSigVerificationDecorator returns true if the signature verification
// decorator should be bypassed. The decorator will be bypassed iff:
//
// - Execution mode is CheckTx
// - Contains a single message of a support type
func BypassSigVerificationDecorator(ctx sdk.Context, tx sdk.Tx) bool {
	if ctx.IsCheckTx() || ctx.IsReCheckTx() {
		if len(tx.GetMsgs()) == 1 {
			msg := tx.GetMsgs()[0]
			switch msg.(type) {
			case *consensustypes.MsgAddEvidence:
				return true

			default:
				return false
			}
		}
	}

	return false
}
