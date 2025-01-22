package paloma

import (
	"context"
	"fmt"
	"math"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errtypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/palomachain/paloma/v2/util/liblog"
	"github.com/palomachain/paloma/v2/util/libmeta"
	"github.com/palomachain/paloma/v2/x/paloma/types"
	vtypes "github.com/palomachain/paloma/v2/x/valset/types"
)

func logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return liblog.FromSDKLogger(sdkCtx.Logger()).With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// HandlerDecorator is an ante decorator wrapper for an ante handler
type HandlerDecorator struct {
	handler sdk.AnteHandler
}

// NewAnteHandlerDecorator constructor for HandlerDecorator
func NewAnteHandlerDecorator(handler sdk.AnteHandler) HandlerDecorator {
	return HandlerDecorator{handler}
}

// AnteHandle wraps the next AnteHandler to perform custom pre- and post-processing
func (decorator HandlerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if newCtx, err = decorator.handler(ctx, tx, simulate); err != nil {
		return newCtx, err
	}

	return next(newCtx, tx, simulate)
}

// LogMsgDecorator logs all messages in blocks
type LogMsgDecorator struct {
	cdc codec.Codec
}

// NewLogMsgDecorator is the constructor for LogMsgDecorator
func NewLogMsgDecorator(cdc codec.Codec) LogMsgDecorator {
	return LogMsgDecorator{cdc: cdc}
}

// AnteHandle logs all messages in blocks
func (d LogMsgDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if simulate || ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()

	for _, msg := range msgs {
		logger(ctx).Debug(fmt.Sprintf("received message of type %s in block %d: %s",
			proto.MessageName(msg),
			ctx.BlockHeight(),
			string(d.cdc.MustMarshalJSON(msg)),
		))
	}

	return next(ctx, tx, simulate)
}

// VerifyAuthorisedSignatureDecorator verifies that the message is signed by at least one signature that has
// active fee grant from the creator address, IF it contains metadata.
type VerifyAuthorisedSignatureDecorator struct {
	fk types.FeegrantKeeper
}

func NewVerifyAuthorisedSignatureDecorator(fk types.FeegrantKeeper) VerifyAuthorisedSignatureDecorator {
	return VerifyAuthorisedSignatureDecorator{fk: fk}
}

// AnteHandle verifies that the message is signed by at least one signature that has
// active fee grant from the creator address, IF the message contains metadata.
func (d VerifyAuthorisedSignatureDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if simulate {
		return next(ctx, tx, simulate)
	}

	for _, msg := range tx.GetMsgs() {
		m, ok := msg.(libmeta.MsgWithMetadata[vtypes.MsgMetadata])
		if !ok {
			logger(ctx).Debug(fmt.Sprintf("msg %s does not contain metadata. skipping ownership verification...", proto.MessageName(msg)))
			continue
		}

		creator := m.GetMetadata().GetCreator()
		signers := libmeta.GetSigners(m)

		signedByCreator := func() bool {
			for _, v := range signers {
				if v.String() == creator {
					return true
				}
			}
			return false
		}()
		if signedByCreator {
			logger(ctx).Debug(fmt.Sprintf("msg %s was signed by creator.", proto.MessageName(msg)))
			continue
		}

		grants, err := d.fk.AllowancesByGranter(ctx, &feegrant.QueryAllowancesByGranterRequest{
			Granter: creator,
		})
		if err != nil {
			return ctx, fmt.Errorf("failed to verify message signature authorisation: %w", err)
		}

		logger(ctx).Debug(fmt.Sprintf("got %d allowances from granter %s", len(grants.GetAllowances()), creator))
		grantsLkUp := map[string]feegrant.Grant{}
		for _, v := range grants.GetAllowances() {
			if v == nil {
				continue
			}

			logger(ctx).Debug("found allowance", "granter", v.GetGranter(), "grantee", v.GetGrantee())
			grantsLkUp[v.GetGrantee()] = *v
		}

		grantees := make([]string, 0, len(signers))
		for _, signer := range signers {
			if v, found := grantsLkUp[signer.String()]; found {
				logger(ctx).Debug("found granted signature", "signature", v.Grantee)
				grantees = append(grantees, v.Grantee)
			}
		}

		if len(grantees) < 1 {
			return ctx, fmt.Errorf("no signature from granted address found for message %s", proto.MessageName(msg))
		}

		logger(ctx).Debug(fmt.Sprintf("found total of %d signatures from granted addresses for message %s", len(grantees), proto.MessageName(msg)))
	}

	return next(ctx, tx, simulate)
}

type PalomaKeeper interface {
	GetParams(context.Context) types.Params
}

type freeGasMeter struct{}

var _ storetypes.GasMeter = (*freeGasMeter)(nil)

func (f *freeGasMeter) ConsumeGas(amount uint64, descriptor string) {}
func (f *freeGasMeter) RefundGas(amount uint64, descriptor string)  {}
func (f *freeGasMeter) GasConsumed() uint64                         { return 0 }
func (f *freeGasMeter) GasConsumedToLimit() uint64                  { return 0 }
func (f *freeGasMeter) GasRemaining() uint64                        { return math.MaxUint64 }
func (f *freeGasMeter) IsOutOfGas() bool                            { return false }
func (f *freeGasMeter) IsPastLimit() bool                           { return false }
func (f *freeGasMeter) Limit() uint64                               { return math.MaxUint64 }
func (f *freeGasMeter) String() string {
	return fmt.Sprintf("freeGasMeter:\n  consumed: %s", "no tracking")
}

// GasExemptAddressDecorator inspects transactions sender for a matching address from the
// whitelist of gas exempt addresses. If a match is found, the gas required for the transaction
// is reduced to 0.
type GasExemptAddressDecorator struct {
	k PalomaKeeper
}

func NewGasExemptAddressDecorator(k PalomaKeeper) GasExemptAddressDecorator {
	return GasExemptAddressDecorator{k}
}

func (d GasExemptAddressDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if simulate {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		// TODO: FIX
		return ctx, sdkerrors.Wrap(errtypes.ErrTxDecode, "invalid transaction type")
	}

	sender := sdk.AccAddress(feeTx.FeePayer()).String()
	params := d.k.GetParams(ctx)
	found := false
	for _, v := range params.GasExemptAddresses {
		if v == sender {
			found = true
			break
		}
	}

	if found {
		ctx = ctx.WithGasMeter(&freeGasMeter{})
	}

	return next(ctx, tx, simulate)
}

// TxFeeSkipper is a TxFeeChecker that skips fee deduction entirely.
// Every transaction will be considered valid and no fee will be deducted.
// This also means that every transaction will be prioritized equally.
func TxFeeSkipper(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	return sdk.NewCoins(), 42, nil
}
