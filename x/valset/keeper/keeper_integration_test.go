package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/app"
	testutil "github.com/palomachain/paloma/testutil"
	"github.com/palomachain/paloma/x/valset/keeper"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	priv1 = secp256k1.GenPrivKey()
	pk1   = priv1.PubKey()

	priv2 = secp256k1.GenPrivKey()
	pk2   = priv2.PubKey()
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
			By("add a single validator")
			validators := testutil.GenValidators(1, 100)
			a.StakingKeeper.SetValidator(ctx, validators[0])
			val = validators[0].GetOperator()
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
