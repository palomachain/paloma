package keeper_test

import (
	"testing"
	"time"

	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	helper "github.com/palomachain/paloma/v2/tests/integration/helper"
	"github.com/palomachain/paloma/v2/testutil"
	"github.com/palomachain/paloma/v2/x/valset/keeper"
)

func TestGenesisGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Valset")
}

var _ = Describe("jailing validators", func() {
	var ctx sdk.Context
	var a *helper.Fixture
	t := GinkgoT()
	BeforeEach(func() {
		a = helper.InitFixture(t)
		ctx = a.Ctx.WithBlockHeight(5)

		validators := testutil.GenValidators(1, 100)
		err := a.StakingKeeper.SetValidator(a.Ctx, validators[0])
		Expect(err).To(BeNil())
		err = a.StakingKeeper.SetValidatorByPowerIndex(a.Ctx, validators[0])
		Expect(err).To(BeNil())
		err = a.StakingKeeper.SetValidatorByConsAddr(a.Ctx, validators[0])
		Expect(err).To(BeNil())
	})

	Context("with a non existing validator", func() {
		It("returns an error", func() {
			validators := testutil.GenValidators(1, 100)
			val, err := a.ValsetKeeper.AddressCodec.StringToBytes(validators[0].GetOperator())
			Expect(err).To(BeNil())
			err1 := a.ValsetKeeper.Jail(ctx, val, "i am bored")
			err1 = whoops.Errorf(err1.Error()).Format()
			Expect(err1).To(MatchError(keeper.ErrValidatorWithAddrNotFound))
		})
	})

	Context("with only a single validator", func() {
		var val sdk.ValAddress

		BeforeEach(func() {
			By("query existing validator")
			vals, err := a.StakingKeeper.GetAllValidators(ctx)
			Expect(err).To(BeNil())
			Expect(len(vals)).To(Equal(1))
			val, err = a.ValsetKeeper.AddressCodec.StringToBytes(vals[0].GetOperator())
			Expect(err).To(BeNil())
		})

		It("returns an error that it cannot jail the validator", func() {
			err := a.ValsetKeeper.Jail(ctx, val, "i am bored")
			Expect(err).To(MatchError(keeper.ErrCannotJailValidator))
		})
	})

	Context("with already a valid valset", func() {
		var val sdk.ValAddress

		BeforeEach(func() {
			By("add a whole lot of validators")
			validators := testutil.GenValidators(10, 100)
			for _, v := range validators {
				err := a.StakingKeeper.SetValidator(ctx, v)
				Expect(err).To(BeNil())
				err = a.StakingKeeper.SetValidatorByConsAddr(ctx, v)
				Expect(err).To(BeNil())
				addr, _ := v.GetConsAddr()
				si := types.NewValidatorSigningInfo(addr, 0, 0, time.Time{}, false, 0)
				err = a.SlashingKeeper.SetValidatorSigningInfo(ctx, addr, si)
				Expect(err).To(BeNil())
			}

			var err error
			val, err = a.ValsetKeeper.AddressCodec.StringToBytes(validators[0].GetOperator())
			Expect(err).To(BeNil())
		})

		It("jails the given validator", func() {
			err := a.ValsetKeeper.Jail(ctx, val, "i am bored")
			Expect(err).To(BeNil())
			isJailed, err := a.ValsetKeeper.IsJailed(ctx, val)
			Expect(err).To(BeNil())
			Expect(isJailed).To(BeTrue())
		})
		When("jailing panics", func() {
			BeforeEach(func() {
				validators := testutil.GenValidators(1, 100)
				for _, v := range validators {
					err := a.StakingKeeper.SetValidator(ctx, v)
					Expect(err).To(BeNil())
				}
				var err error
				val, err = a.ValsetKeeper.AddressCodec.StringToBytes(validators[0].GetOperator())
				Expect(err).To(BeNil())
			})
			It("returns the error if it's of type error or string", func() {
				err := a.ValsetKeeper.Jail(ctx, val, "i am bored")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("jailing already jailed validator", func() {
			BeforeEach(func() {
				err := a.ValsetKeeper.Jail(ctx, val, "i am bored")
				Expect(err).To(BeNil())
			})
			It("returns an error", func() {
				err := a.ValsetKeeper.Jail(ctx, val, "i am bored")
				Expect(err).To(MatchError(keeper.ErrValidatorAlreadyJailed))
			})
		})

		Context("jailing validator with too much network stake", func() {
			BeforeEach(func() {
				By("change validator stake")
				validators := testutil.GenValidators(1, 100)
				err := a.StakingKeeper.SetValidator(ctx, validators[0])
				Expect(err).To(BeNil())
				err = a.StakingKeeper.SetValidatorByConsAddr(ctx, validators[0])
				Expect(err).To(BeNil())
				val, err = a.ValsetKeeper.AddressCodec.StringToBytes(validators[0].GetOperator())
				Expect(err).To(BeNil())
			})
			It("returns an error", func() {
				err := a.ValsetKeeper.Jail(ctx, val, "i am bored")
				Expect(err).To(MatchError(keeper.ErrCannotJailValidator))
			})
		})
	})
})
