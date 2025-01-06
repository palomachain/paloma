package libwasm

import (
	"encoding/json"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/liblog"
	schedulerbindings "github.com/palomachain/paloma/v2/x/scheduler/bindings/types"
	skywaybindings "github.com/palomachain/paloma/v2/x/skyway/bindings/types"
	tfbindings "github.com/palomachain/paloma/v2/x/tokenfactory/bindings/types"
)

type Messenger struct {
	legacyFallback wasmkeeper.Messenger
	scheduler      wasmkeeper.Messenger
	skyway         wasmkeeper.Messenger
	tokenfactory   wasmkeeper.Messenger
	wrapped        wasmkeeper.Messenger
}

type Querier struct {
	scheduler    wasmkeeper.CustomQuerier
	skyway       wasmkeeper.CustomQuerier
	tokenfactory wasmkeeper.CustomQuerier
}

type CustomMessage struct {
	Scheduler    *schedulerbindings.Message `json:"scheduler_msg,omitempty"`
	Skyway       *skywaybindings.Message    `json:"skyway_msg,omitempty"`
	TokenFactory *tfbindings.Message        `json:"token_factory_msg,omitempty"`
}

type CustomQuery struct {
	Scheduler    *schedulerbindings.Query `json:"scheduler_query,omitempty"`
	Skyway       *skywaybindings.Query    `json:"skyway_query,omitempty"`
	TokenFactory *tfbindings.Query        `json:"token_factory_query,omitempty"`
}

func NewMessenger(
	legacyFallback wasmkeeper.Messenger,
	scheduler wasmkeeper.Messenger,
	skyway wasmkeeper.Messenger,
	tokenfactory wasmkeeper.Messenger,
) func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &Messenger{
			legacyFallback: legacyFallback,
			scheduler:      scheduler,
			skyway:         skyway,
			tokenfactory:   tokenfactory,
			wrapped:        old,
		}
	}
}

func (h Messenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (a []sdk.Event, b [][]byte, c [][]*codectypes.Any, err error) {
	logger := liblog.FromSDKLogger(ctx.Logger()).WithComponent("wasm-message-decorator")

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered panic: %v, original error if any: %w", r, err)
			logger.WithError(err).Error("Recovered panic.")
		}

		if err != nil {
			logger.WithError(err).Error("Failed to dispatch message.")
		}
	}()

	if msg.Custom != nil {
		var contractMsg CustomMessage
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			// Since all fields are marked as optional, this should never happen
			logger.WithError(err).Error("Failed to unmarshal custom message")
			return nil, nil, nil, sdkerrors.Wrap(err, "custom message type mismatch")
		}
		switch {
		case contractMsg.Scheduler != nil:
			return h.scheduler.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
		case contractMsg.Skyway != nil:
			return h.skyway.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
		case contractMsg.TokenFactory != nil:
			return h.tokenfactory.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
		}
	}

	logger.Debug("Calling fallback message handler...")
	return h.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

func NewQuerier(
	scheduler wasmkeeper.CustomQuerier,
	skyway wasmkeeper.CustomQuerier,
	tokenfactory wasmkeeper.CustomQuerier,
) wasmkeeper.CustomQuerier {
	return Querier{
		scheduler:    scheduler,
		skyway:       skyway,
		tokenfactory: tokenfactory,
	}.CustomQuerier
}

func (q Querier) CustomQuerier(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	var contractQuery CustomQuery
	if err := json.Unmarshal(request, &contractQuery); err != nil {
		// Since all fields are marked as optional, this should never happen
		return nil, sdkerrors.Wrap(err, "custom query type mismatch")
	}
	switch {
	case contractQuery.Scheduler != nil:
		return q.scheduler(ctx, request)
	case contractQuery.Skyway != nil:
		return q.skyway(ctx, request)
	case contractQuery.TokenFactory != nil:
		return q.tokenfactory(ctx, request)
	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown scheduler query variant"}
	}
}
