package keeper_test

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/app"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	xchainmocks "github.com/palomachain/paloma/internal/x-chain/mocks"
	"github.com/palomachain/paloma/x/scheduler/keeper"
	"github.com/palomachain/paloma/x/scheduler/types"
	"github.com/stretchr/testify/mock"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("wasm message handler", func() {
	var a app.TestApp
	var ctx sdk.Context

	BeforeEach(func() {
		a = app.NewTestApp(GinkgoT(), false)
		ctx = a.NewContext(false, tmproto.Header{
			Height: 5,
		})
	})

	var subjectMsg keeper.ExecuteJobWasmEvent
	var serializedSubjectMessage []byte

	subject := func() error {
		_, _, err := a.SchedulerKeeper.ExecuteWasmJobEventListener()(ctx, sdk.AccAddress("contract-addr"), "ibc-port", wasmvmtypes.CosmosMsg{
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
			a.SchedulerKeeper.Chains[typ] = bm
		})

		BeforeEach(func() {
			bm.On("VerifyJob", mock.Anything, mock.Anything, mock.Anything, refID).Return(nil)
			var err error
			_, err = a.SchedulerKeeper.AddNewJob(ctx, &types.Job{
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
				bm.On("ExecuteJob", mock.Anything, mock.Anything, mock.Anything, refID).Return(nil)
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
