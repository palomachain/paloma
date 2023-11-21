package types

// DONTCOVER

import (
	"github.com/VolumeFi/whoops"
	// sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdkerrors "cosmossdk.io/errors"
)

// x/evm module sentinel errors
var (
	ErrSample = sdkerrors.Register(ModuleName, 1100, "sample error")
)

var (
	ErrEthTxNotVerified = whoops.String("transaction not verified")
	ErrInvalidBalance   = whoops.Errorf("invalid balance: %s")
)
