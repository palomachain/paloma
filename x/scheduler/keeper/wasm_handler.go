package keeper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/VolumeFi/whoops"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/liblog"
	wasmutil "github.com/palomachain/paloma/v2/util/wasm"
	"github.com/palomachain/paloma/v2/x/scheduler/types"
)

type ExecuteJobWasmEvent struct {
	JobID   string `json:"job_id"`
	Sender  string `json:"sender"`
	Payload []byte `json:"payload"`
}

func (e ExecuteJobWasmEvent) valid() error {
	if len(e.JobID) == 0 {
		return fmt.Errorf("you must provide a jobID")
	}

	if len(e.Payload) == 0 {
		return fmt.Errorf("payload bytes is empty")
	}

	// todo: add more in the future
	return nil
}

func (k Keeper) UnmarshallJob(msg []byte) (ExecuteJobWasmEvent, error) {
	var executeMsg ExecuteJobWasmEvent
	err := json.Unmarshal(msg, &executeMsg)

	hexString := hex.EncodeToString(executeMsg.Payload)

	executeMsg.Payload = []byte(fmt.Sprintf("{\"hexPayload\":\"%s\"}", hexString))

	return executeMsg, err
}

func (k Keeper) ExecuteWasmJobEventListener() wasmutil.MessengerFnc {
	return func(ctx sdk.Context, contractAddr sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
		logger := liblog.FromSDKLogger(k.Logger(ctx))
		executeMsg, err := k.UnmarshallJob(msg.Custom)
		if err != nil {
			logger.WithError(err).Error("Failed to unmarshal job.")
			return nil, nil, nil, err
		}

		logger = logger.WithFields("job-id", executeMsg.JobID)
		logger.Debug("Got a request to trigger a job via CosmWasm smart contract.")
		if err = executeMsg.valid(); err != nil {
			logger.WithError(err).Error("Message validation failed.")
			return nil, nil, nil, whoops.Wrap(err, types.ErrWasmExecuteMessageNotValid)
		}

		msgID, err := k.ExecuteJob(ctx, executeMsg.JobID, executeMsg.Payload, contractAddr, contractAddr)
		if err != nil {
			logger.WithError(err).Error("Failed to trigger job execution.")
			return nil, nil, nil, err
		}

		logger.WithFields("msg-id", strconv.FormatUint(msgID, 10)).Debug("Job execution triggered.")
		return nil, nil, nil, nil
	}
}
