package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/testutil"
	utilkeeper "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/palomath"
	"github.com/palomachain/paloma/x/metrix/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

func TestGenesisGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrix")
}

var _ = Describe("updating uptime", func() {
	var a *fixture
	var ctx sdk.Context
	var validators []stakingtypes.Validator
	var cons []sdk.ConsAddress
	var uptimes []math.LegacyDec
	BeforeEach(func() {
		t := GinkgoT()
		a = initFixture(t)
		ctx = a.ctx.WithBlockHeight(5)
		validators := testutil.GenValidators(3, 1000)
		for _, v := range validators {
			bz, err := utilkeeper.ValAddressFromBech32(a.metrixkeeper.AddressCodec, v.GetOperator())
			Expect(err).To(BeNil())
			r, err := a.metrixkeeper.GetValidatorMetrics(ctx, bz)
			Expect(r).To(BeNil())
			Expect(err).To(BeNil())

			a.stakingKeeper.SetValidator(ctx, v)
			a.stakingKeeper.SetValidatorByConsAddr(ctx, v)
		}

		cons = make([]sdk.ConsAddress, len(validators))
		for i, v := range validators {
			var err error
			cons[i], err = v.GetConsAddr()
			Expect(err).To(BeNil())
		}

		a.slashingKeeper.SetValidatorSigningInfo(ctx, cons[0], slashingtypes.NewValidatorSigningInfo(cons[0], 0, 0, time.Time{}, false, 1))
		a.slashingKeeper.SetValidatorSigningInfo(ctx, cons[1], slashingtypes.NewValidatorSigningInfo(cons[1], 0, 0, time.Time{}, false, 10))
		a.slashingKeeper.SetValidatorSigningInfo(ctx, cons[2], slashingtypes.NewValidatorSigningInfo(cons[2], 0, 0, time.Time{}, false, 42))
		a.metrixkeeper.UpdateUptime(ctx)
	})

	Context("with nonexisting validator", func() {
		It("creates new validator metrics", func() {
			uptimes = []math.LegacyDec{
				math.LegacyMustNewDecFromStr("0.99"),
				math.LegacyMustNewDecFromStr("0.9"),
				math.LegacyMustNewDecFromStr("0.58"),
			}
			for i, v := range validators {
				bz, err := utilkeeper.ValAddressFromBech32(a.metrixkeeper.AddressCodec, v.GetOperator())
				Expect(err).To(BeNil())
				r, err := a.metrixkeeper.GetValidatorMetrics(ctx, bz)
				Expect(r).To(Not(BeNil()))
				Expect(err).To(BeNil())
				Expect(r.Uptime).To(Equal(uptimes[i]))
			}
		})
	})
})

