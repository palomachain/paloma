package keeper_test

import (
	"testing"
	"time"

	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/testutil"
	"github.com/palomachain/paloma/x/valset/keeper"
)

func TestGenesisGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Valset")
}

var _ = Describe("jailing validators", func() {
	var ctx sdk.Context
	var a *fixture
	t := GinkgoT()
	BeforeEach(func() {
		a = initFixture(t)
		ctx = a.ctx.WithBlockHeight(5)
	})

	Context("with a non existing validator", func() {
		It("returns an error", func() {
			validators := testutil.GenValidators(1, 100)
			val, err := a.valsetKeeper.AddressCodec.StringToBytes(validators[0].GetOperator())
			Expect(err).To(BeNil())
			err1 := a.valsetKeeper.Jail(ctx, val, "i am bored")
			err1 = whoops.Errorf(err1.Error()).Format()
			Expect(err1).To(MatchError(keeper.ErrValidatorWithAddrNotFound))
		})
	})

	Context("with only a single validator", func() {
		var val sdk.ValAddress

		BeforeEach(func() {
			By("query existing validator")
			vals, err := a.stakingkeeper.GetAllValidators(ctx)
			Expect(err).To(BeNil())
			Expect(len(vals)).To(Equal(1))
			val, err = a.valsetKeeper.AddressCodec.StringToBytes(vals[0].GetOperator())
			Expect(err).To(BeNil())
		})

		It("returns an error that it cannot jail the validator", func() {
			err := a.valsetKeeper.Jail(ctx, val, "i am bored")
			Expect(err).To(MatchError(keeper.ErrCannotJailValidator))
		})
	})

	Context("with already a valid valset", func() {
		var val sdk.ValAddress

		BeforeEach(func() {
			By("add a whole lot of validators")
			validators := testutil.GenValidators(10, 100)
			for _, v := range validators {
				err := a.stakingkeeper.SetValidator(ctx, v)
				Expect(err).To(BeNil())
				err = a.stakingkeeper.SetValidatorByConsAddr(ctx, v)
				Expect(err).To(BeNil())
				addr, _ := v.GetConsAddr()
				si := types.NewValidatorSigningInfo(addr, 0, 0, time.Time{}, false, 0)
				err = a.SlashingKeeper.SetValidatorSigningInfo(ctx, addr, si)
				Expect(err).To(BeNil())
			}

			var err error
			val, err = a.valsetKeeper.AddressCodec.StringToBytes(validators[0].GetOperator())
			Expect(err).To(BeNil())
		})

		It("jails the given validator", func() {
			err := a.valsetKeeper.Jail(ctx, val, "i am bored")
			Expect(err).To(BeNil())
			isJailed, err := a.valsetKeeper.IsJailed(ctx, val)
			Expect(err).To(BeNil())
			Expect(isJailed).To(BeTrue())
		})
		When("jailing panics", func() {
			BeforeEach(func() {
				validators := testutil.GenValidators(1, 100)
				for _, v := range validators {
					err := a.stakingkeeper.SetValidator(ctx, v)
					Expect(err).To(BeNil())
				}
				var err error
				val, err = a.valsetKeeper.AddressCodec.StringToBytes(validators[0].GetOperator())
				Expect(err).To(BeNil())
			})
			It("returns the error if it's of type error or string", func() {
				err := a.valsetKeeper.Jail(ctx, val, "i am bored")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("jailing already jailed validator", func() {
			BeforeEach(func() {
				err := a.valsetKeeper.Jail(ctx, val, "i am bored")
				Expect(err).To(BeNil())
			})
			It("returns an error", func() {
				err := a.valsetKeeper.Jail(ctx, val, "i am bored")
				Expect(err).To(MatchError(keeper.ErrValidatorAlreadyJailed))
			})
		})

		Context("jailing validator with too much network stake", func() {
			BeforeEach(func() {
				By("change validator stake")
				validators := testutil.GenValidators(1, 100)
				err := a.stakingkeeper.SetValidator(ctx, validators[0])
				Expect(err).To(BeNil())
				err = a.stakingkeeper.SetValidatorByConsAddr(ctx, validators[0])
				Expect(err).To(BeNil())
				val, err = a.valsetKeeper.AddressCodec.StringToBytes(validators[0].GetOperator())
				Expect(err).To(BeNil())
			})
			It("returns an error", func() {
				err := a.valsetKeeper.Jail(ctx, val, "i am bored")
				Expect(err).To(MatchError(keeper.ErrCannotJailValidator))
			})
		})
	})
})
