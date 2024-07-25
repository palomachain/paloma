package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MessageFeeSettings struct {
	RelayerFee   math.LegacyDec
	CommunityFee math.LegacyDec
	SecurityFee  math.LegacyDec
}

type FeeProvider func(
	ctx context.Context,
	valAddress sdk.ValAddress,
	ChainReferenceID string) (*MessageFeeSettings, error)
