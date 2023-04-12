package keeper_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/testutil"
	"github.com/palomachain/paloma/x/valset/keeper"
)

func TestGenesisGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Valset")
}

var _ = Describe("jaling validators", func() {
	var a app.TestApp
	var ctx sdk.Context

	BeforeEach(func() {
		a = app.NewTestApp(GinkgoT(), false)
		ctx = a.NewContext(false, tmproto.Header{
			Height: 5,
		})
	})

	Context("with a non existing validator", func() {
		It("returns an error", func() {
			validators := testutil.GenValidators(1, 100)
			err := a.ValsetKeeper.Jail(ctx, validators[0].GetOperator(), "i am bored")
			Expect(err).To(MatchError(keeper.ErrValidatorWithAddrNotFound))
		})
	})

	Context("with only a single validator", func() {
		var val sdk.ValAddress

		BeforeEach(func() {
			By("query existing validator")
			vals := a.StakingKeeper.GetAllValidators(ctx)
			Expect(len(vals)).To(Equal(1))
			val = vals[0].GetOperator()
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
				a.StakingKeeper.SetValidator(ctx, v)
				a.StakingKeeper.SetValidatorByConsAddr(ctx, v)
			}
			val = validators[0].GetOperator()
		})

		It("jailes the given validator", func() {
			err := a.ValsetKeeper.Jail(ctx, val, "i am bored")
			Expect(err).To(BeNil())
			isJailed := a.ValsetKeeper.IsJailed(ctx, val)
			Expect(isJailed).To(BeTrue())
		})
		When("jailing panics", func() {
			BeforeEach(func() {
				validators := testutil.GenValidators(1, 100)
				for _, v := range validators {
					a.StakingKeeper.SetValidator(ctx, v)
				}
				val = validators[0].GetOperator()
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
	})
})