var _ = Describe("updating feature set", func() {
	var a *fixture
	var ctx sdk.Context
	var snapshot *valsettypes.Snapshot
	validators := testutil.GenValidators(3, 1000)
	BeforeEach(func() {
		t := GinkgoT()
		a = initFixture(t)
		ctx = a.ctx.WithBlockHeight(5)
		Address := make([]sdk.ValAddress, 3)
		for i := 0; i < 3; i++ {
			bz, _ := utilkeeper.ValAddressFromBech32(a.metrixkeeper.AddressCodec, validators[i].GetOperator())
			Address[i] = bz
		}
		snapshot = &valsettypes.Snapshot{
			Validators: []valsettypes.Validator{
				{
					State: valsettypes.ValidatorState_ACTIVE,
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: "1",
							Traits:           []string{valsettypes.PIGEON_TRAIT_MEV},
						},
						{
							ChainReferenceID: "2",
							Traits:           []string{valsettypes.PIGEON_TRAIT_MEV},
						},
						{
							ChainReferenceID: "3",
							Traits:           []string{valsettypes.PIGEON_TRAIT_MEV},
						},
					},
					Address: Address[0],
				},
				{
					State: valsettypes.ValidatorState_ACTIVE,
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: "1",
							Traits:           []string{valsettypes.PIGEON_TRAIT_MEV},
						},
						{
							ChainReferenceID: "2",
							Traits:           []string{valsettypes.PIGEON_TRAIT_MEV},
						},
						{
							ChainReferenceID: "3",
						},
					},
					Address: Address[1],
				},
				{
					State: valsettypes.ValidatorState_ACTIVE,
					ExternalChainInfos: []*valsettypes.ExternalChainInfo{
						{
							ChainReferenceID: "1",
							Traits:           []string{valsettypes.PIGEON_TRAIT_MEV},
						},
						{
							ChainReferenceID: "2",
						},
						{
							ChainReferenceID: "3",
						},
					},
					Address: Address[2],
				},
			},
		}
	})

	Context("with no existing records", func() {
		It("adds new records", func() {
			for _, v := range validators {
				bz, err := utilkeeper.ValAddressFromBech32(a.metrixkeeper.AddressCodec, v.GetOperator())
				Expect(err).To(BeNil())
				r, err := a.metrixkeeper.GetValidatorMetrics(ctx, bz)
				Expect(r).To(BeNil())
				Expect(err).To(BeNil())
				a.stakingKeeper.SetValidator(ctx, v)
				a.stakingKeeper.SetValidatorByConsAddr(ctx, v)
			}
			a.metrixkeeper.OnSnapshotBuilt(ctx, snapshot)

			featureSets := []math.LegacyDec{
				math.LegacyMustNewDecFromStr("1"),
				math.LegacyMustNewDecFromStr("0.66666667"),
				math.LegacyMustNewDecFromStr("0.33333333"),
			}
			for i, v := range validators {
				bz, err := utilkeeper.ValAddressFromBech32(a.metrixkeeper.AddressCodec, v.GetOperator())
				Expect(err).To(BeNil())
				r, err := a.metrixkeeper.GetValidatorMetrics(ctx, bz)
				Expect(r).To(Not(BeNil()))
				Expect(err).To(BeNil())
				Expect(r.FeatureSet).To(Equal(featureSets[i]))
			}
		})
	})

	Context("with existing records and no change", func() {
		It("doesn't change records", func() {
			for _, v := range validators {
				a.stakingKeeper.SetValidator(ctx, v)
				a.stakingKeeper.SetValidatorByConsAddr(ctx, v)
			}

			a.metrixkeeper.OnSnapshotBuilt(ctx, snapshot)
			a.metrixkeeper.OnSnapshotBuilt(ctx, snapshot)

			featureSets := []math.LegacyDec{
				math.LegacyMustNewDecFromStr("1"),
				math.LegacyMustNewDecFromStr("0.66666667"),
				math.LegacyMustNewDecFromStr("0.33333333"),
			}
			for i, v := range validators {
				bz, err := utilkeeper.ValAddressFromBech32(a.metrixkeeper.AddressCodec, v.GetOperator())
				Expect(err).To(BeNil())
				r, err := a.metrixkeeper.GetValidatorMetrics(ctx, bz)
				Expect(r).To(Not(BeNil()))
				Expect(err).To(BeNil())
				Expect(r.FeatureSet).To(Equal(featureSets[i]))
			}
		})
	})

	Context("with existing records and changed traits", func() {
		It("updates records", func() {
			for _, v := range validators {
				a.stakingKeeper.SetValidator(ctx, v)
				a.stakingKeeper.SetValidatorByConsAddr(ctx, v)
			}

			a.metrixkeeper.OnSnapshotBuilt(ctx, snapshot)
			snapshot.Validators[0].ExternalChainInfos[0].Traits = []string{}
			snapshot.Validators[1].ExternalChainInfos[1].Traits = []string{}
			snapshot.Validators[2].ExternalChainInfos[0].Traits = []string{valsettypes.PIGEON_TRAIT_MEV}
			snapshot.Validators[2].ExternalChainInfos[1].Traits = []string{valsettypes.PIGEON_TRAIT_MEV}
			a.metrixkeeper.OnSnapshotBuilt(ctx, snapshot)

			featureSets := []math.LegacyDec{
				math.LegacyMustNewDecFromStr("0.66666667"),
				math.LegacyMustNewDecFromStr("0.33333333"),
				math.LegacyMustNewDecFromStr("0.66666667"),
			}
			for i, v := range validators {
				bz, err := utilkeeper.ValAddressFromBech32(a.metrixkeeper.AddressCodec, v.GetOperator())
				Expect(err).To(BeNil())
				r, err := a.metrixkeeper.GetValidatorMetrics(ctx, bz)
				Expect(r).To(Not(BeNil()))
				Expect(err).To(BeNil())
				Expect(r.FeatureSet).To(Equal(featureSets[i]))
			}
		})
	})
})

