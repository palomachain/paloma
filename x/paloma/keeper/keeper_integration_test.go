package keeper_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/testutil"
	"github.com/palomachain/paloma/x/paloma/keeper"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestPalomaGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Paloma")
}

var _ = Describe("jailing validators with missing external chain infos", func() {

	var k *keeper.Keeper
	var ctx sdk.Context
	var a app.TestApp

	var vals []stakingtypes.Validator

	BeforeEach(func() {
		t := GinkgoT()
		a = app.NewTestApp(t, false)
		ctx = a.NewContext(false, tmproto.Header{
			Height: 5,
		})
		k = &a.PalomaKeeper
	})

	BeforeEach(func() {
		vals = testutil.GenValidators(3, 100)

		for _, val := range vals {
			a.StakingKeeper.SetValidator(ctx, val)
			a.StakingKeeper.SetValidatorByConsAddr(ctx, val)
		}
	})

	Context("test all cases", func() {
		BeforeEach(func() {
			// val[0] has everything
			// val[1] has only one chain
			// val[2] doesn't have anything
			err := a.ValsetKeeper.AddExternalChainInfo(ctx, vals[0].GetOperator(), []*valsettypes.ExternalChainInfo{
				{
					ChainType:        "evm",
					ChainReferenceID: "c1",
					Address:          "abc",
					Pubkey:           []byte("abc"),
				},
				{
					ChainType:        "evm",
					ChainReferenceID: "c2",
					Address:          "abc1",
					Pubkey:           []byte("abc2"),
				},
			})
			Expect(err).To(BeNil())
			err = a.ValsetKeeper.AddExternalChainInfo(ctx, vals[1].GetOperator(), []*valsettypes.ExternalChainInfo{
				{
					ChainType:        "evm",
					ChainReferenceID: "c1",
					Address:          "aaa",
					Pubkey:           []byte("ccc"),
				},
			})
			Expect(err).To(BeNil())
		})

		BeforeEach(func() {
			err := a.EvmKeeper.AddSupportForNewChain(ctx, "c1", 1, 123, "abc", big.NewInt(555))
			Expect(err).To(BeNil())
			err = a.EvmKeeper.AddSupportForNewChain(ctx, "c2", 2, 123, "abc", big.NewInt(555))
			Expect(err).To(BeNil())
		})

		It("jails val[1] and val[2], but no val[0]", func() {
			By("validators are not jailed")
			Expect(a.ValsetKeeper.IsJailed(ctx, vals[0].GetOperator())).To(BeFalse())
			Expect(a.ValsetKeeper.IsJailed(ctx, vals[1].GetOperator())).To(BeFalse())
			Expect(a.ValsetKeeper.IsJailed(ctx, vals[2].GetOperator())).To(BeFalse())

			By("jail validators with missing external chain infos")
			err := k.JailValidatorsWithMissingExternalChainInfos(ctx)
			Expect(err).To(BeNil())

			By("check for jailed validators")
			Expect(a.ValsetKeeper.IsJailed(ctx, vals[0].GetOperator())).To(BeFalse())
			Expect(a.ValsetKeeper.IsJailed(ctx, vals[1].GetOperator())).To(BeTrue())
			Expect(a.ValsetKeeper.IsJailed(ctx, vals[2].GetOperator())).To(BeTrue())
		})
	})
})

var _ = Describe("checking the chain version", func() {

	var k *keeper.Keeper
	var ctx sdk.Context
	var a app.TestApp

	BeforeEach(func() {
		t := GinkgoT()
		a = app.NewTestApp(t, false)
		ctx = a.NewContext(false, tmproto.Header{
			Height: 5,
		})
		k = &a.PalomaKeeper
	})

	Context("without any previous chain version", func() {
		It("does nothing", func() {
			Expect(func() {
				k.CheckChainVersion(ctx)
			}).NotTo(Panic())
		})
	})

	Context("with a proposed chain version", func() {

		DescribeTable("it checks the version",
			func(appVersion string, panics bool) {
				k.AppVersion = appVersion
				a.UpgradeKeeper.ApplyUpgrade(ctx, upgradetypes.Plan{
					Name:   "v5.1.6",
					Height: 123,
				})
				exp := Expect(func() {
					k.CheckChainVersion(ctx)
				})
				if panics {
					exp.To(Panic())
				} else {
					exp.NotTo(Panic())
				}
			},
			Entry("an exact version", "v5.1.6", false),
			Entry("higher patch version", "v5.1.7", false),

			Entry("lower patch version", "v5.1.4", true),
			Entry("lower minor version", "v5.0.10", true),
			Entry("lower major version", "v4.10.10", true),

			Entry("higher major version", "v7.1.6", true),
			Entry("higher minor version", "v5.2.6", true),
		)
	})
})
