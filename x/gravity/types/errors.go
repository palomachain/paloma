package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrInternal                 = sdkerrors.Register(ModuleName, 1, "internal")
	ErrDuplicate                = sdkerrors.Register(ModuleName, 2, "duplicate")
	ErrInvalid                  = sdkerrors.Register(ModuleName, 3, "invalid")
	ErrTimeout                  = sdkerrors.Register(ModuleName, 4, "timeout")
	ErrUnknown                  = sdkerrors.Register(ModuleName, 5, "unknown")
	ErrEmpty                    = sdkerrors.Register(ModuleName, 6, "empty")
	ErrOutdated                 = sdkerrors.Register(ModuleName, 7, "outdated")
	ErrUnsupported              = sdkerrors.Register(ModuleName, 8, "unsupported")
	ErrNonContiguousEventNonce  = sdkerrors.Register(ModuleName, 9, "non contiguous event nonce, expected: %v received: %v")
	ErrResetDelegateKeys        = sdkerrors.Register(ModuleName, 10, "can not set orchestrator addresses more than once")
	ErrMismatched               = sdkerrors.Register(ModuleName, 11, "mismatched")
	ErrInvalidEthAddress        = sdkerrors.Register(ModuleName, 14, "discovered invalid eth address stored for validator %v")
	ErrDuplicateEthereumKey     = sdkerrors.Register(ModuleName, 16, "duplicate ethereum key")
	ErrDuplicateOrchestratorKey = sdkerrors.Register(ModuleName, 17, "duplicate orchestrator key")
	ErrInvalidAttestation       = sdkerrors.Register(ModuleName, 18, "invalid attestation submitted")
	ErrInvalidClaim             = sdkerrors.Register(ModuleName, 19, "invalid claim submitted")
	ErrDenomNotFound            = sdkerrors.Register(ModuleName, 21, "denom not found")
	ErrERC20NotFound            = sdkerrors.Register(ModuleName, 22, "erc20 not found")
)