var _ = Describe("handle message attested event", func() {
	var a *fixture
	var ctx sdk.Context
	valAddress := sdk.ValAddress("val-addr")

	BeforeEach(func() {
		t := GinkgoT()
		a = initFixture(t)
		ctx = a.ctx.WithBlockHeight(100)
	})

	Context("with no assigned at after handled at block height", func() {
		BeforeEach(func() {
			a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
				Assignee:              valAddress,
				AssignedAtBlockHeight: math.NewInt(20),
				HandledAtBlockHeight:  math.NewInt(5),
			})
		})
		It("logs an error and returns", func() {
			data, err := a.metrixkeeper.GetValidatorHistory(ctx, valAddress)
			Expect(err).To(BeNil())
			Expect(data).To(BeNil())
		})
		It("doesn't update the cache", func() {
			data, err := a.metrixkeeper.GetMessageNonceCache(ctx)
			Expect(err).To(BeNil())
			Expect(data).To((BeNil()))
		})
	})

	Context("with handled at block height after current block height", func() {
		BeforeEach(func() {
			a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
				Assignee:              valAddress,
				AssignedAtBlockHeight: math.NewInt(5),
				HandledAtBlockHeight:  math.NewInt(200),
			})
		})
		It("logs an error and returns", func() {
			data, err := a.metrixkeeper.GetValidatorHistory(ctx, valAddress)
			Expect(err).To(BeNil())
			Expect(data).To(BeNil())
		})
		It("doesn't update the cache", func() {
			data, err := a.metrixkeeper.GetMessageNonceCache(ctx)
			Expect(err).To(BeNil())
			Expect(data).To((BeNil()))
		})
	})

	Context("with no pre existing history", func() {
		BeforeEach(func() {
			a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
				AssignedAtBlockHeight:  math.NewInt(50),
				HandledAtBlockHeight:   math.NewInt(75),
				Assignee:               valAddress,
				MessageID:              42,
				WasRelayedSuccessfully: true,
			})
		})
		It("creates a new record with the event logged", func() {
			data, err := a.metrixkeeper.GetValidatorHistory(ctx, valAddress)
			Expect(err).To(BeNil())
			Expect(data).To(Not(BeNil()))
			Expect(data.Records).To(HaveLen(1))
			Expect(data.Records[0].Success).To(BeTrue())
			Expect(data.Records[0].MessageId).To(Equal(uint64(42)))
			Expect(data.Records[0].ExecutionSpeedInBlocks).To(Equal(uint64(25)))
		})

		It("updates the cache", func() {
			data, err := a.metrixkeeper.GetMessageNonceCache(ctx)
			Expect(err).To(BeNil())
			Expect(data).To(Not(BeNil()))
			Expect(data.MessageId).To(Equal(uint64(42)))
		})
	})

	Context("with existing history", func() {
		Context("and less than 100 messages on record", func() {
			BeforeEach(func() {
				a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
					AssignedAtBlockHeight:  math.NewInt(50),
					HandledAtBlockHeight:   math.NewInt(75),
					Assignee:               valAddress,
					MessageID:              42,
					WasRelayedSuccessfully: true,
				})
				a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
					AssignedAtBlockHeight:  math.NewInt(60),
					HandledAtBlockHeight:   math.NewInt(80),
					Assignee:               valAddress,
					MessageID:              43,
					WasRelayedSuccessfully: false,
				})
			})
			It("appends the new event to the history", func() {
				data, err := a.metrixkeeper.GetValidatorHistory(ctx, valAddress)
				Expect(err).To(BeNil())
				Expect(data).To(Not(BeNil()))
				Expect(data.Records).To(HaveLen(2))
				Expect(data.Records[1].Success).To(BeFalse())
				Expect(data.Records[1].MessageId).To(Equal(uint64(43)))
				Expect(data.Records[1].ExecutionSpeedInBlocks).To(Equal(uint64(20)))
			})
			It("updates the cache", func() {
				data, err := a.metrixkeeper.GetMessageNonceCache(ctx)
				Expect(err).To(BeNil())
				Expect(data).To(Not(BeNil()))
				Expect(data.MessageId).To(Equal(uint64(43)))
			})
			Context("and new message ID smaller than latest cache", func() {
				It("doesn't update the cache", func() {
					a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
						AssignedAtBlockHeight:  math.NewInt(50),
						HandledAtBlockHeight:   math.NewInt(75),
						Assignee:               valAddress,
						MessageID:              10,
						WasRelayedSuccessfully: true,
					})
					data, err := a.metrixkeeper.GetMessageNonceCache(ctx)
					Expect(err).To(BeNil())
					Expect(data).To(Not(BeNil()))
					Expect(data.MessageId).To(Equal(uint64(43)))
				})
			})
		})

		Context("with more than 100 messages on record", func() {
			BeforeEach(func() {
				ctx = a.ctx.WithBlockHeight(1000)
				for k := 0; k < 100; k++ {
					a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
						AssignedAtBlockHeight:  math.NewInt(int64(k + 10)),
						HandledAtBlockHeight:   math.NewInt(int64(k + 20)),
						Assignee:               valAddress,
						MessageID:              uint64(k + 1),
						WasRelayedSuccessfully: k%2 == 0,
					})
				}
				a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
					AssignedAtBlockHeight:  math.NewInt(111),
					HandledAtBlockHeight:   math.NewInt(121),
					Assignee:               valAddress,
					MessageID:              101,
					WasRelayedSuccessfully: true,
				})
			})
			It("rolls over the collected messages", func() {
				data, err := a.metrixkeeper.GetValidatorHistory(ctx, valAddress)
				Expect(err).To(BeNil())
				Expect(data).To(Not(BeNil()))
				Expect(data.Records).To(HaveLen(100))
				for k := 0; k < 100; k++ {
					Expect(data.Records[k].Success).To(Equal(k%2 == 1)) // Everything has shifted 1
					Expect(data.Records[k].MessageId).To(Equal(uint64(k + 2)))
					Expect(data.Records[k].ExecutionSpeedInBlocks).To(Equal(uint64(10)))
				}
			})
			It("updates the cache", func() {
				data, err := a.metrixkeeper.GetMessageNonceCache(ctx)
				Expect(err).To(BeNil())
				Expect(data).To(Not(BeNil()))
				Expect(data.MessageId).To(Equal(uint64(101)))
			})
		})
	})
})

