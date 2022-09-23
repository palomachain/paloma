package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/app"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	xchainmocks "github.com/palomachain/paloma/internal/x-chain/mocks"
	"github.com/palomachain/paloma/x/scheduler/keeper"
	"github.com/palomachain/paloma/x/scheduler/types"
	"github.com/stretchr/testify/mock"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestScheduler(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Scheduler")
}

var _ = Describe("jailing validators with missing external chain infos", func() {

	var k *keeper.Keeper
	var ctx sdk.Context
	var a app.TestApp

	BeforeEach(func() {
		t := GinkgoT()
		a = app.NewTestApp(t, false)
		ctx = a.NewContext(false, tmproto.Header{
			Height: 5,
		})
		k = &a.SchedulerKeeper
	})

	_ = k
	_ = ctx

	Context("Verify the validate basics on the job", func() {
		DescribeTable("verify job's ValidateBasic function",
			func(job *types.Job, expectedErr error) {
				if expectedErr == nil {
					Expect(job.ValidateBasic()).To(BeNil())
				} else {
					Expect(job.ValidateBasic()).To(MatchError(expectedErr))
				}
			},
			Entry("with empty ID it returns an error", &types.Job{
				ID: "",
			}, types.ErrInvalid),
			Entry("uppercase letters it returns an error", &types.Job{
				ID: "aAa",
			}, types.ErrInvalid),
			Entry("with not allowed characters it returns an error", &types.Job{
				ID: "?",
			}, types.ErrInvalid),
			Entry("with word paloma it returns an error", &types.Job{
				ID: "oh-look-paloma-man",
			}, types.ErrInvalid),
			Entry("with word pigeon it returns an error", &types.Job{
				ID: "oh-look-pigeon-man",
			}, types.ErrInvalid),
			Entry("with a long string it returns an error", &types.Job{
				ID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			}, types.ErrInvalid),
			Entry("with an empty definition it returns an error", &types.Job{
				ID: "this-is-valid",
			}, types.ErrInvalid),
			Entry("with an empty payload it returns an error", &types.Job{
				ID:         "this-is-valid",
				Definition: []byte("not empty"),
			}, types.ErrInvalid),
			Entry("with an empty routing chain type it returns an error", &types.Job{
				ID:         "this-is-valid",
				Definition: []byte("not empty"),
				Payload:    []byte("not empty"),
				Routing:    types.Routing{},
			}, types.ErrInvalid),
			Entry("with an empty routing chain type it returns an error", &types.Job{
				ID:         "this-is-valid",
				Definition: []byte("not empty"),
				Payload:    []byte("not empty"),
				Routing: types.Routing{
					ChainType: "bla",
				},
			}, types.ErrInvalid),
			Entry("with invalid whitelist permission address it returns an error", &types.Job{
				ID:         "this-is-valid",
				Definition: []byte("not empty"),
				Payload:    []byte("not empty"),
				Routing: types.Routing{
					ChainType:        "bla",
					ChainReferenceID: "bla",
				},
				Permissions: types.Permissions{
					Whitelist: []*types.Runner{
						{
							Address: []byte("bla"),
						},
					},
				},
			}, types.ErrInvalid),
			Entry("with invalid whitelist permission chain reference id it returns an error", &types.Job{
				ID:         "this-is-valid",
				Definition: []byte("not empty"),
				Payload:    []byte("not empty"),
				Routing: types.Routing{
					ChainType:        "bla",
					ChainReferenceID: "bla",
				},
				Permissions: types.Permissions{
					Whitelist: []*types.Runner{
						{
							ChainReferenceID: "bla",
						},
					},
				},
			}, types.ErrInvalid),
			Entry("with invalid blacklist permission address it returns an error", &types.Job{
				ID:         "this-is-valid",
				Definition: []byte("not empty"),
				Payload:    []byte("not empty"),
				Routing: types.Routing{
					ChainType:        "bla",
					ChainReferenceID: "bla",
				},
				Permissions: types.Permissions{
					Blacklist: []*types.Runner{
						{
							Address: []byte("bla"),
						},
					},
				},
			}, types.ErrInvalid),
			Entry("with invalid blacklist permission chain reference id it returns an error", &types.Job{
				ID:         "this-is-valid",
				Definition: []byte("not empty"),
				Payload:    []byte("not empty"),
				Routing: types.Routing{
					ChainType:        "bla",
					ChainReferenceID: "bla",
				},
				Permissions: types.Permissions{
					Blacklist: []*types.Runner{
						{
							ChainReferenceID: "bla",
						},
					},
				},
			}, types.ErrInvalid),
			Entry("with valid job it returns no error", &types.Job{
				ID:         "this-is-valid",
				Definition: []byte("not empty"),
				Payload:    []byte("not empty"),
				Routing: types.Routing{
					ChainType:        "bla",
					ChainReferenceID: "bla",
				},
				Permissions: types.Permissions{
					Whitelist: []*types.Runner{
						{
							ChainType:        "EVM",
							ChainReferenceID: "bla",
						},
						{
							Address:          []byte("bla"),
							ChainType:        "EVM",
							ChainReferenceID: "bla",
						},
					},
					Blacklist: []*types.Runner{
						{
							ChainType:        "EVM",
							ChainReferenceID: "bla",
						},
						{
							Address:          []byte("bla"),
							ChainType:        "EVM",
							ChainReferenceID: "bla",
						},
					},
				},
			}, nil),
			Entry("with valid job it returns no error", &types.Job{
				ID:         "this-is-valid",
				Definition: []byte("not empty"),
				Payload:    []byte("not empty"),
				Routing: types.Routing{
					ChainType:        "bla",
					ChainReferenceID: "bla",
				},
				Permissions: types.Permissions{},
			}, nil),
		)
	})

	Context("adding a new job", func() {
		var job types.Job
		subject := func() error {
			return k.AddNewJob(ctx, &job)
		}
		Context("job is invalid", func() {
			When("owner is nil", func() {
				It("returns an error", func() {
					Expect(subject()).To(MatchError(types.ErrInvalid))
				})
			})
			When("job doesn't pass ValidateBasic", func() {
				BeforeEach(func() {
					job.Owner = sdk.AccAddress("bla bla")
				})
				It("returns an error", func() {
					Expect(subject()).To(MatchError(types.ErrInvalid))
				})
			})
		})
		Context("job is valid", func() {
			var bm *xchainmocks.Bridge
			BeforeEach(func() {
				bm = xchainmocks.NewBridge(GinkgoT())
				typ := xchain.Type("abc")
				bm.On("UnmarshalJob", mock.Anything, mock.Anything, xchain.ReferenceID("def")).Return(xchain.JobInfo{}, nil)
				k.Chains[typ] = bm
			})
			BeforeEach(func() {
				job = types.Job{
					ID:    "abcd",
					Owner: sdk.AccAddress("bla"),
					Routing: types.Routing{
						ChainType:        "abc",
						ChainReferenceID: "def",
					},
					Definition: []byte("definition"),
					Payload:    []byte("payload"),
				}
			})
			It("saves it", func() {
				By("it doesn't exist now")
				Expect(k.JobIDExists(ctx, job.ID)).To(BeFalse())
				Expect(subject()).To(BeNil())
				By("it exists now")
				Expect(k.JobIDExists(ctx, job.ID)).To(BeTrue())
				By("calling save again returns an error that job already exists")
				Expect(subject()).To(MatchError(types.ErrJobWithIDAlreadyExists))
			})
		})
	})
})
