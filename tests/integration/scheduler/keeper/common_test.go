package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	xchainmocks "github.com/palomachain/paloma/internal/x-chain/mocks"
	"github.com/palomachain/paloma/x/scheduler/keeper"
	"github.com/palomachain/paloma/x/scheduler/types"
	"github.com/stretchr/testify/mock"
)

func TestScheduler(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Scheduler")
}

var _ = Describe("jobs!", func() {
	var k *keeper.Keeper
	var ctx sdk.Context
	BeforeEach(func() {
		t := GinkgoT()
		f := initFixture(t)
		k = &f.schedulerKeeper
		ctx = f.ctx.WithBlockHeight(5)
	})

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
							ChainType:        "evm",
							ChainReferenceID: "bla",
						},
						{
							Address:          []byte("bla"),
							ChainType:        "evm",
							ChainReferenceID: "bla",
						},
					},
					Blacklist: []*types.Runner{
						{
							ChainType:        "evm",
							ChainReferenceID: "bla",
						},
						{
							Address:          []byte("bla"),
							ChainType:        "evm",
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
			err := k.AddNewJob(ctx, &job)
			return err
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
				bm.On("VerifyJob", mock.Anything, mock.Anything, mock.Anything, xchain.ReferenceID("def")).Return(nil)
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

			When("scheduling a job", func() {
				var jobPayloadModifiable types.Job

				BeforeEach(func() {
					job = types.Job{
						ID:    "job1",
						Owner: sdk.AccAddress("bla"),
						Routing: types.Routing{
							ChainType:        "abc",
							ChainReferenceID: "def",
						},
						Definition: []byte("definition"),
						Payload:    []byte("payload"),
					}
					jobPayloadModifiable = types.Job{
						ID:    "job2",
						Owner: sdk.AccAddress("bla"),
						Routing: types.Routing{
							ChainType:        "abc",
							ChainReferenceID: "def",
						},
						Definition:          []byte("definition"),
						Payload:             []byte("payload"),
						IsPayloadModifiable: true,
					}
				})

				BeforeEach(func() {
					var err error
					err = k.AddNewJob(ctx, &job)
					Expect(err).To(BeNil())
					err = k.AddNewJob(ctx, &jobPayloadModifiable)
					Expect(err).To(BeNil())
				})

				It("schedules both jobs with the default payload", func() {
					bm.On("ExecuteJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uint64(10), nil)

					mid, err := k.ScheduleNow(ctx, job.GetID(), nil, nil, nil)
					Expect(mid).To(Equal(uint64(10)))
					Expect(err).To(BeNil())

					mid, err = k.ScheduleNow(ctx, jobPayloadModifiable.GetID(), nil, nil, nil)
					Expect(mid).To(Equal(uint64(10)))
					Expect(err).To(BeNil())
				})

				Context("modifying payload", func() {
					It("returns an error", func() {
						_, err := k.ScheduleNow(ctx, job.GetID(), []byte("new payload"), nil, nil)
						Expect(err).To(MatchError(types.ErrCannotModifyJobPayload))
					})
					It("doesn't return an error", func() {
						bm.On("ExecuteJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uint64(10), nil)
						mid, err := k.ScheduleNow(ctx, jobPayloadModifiable.GetID(), []byte("new payload"), nil, nil)
						Expect(mid).To(Equal(uint64(10)))
						Expect(err).To(BeNil())
					})
				})
			})
		})
	})
})
