package bindings

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// `ExecuteJobWasmEvent` is a struct that was used as the only
// custom message emitted by cw contracts early on, but now
// clashes with the new implementation in `libwasm`.
// This legacy decorator exists to support the old implementation
// until all CW contracts have been phased out.
func NewLegacyMessenger(k Schedulerkeeper) wasmkeeper.Messenger {
	return &customLegacyMessenger{
		k: k,
	}
}

type customLegacyMessenger struct {
	k Schedulerkeeper
}

type executeJobWasmEvent struct {
	JobID   string `json:"job_id"`
	Payload []byte `json:"payload"`
}

func (e executeJobWasmEvent) valid() error {
	if len(e.JobID) == 0 {
		return fmt.Errorf("you must provide a jobID")
	}

	if len(e.Payload) == 0 {
		return fmt.Errorf("payload bytes is empty")
	}

	return nil
}

func (m *customLegacyMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	executeMsg, err := unmarshallJob(msg.Custom)
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to unmarshal job")
	}

	if err = executeMsg.valid(); err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to validate message")
	}

	_, err = m.k.ExecuteJob(ctx, executeMsg.JobID, executeMsg.Payload, contractAddr, contractAddr)
	if err != nil {
		return nil, nil, nil, sdkerrors.Wrap(err, "failed to trigger job execution")
	}

	return nil, nil, nil, nil
}

func unmarshallJob(msg []byte) (executeJobWasmEvent, error) {
	var executeMsg executeJobWasmEvent
	err := json.Unmarshal(msg, &executeMsg)

	hexString := hex.EncodeToString(executeMsg.Payload)

	executeMsg.Payload = []byte(fmt.Sprintf("{\"hexPayload\":\"%s\"}", hexString))

	return executeMsg, err
}