var _ = Describe("purge historic relay data", func() {
	var a *fixture
	var ctx sdk.Context
	valAddress := []sdk.ValAddress{
		sdk.ValAddress("val-addr1"),
		sdk.ValAddress("val-addr2"),
		sdk.ValAddress("val-addr3"),
	}

	BeforeEach(func() {
		t := GinkgoT()
		a = initFixture(t)
		ctx = a.ctx.WithBlockHeight(20000)
	})

	Context("with no cache entry", func() {
		It("logs an error and returns", func() {
			a.metrixkeeper.PurgeRelayMetrics(ctx)
		})
	})

	Context("with cached message ID lower than scoring window", func() {
		BeforeEach(func() {
			a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
				Assignee:              valAddress[0],
				AssignedAtBlockHeight: math.NewInt(20),
				HandledAtBlockHeight:  math.NewInt(50),
				MessageID:             uint64(100),
			})
		})
		It("logs an error and returns", func() {
			a.metrixkeeper.PurgeRelayMetrics(ctx)
		})
	})

	Context("with cached message ID larger than scoring window", func() {
		Context("and validators with no messages outside scoring window", func() {
			BeforeEach(func() {
				for i, v := range valAddress {
					a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
						Assignee:              v,
						AssignedAtBlockHeight: math.NewInt(20),
						HandledAtBlockHeight:  math.NewInt(50),
						MessageID:             uint64(1500 + i),
					})
				}

				a.metrixkeeper.PurgeRelayMetrics(ctx)
			})
			It("does not purge any records", func() {
				for _, v := range valAddress {
					data, err := a.metrixkeeper.GetValidatorHistory(ctx, v)
					Expect(err).To(BeNil())
					Expect(data).To(Not(BeNil()))
					Expect(data.Records).To(HaveLen(1))
				}
			})
		})

		Context("and validators with some messages outside scoring window", func() {
			BeforeEach(func() {
				for i, v := range valAddress {
					a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
						Assignee:              v,
						AssignedAtBlockHeight: math.NewInt(20),
						HandledAtBlockHeight:  math.NewInt(50),
						MessageID:             uint64(500 + i),
					})
					a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
						Assignee:              v,
						AssignedAtBlockHeight: math.NewInt(20),
						HandledAtBlockHeight:  math.NewInt(50),
						MessageID:             uint64(600 + i),
					})
					a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
						Assignee:              v,
						AssignedAtBlockHeight: math.NewInt(20),
						HandledAtBlockHeight:  math.NewInt(50),
						MessageID:             uint64(1500 + i),
					})
				}
				a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
					Assignee:              valAddress[0],
					AssignedAtBlockHeight: math.NewInt(20),
					HandledAtBlockHeight:  math.NewInt(50),
					MessageID:             uint64(1550),
				})

				a.metrixkeeper.PurgeRelayMetrics(ctx)
			})
			It("purges any records outside scoring window", func() {
				data, err := a.metrixkeeper.GetValidatorHistory(ctx, valAddress[0])
				Expect(err).To(BeNil())
				Expect(data).To(Not(BeNil()))
				Expect(data.Records).To(HaveLen(3))
				Expect(data.Records[0].MessageId).To(Equal(uint64(600)))
				Expect(data.Records[1].MessageId).To(Equal(uint64(1500)))
				Expect(data.Records[2].MessageId).To(Equal(uint64(1550)))

				data, err = a.metrixkeeper.GetValidatorHistory(ctx, valAddress[1])
				Expect(err).To(BeNil())
				Expect(data).To(Not(BeNil()))
				Expect(data.Records).To(HaveLen(2))
				Expect(data.Records[0].MessageId).To(Equal(uint64(601)))
				Expect(data.Records[1].MessageId).To(Equal(uint64(1501)))

				data, err = a.metrixkeeper.GetValidatorHistory(ctx, valAddress[2])
				Expect(err).To(BeNil())
				Expect(data).To(Not(BeNil()))
				Expect(data.Records).To(HaveLen(2))
				Expect(data.Records[0].MessageId).To(Equal(uint64(602)))
				Expect(data.Records[1].MessageId).To(Equal(uint64(1502)))
			})
		})
	})
})

