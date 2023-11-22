package wasm

import (
	// wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MessengerFnc func(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string) (events []sdk.Event, data [][]byte, err error)

func (m MessengerFnc) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string) (events []sdk.Event, data [][]byte, err error) {
	return m(ctx, contractAddr, contractIBCPortID)
}
