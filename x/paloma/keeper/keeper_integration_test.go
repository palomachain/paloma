package keeper_test

import (
	"fmt"
	"math/big"
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/testutil"
	"github.com/palomachain/paloma/x/paloma/keeper"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
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
		ctx = a.NewContext(false).WithBlockHeight(5)
		k = &a.PalomaKeeper
	})

	BeforeEach(func() {
		// Generate enough validators to ensure jailing is not prevented due to high
		// amount of stake in network.
		vals = testutil.GenValidators(10, 100)

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
			// All other vals have everything
			valAddress, err := k.MustGetValAddr(vals[0].GetOperator())
			err = a.ValsetKeeper.AddExternalChainInfo(ctx, valAddress, []*valsettypes.ExternalChainInfo{
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
			valAddress2, err := k.MustGetValAddr(vals[1].GetOperator())
			err = a.ValsetKeeper.AddExternalChainInfo(ctx, valAddress2, []*valsettypes.ExternalChainInfo{
				{
					ChainType:        "evm",
					ChainReferenceID: "c1",
					Address:          "aaa",
					Pubkey:           []byte("ccc"),
				},
			})
			Expect(err).To(BeNil())
			for i, v := range vals[3:] {
				valAddress, err := k.MustGetValAddr(v.GetOperator())
				err = a.ValsetKeeper.AddExternalChainInfo(ctx, valAddress, []*valsettypes.ExternalChainInfo{
					{
						ChainType:        "evm",
						ChainReferenceID: "c1",
						Address:          fmt.Sprintf("abc%d", i+3),
						Pubkey:           []byte(fmt.Sprintf("abc%d", i+3)),
					},
					{
						ChainType:        "evm",
						ChainReferenceID: "c2",
						Address:          fmt.Sprintf("abcd%d", i+3),
						Pubkey:           []byte(fmt.Sprintf("abcd%d", i+3)),
					},
				})
				Expect(err).To(BeNil())
			}
		})

		BeforeEach(func() {
			err := a.EvmKeeper.AddSupportForNewChain(ctx, "c1", 1, 123, "abc", big.NewInt(555))
			Expect(err).To(BeNil())
			err = a.EvmKeeper.AddSupportForNewChain(ctx, "c2", 2, 123, "abc", big.NewInt(555))
			Expect(err).To(BeNil())
		})

		It("jails val[1] and val[2], but no val[0]", func() {
			By("validators are not jailed")
			valAddress1, err := k.MustGetValAddr(vals[0].GetOperator())
			valAddress2, err := k.MustGetValAddr(vals[1].GetOperator())
			valAddress3, err := k.MustGetValAddr(vals[2].GetOperator())
			Expect(a.ValsetKeeper.IsJailed(ctx, valAddress1)).To(BeFalse())
			Expect(a.ValsetKeeper.IsJailed(ctx, valAddress2)).To(BeFalse())
			Expect(a.ValsetKeeper.IsJailed(ctx, valAddress3)).To(BeFalse())

			By("jail validators with missing external chain infos")
			err = k.JailValidatorsWithMissingExternalChainInfos(ctx)
			Expect(err).To(BeNil())

			By("check for jailed validators")
			Expect(a.ValsetKeeper.IsJailed(ctx, valAddress1)).To(BeFalse())
			Expect(a.ValsetKeeper.IsJailed(ctx, valAddress2)).To(BeTrue())
			Expect(a.ValsetKeeper.IsJailed(ctx, valAddress3)).To(BeTrue())
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
		ctx = a.NewContext(false).WithBlockHeight(5)

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
