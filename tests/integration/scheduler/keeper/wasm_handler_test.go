package keeper_test

import (
	"encoding/json"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	xchainmocks "github.com/palomachain/paloma/internal/x-chain/mocks"
	"github.com/palomachain/paloma/x/scheduler/keeper"
	"github.com/palomachain/paloma/x/scheduler/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("wasm message handler", func() {
	t := GinkgoT()
	f := initFixture(t)
	var ctx sdk.Context

	BeforeEach(func() {
		f = initFixture(t)
		ctx = f.ctx.WithBlockHeight(5)
	})

	var subjectMsg keeper.ExecuteJobWasmEvent
	var serializedSubjectMessage []byte

	subject := func() error {
		_, _, err := f.schedulerKeeper.ExecuteWasmJobEventListener()(ctx, sdk.AccAddress("contract-addr"), "ibc-port", wasmvmtypes.CosmosMsg{
			Custom: serializedSubjectMessage,
		})
		return err
	}

	JustBeforeEach(func() {
		serializedSubjectMessage, _ = json.Marshal(subjectMsg)
	})

	Context("invalid message", func() {
		When("message is empty", func() {
			It("returns an error", func() {
				Expect(subject()).To(MatchError(types.ErrWasmExecuteMessageNotValid))
			})
		})
	})

	Context("valid message", func() {
		jobID := "wohoo"
		typ := xchain.Type("abc")
		refID := xchain.ReferenceID("def")

		BeforeEach(func() {
			subjectMsg.Payload = []byte("hello")
		})

		var bm *xchainmocks.Bridge
		BeforeEach(func() {
			bm = xchainmocks.NewBridge(GinkgoT())
			f.schedulerKeeper.Chains[typ] = bm
		})

		BeforeEach(func() {
			bm.On("VerifyJob", mock.Anything, mock.Anything, mock.Anything, refID).Return(nil)
			var err error
			err = f.schedulerKeeper.AddNewJob(ctx, &types.Job{
				ID:    jobID,
				Owner: sdk.AccAddress("me"),
				Routing: types.Routing{
					ChainType:        typ,
					ChainReferenceID: refID,
				},
				Definition:          []byte("definition"),
				Payload:             []byte("payload"),
				IsPayloadModifiable: true,
			})
			Expect(err).To(BeNil())
		})

		When("job exists", func() {
			BeforeEach(func() {
				subjectMsg.JobID = jobID
			})
			BeforeEach(func() {
				bm.On("ExecuteJob", mock.Anything, mock.Anything).Return(uint64(1), nil)
			})

			It("schedules the job", func() {
				Expect(subject()).To(BeNil())
			})
		})

		When("job doesn't exist", func() {
			BeforeEach(func() {
				subjectMsg.JobID = "i don't exist"
			})

			It("returns an error when trying to schedule a job", func() {
				Expect(subject()).To(MatchError(types.ErrJobNotFound))
			})
		})
	})
})

func TestKeeper_UnmarshallJob(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		expected      keeper.ExecuteJobWasmEvent
		expectedError error
	}{
		{
			name:  "Happy Path",
			input: []byte("{\"job_id\":\"dca_test_job_2\",\"payload\":\"2WBzzwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAZmYQoclhTARc=\"}"),
			expected: keeper.ExecuteJobWasmEvent{
				JobID:   "dca_test_job_2",
				Payload: []byte("{\"hexPayload\":\"d96073cf00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000199984287258530117\"}"),
			},
		},
	}
	asserter := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testKeeper := keeper.Keeper{}
			actual, _ := testKeeper.UnmarshallJob(tt.input)
			asserter.Equal(string(tt.expected.Payload), string(actual.Payload))
		})
	}
}