var _ = Describe("update relay metrics", func() {
	var a *fixture
	var ctx sdk.Context
	valAddress := []sdk.ValAddress{
		sdk.ValAddress("val-addr1"),
		sdk.ValAddress("val-addr2"),
		sdk.ValAddress("val-addr3"),
	}

	BeforeEach(func() {
		t := GinkgoT()
		a = initFixture(t)
		ctx = a.ctx.WithBlockHeight(2000)
		for i, v := range valAddress {
			for j := 0; j <= 100; j++ {
				a.metrixkeeper.OnConsensusMessageAttested(ctx, types.MessageAttestedEvent{
					Assignee: v,
					AssignedAtBlockHeight: func() math.Int {
						switch true {
						case i == 0:
							return math.NewInt(100)
						case i == 1:
							return math.NewInt(100 + int64(j))
						}
						if j%2 == 0 {
							return math.NewInt(100)
						}

						return math.NewInt(100 + int64(j))
					}(),
					HandledAtBlockHeight: math.NewInt(500),
					MessageID:            uint64(1500 + i + j),
					WasRelayedSuccessfully: func() bool {
						switch true {
						case i == 0:
							return true
						case i == 1:
							return false
						}
						return j%5 != 0
					}(),
				})
			}
			m, err := a.metrixkeeper.GetValidatorMetrics(ctx, v)
			Expect(m).To(BeNil())
			Expect(err).To(BeNil())
		}
	})

	Context("with no existing metrics on record", func() {
		metrics := make([]types.ValidatorMetrics, len(valAddress))
		BeforeEach(func() {
			a.metrixkeeper.UpdateRelayMetrics(ctx)
			for i, v := range valAddress {
				m, err := a.metrixkeeper.GetValidatorMetrics(ctx, v)
				Expect(m).To(Not(BeNil()))
				Expect(err).To(BeNil())
				metrics[i] = *m
			}
		})
		It("creates new records with the captured history", func() {
			Expect(metrics).To(HaveLen(3))
			for i, v := range valAddress {
				Expect(metrics[i].ValAddress).To(Equal(v.String()))
			}
		})
		It("creates calculates new metrics from historic records", func() {
			Expect(metrics).To(HaveLen(3))
			Expect(metrics[0].ExecutionTime).To(Equal(math.NewInt(400)))
			Expect(metrics[1].ExecutionTime).To(Equal(math.NewInt(349)))
			Expect(metrics[2].ExecutionTime).To(Equal(math.NewInt(399)))

			Expect(metrics[0].SuccessRate).To(Equal(palomath.LegacyDecFromFloat64(1)))
			Expect(metrics[1].SuccessRate).To(Equal(palomath.LegacyDecFromFloat64(0)))
			Expect(metrics[2].SuccessRate).To(Equal(palomath.LegacyDecFromFloat64(.9)))
		})
	})
})
