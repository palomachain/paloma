package common

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func SdkContext(ctx context.Context) sdk.Context {
	return sdk.UnwrapSDKContext(ctx)
}
