package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type EventAttribute string

func (e EventAttribute) With(value string) sdk.Attribute {
	return sdk.NewAttribute(string(e), value)
}

type HasModuleName interface {
	ModuleName() string
}

type hasModuleNameFnc func() string

func (f hasModuleNameFnc) ModuleName() string {
	return f()
}

func ModuleNameFunc(moduleName string) HasModuleName {
	return hasModuleNameFnc(func() string {
		return moduleName
	})
}

func EmitEvent(k HasModuleName, ctx context.Context, name string, attrs ...sdk.Attribute) {
	sdkctx := sdk.UnwrapSDKContext(ctx)
	sdkctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			append([]sdk.Attribute{
				sdk.NewAttribute(sdk.AttributeKeyModule, k.ModuleName()),
				sdk.NewAttribute(sdk.AttributeKeyAction, name),
			},
				attrs...,
			)...,
		),
	)
}
