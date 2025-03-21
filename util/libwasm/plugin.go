package libwasm

import (
	"encoding/json"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/liberr"
	"github.com/palomachain/paloma/v2/util/liblog"
	schedulerbindings "github.com/palomachain/paloma/v2/x/scheduler/bindings/types"
	skywaybindings "github.com/palomachain/paloma/v2/x/skyway/bindings/types"
	tfbindings "github.com/palomachain/paloma/v2/x/tokenfactory/bindings/types"
)

const ErrUnrecognizedMessage = liberr.Error("unrecognized message type")

type Messenger[T any] interface {
	DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, t T) (events []sdk.Event, data [][]byte, msgResponses [][]*codectypes.Any, err error)
}

type router struct {
	legacyFallback Messenger[wasmvmtypes.CosmosMsg]
	scheduler      Messenger[schedulerbindings.Message]
	skyway         Messenger[skywaybindings.Message]
	tokenfactory   Messenger[tfbindings.Message]
	wrapped        Messenger[wasmvmtypes.CosmosMsg]
	log            log.Logger
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

func NewRouterMessageDecorator(
	log log.Logger,
	legacyFallback Messenger[wasmvmtypes.CosmosMsg],
	scheduler Messenger[schedulerbindings.Message],
	skyway Messenger[skywaybindings.Message],
	tokenfactory Messenger[tfbindings.Message],
) func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &router{
			log:            log,
			legacyFallback: legacyFallback,
			scheduler:      scheduler,
			skyway:         skyway,
			tokenfactory:   tokenfactory,
			wrapped:        old,
		}
	}
}

func (h router) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (a []sdk.Event, b [][]byte, c [][]*codectypes.Any, err error) {
	logger := liblog.FromSDKLogger(h.log).WithComponent("wasm-message-decorator")
	logger.Debug("Dispatching message...", "contract", contractAddr, "msg", msg)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered panic: %v, original error if any: %w", r, err)
			logger.WithError(err).Error("Recovered panic.")
		}

		if err != nil {
			logger.WithError(err).Error("Failed to dispatch message.")
		}

		// Annoyingly enough, we cannot get the events which have been
		// emitted during the dispatch directly without parsing.
		evts := make([]sdk.Event, len(ctx.EventManager().Events()))
		for i, v := range ctx.EventManager().Events() {
			attrs := make([]sdk.Attribute, len(v.Attributes))
			for j, k := range v.Attributes {
				attrs[j] = sdk.Attribute{
					Key:   k.Key,
					Value: k.Value,
				}
			}
			evts[i] = sdk.NewEvent(fmt.Sprintf("dispatcher_%s", v.Type), attrs...)
		}

		// Attaching any events emitted during the dispatch to the transaction
		a = evts
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
			return h.scheduler.DispatchMsg(ctx, contractAddr, contractIBCPortID, *contractMsg.Scheduler)
		case contractMsg.Skyway != nil:
			return h.skyway.DispatchMsg(ctx, contractAddr, contractIBCPortID, *contractMsg.Skyway)
		case contractMsg.TokenFactory != nil:
			return h.tokenfactory.DispatchMsg(ctx, contractAddr, contractIBCPortID, *contractMsg.TokenFactory)
		}

		logger.Debug("Calling legacy fallback message handler...")
		return h.legacyFallback.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
	}

	logger.Debug("No handler found, forwarding to chain...")
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
