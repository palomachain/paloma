package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/vizualni/whoops"
)

// x/evm module sentinel errors
var (
	ErrSample = sdkerrors.Register(ModuleName, 1100, "sample error")
)

var (
	ErrEthTxNotVerified = whoops.String("transaction not verified")
)
