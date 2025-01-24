package libwasm

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/log"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	schedulerbindings "github.com/palomachain/paloma/v2/x/scheduler/bindings/types"
	skywaybindings "github.com/palomachain/paloma/v2/x/skyway/bindings/types"
	tfbindings "github.com/palomachain/paloma/v2/x/tokenfactory/bindings/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessenger is a mock implementation of wasmkeeper.Messenger
type MockMessenger[T any] struct {
	mock.Mock
}

func (m *MockMessenger[T]) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg T) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	args := m.Called(ctx, contractAddr, contractIBCPortID, msg)
	return args.Get(0).([]sdk.Event), args.Get(1).([][]byte), args.Get(2).([][]*codectypes.Any), args.Error(3)
}

func TestDispatchMsg(t *testing.T) {
	ctx := sdk.Context{}
	contractAddr := sdk.AccAddress([]byte("test_address"))
	contractIBCPortID := "test_port"

	logger := log.NewNopLogger()
	mockScheduler := new(MockMessenger[schedulerbindings.Message])
	mockSkyway := new(MockMessenger[skywaybindings.Message])
	mockTokenFactory := new(MockMessenger[tfbindings.Message])
	mockLegacyFallback := new(MockMessenger[wasmvmtypes.CosmosMsg])
	mockWrapped := new(MockMessenger[wasmvmtypes.CosmosMsg])

	h := router{
		log:            logger,
		legacyFallback: mockLegacyFallback,
		scheduler:      mockScheduler,
		skyway:         mockSkyway,
		tokenfactory:   mockTokenFactory,
		wrapped:        mockWrapped,
	}

	expectedEvents := []sdk.Event{}
	expectedData := [][]byte{}
	expectedTypes := [][]*codectypes.Any{}

	tests := []struct {
		name        string
		msg         wasmvmtypes.CosmosMsg
		mockSetup   func()
		expectError bool
	}{
		{
			name: "Scheduler message",
			msg: wasmvmtypes.CosmosMsg{
				Custom: buildCustomMessage(t, &CustomMessage{
					Scheduler: &schedulerbindings.Message{},
				}),
			},
			mockSetup: func() {
				mockScheduler.On("DispatchMsg", ctx, contractAddr, contractIBCPortID, mock.Anything).
					Return(expectedEvents, expectedData, expectedTypes, nil)
			},
		},
		{
			name: "Skyway message",
			msg: wasmvmtypes.CosmosMsg{
				Custom: buildCustomMessage(t, &CustomMessage{
					Skyway: &skywaybindings.Message{},
				}),
			},
			mockSetup: func() {
				mockSkyway.On("DispatchMsg", ctx, contractAddr, contractIBCPortID, mock.Anything).
					Return(expectedEvents, expectedData, expectedTypes, nil)
			},
		},
		{
			name: "TokenFactory message",
			msg: wasmvmtypes.CosmosMsg{
				Custom: buildCustomMessage(t, &CustomMessage{
					TokenFactory: &tfbindings.Message{},
				}),
			},
			mockSetup: func() {
				mockTokenFactory.On("DispatchMsg", ctx, contractAddr, contractIBCPortID, mock.Anything).
					Return(expectedEvents, expectedData, expectedTypes, nil)
			},
		},
		{
			name: "Legacy fallback message",
			msg: wasmvmtypes.CosmosMsg{
				Custom: json.RawMessage(`{"unknown_type": {}}`),
			},
			mockSetup: func() {
				mockLegacyFallback.On("DispatchMsg", ctx, contractAddr, contractIBCPortID, mock.Anything).
					Return(expectedEvents, expectedData, expectedTypes, nil)
			},
		},
		{
			name: "Non-custom message",
			msg:  wasmvmtypes.CosmosMsg{},
			mockSetup: func() {
				mockWrapped.On("DispatchMsg", ctx, contractAddr, contractIBCPortID, mock.Anything).
					Return(expectedEvents, expectedData, expectedTypes, nil)
			},
		},
		{
			name: "Invalid custom message",
			msg: wasmvmtypes.CosmosMsg{
				Custom: json.RawMessage(`invalid json`),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			events, data, types, err := h.DispatchMsg(ctx, contractAddr, contractIBCPortID, tt.msg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedEvents, events)
				assert.Equal(t, expectedData, data)
				assert.Equal(t, expectedTypes, types)
			}
		})
	}
}

func buildCustomMessage(t *testing.T, msg *CustomMessage) json.RawMessage {
	bytes, err := json.Marshal(msg)
	assert.NoError(t, err)
	return bytes
}
