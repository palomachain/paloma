package keeper_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/core/header"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/v2/testutil"
	"github.com/palomachain/paloma/v2/testutil/rand"
	utilkeeper "github.com/palomachain/paloma/v2/util/keeper"
	"github.com/palomachain/paloma/v2/x/paloma/keeper"
	palomatypes "github.com/palomachain/paloma/v2/x/paloma/types"
	valsettypes "github.com/palomachain/paloma/v2/x/valset/types"
)

func TestPalomaGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Paloma")
}

var _ = Describe("jailing validators with missing external chain infos", func() {
	var k *keeper.Keeper
	var ctx sdk.Context

	var vals []stakingtypes.Validator
	var f *fixture
	BeforeEach(func() {
		t := GinkgoT()
		f = initFixture(t)
		ctx = f.ctx.WithBlockHeight(5)
		k = &f.palomaKeeper
	})
	BeforeEach(func() {
		// Generate enough validators to ensure jailing is not prevented due to high
		// amount of stake in network.
		vals = testutil.GenValidators(10, 100)

		for _, val := range vals {
			err := f.stakingKeeper.SetValidator(ctx, val)
			Expect(err).To(BeNil())
			f.stakingKeeper.SetValidatorByConsAddr(ctx, val)
		}

		for _, val := range f.valsetKeeper.GetUnjailedValidators(ctx) {
			addr, err := val.GetConsAddr()
			Expect(err).To(BeNil())
			f.slashingKeeper.SetValidatorSigningInfo(ctx, addr, types.NewValidatorSigningInfo(addr, 0, 0, time.Time{}, false, 0))
		}
	})

	Context("test all cases", func() {
		BeforeEach(func() {
			// val[0] has everything
			// val[1] has only one chain
			// val[2] doesn't have anything
			// All other vals have everything
			valAddress, err := utilkeeper.ValAddressFromBech32(f.palomaKeeper.AddressCodec, vals[0].GetOperator())
			Expect(err).To(BeNil())
			err = f.valsetKeeper.AddExternalChainInfo(ctx, valAddress, []*valsettypes.ExternalChainInfo{
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
			valAddress2, err := utilkeeper.ValAddressFromBech32(f.palomaKeeper.AddressCodec, vals[1].GetOperator())
			Expect(err).To(BeNil())
			err = f.valsetKeeper.AddExternalChainInfo(ctx, valAddress2, []*valsettypes.ExternalChainInfo{
				{
					ChainType:        "evm",
					ChainReferenceID: "c1",
					Address:          "aaa",
					Pubkey:           []byte("ccc"),
				},
			})
			Expect(err).To(BeNil())
			for i, v := range vals[3:] {
				valAddress, err := utilkeeper.ValAddressFromBech32(f.palomaKeeper.AddressCodec, v.GetOperator())
				Expect(err).To(BeNil())
				err = f.valsetKeeper.AddExternalChainInfo(ctx, valAddress, []*valsettypes.ExternalChainInfo{
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
			err := f.evmKeeper.AddSupportForNewChain(ctx, "c1", 1, 123, "abc", big.NewInt(555))
			Expect(err).To(BeNil())
			err = f.evmKeeper.AddSupportForNewChain(ctx, "c2", 2, 123, "abc", big.NewInt(555))
			Expect(err).To(BeNil())
		})

		It("jails val[1] and val[2], but no val[0]", func() {
			By("validators are not jailed")
			valAddress1, err := utilkeeper.ValAddressFromBech32(f.palomaKeeper.AddressCodec, vals[0].GetOperator())
			Expect(err).To(BeNil())
			valAddress2, err := utilkeeper.ValAddressFromBech32(f.palomaKeeper.AddressCodec, vals[1].GetOperator())
			Expect(err).To(BeNil())
			valAddress3, err := utilkeeper.ValAddressFromBech32(f.palomaKeeper.AddressCodec, vals[2].GetOperator())
			Expect(err).To(BeNil())
			Expect(f.valsetKeeper.IsJailed(ctx, valAddress1)).To(BeFalse())
			Expect(f.valsetKeeper.IsJailed(ctx, valAddress2)).To(BeFalse())
			Expect(f.valsetKeeper.IsJailed(ctx, valAddress3)).To(BeFalse())

			By("jail validators with missing external chain infos")
			err = k.JailValidatorsWithMissingExternalChainInfos(ctx)
			Expect(err).To(BeNil())

			By("check for jailed validators")
			Expect(f.valsetKeeper.IsJailed(ctx, valAddress1)).To(BeFalse())
			Expect(f.valsetKeeper.IsJailed(ctx, valAddress2)).To(BeTrue())
			Expect(f.valsetKeeper.IsJailed(ctx, valAddress3)).To(BeTrue())
		})
	})
})

var _ = Describe("checking the chain version", func() {
	var k *keeper.Keeper
	var f *fixture
	var ctx sdk.Context
	BeforeEach(func() {
		t := GinkgoT()
		f = initFixture(t)
		ctx = f.ctx.WithHeaderInfo(header.Info{Height: 123})
		k = &f.palomaKeeper
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

				f.upgradeKeeper.ApplyUpgrade(ctx, upgradetypes.Plan{
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

var _ = Describe("updating params", func() {
	var ctx sdk.Context
	var k *keeper.Keeper
	var srv palomatypes.MsgServer
	var addr []string
	var f *fixture
	var authority string

	BeforeEach(func() {
		t := GinkgoT()
		f = initFixture(t)
		authority = authtypes.NewModuleAddress(govtypes.ModuleName).String()
		ctx = f.ctx.WithBlockHeight(5)
		k = &f.palomaKeeper
		srv = keeper.NewMsgServerImpl(*k)
	})

	When("with a valid request", func() {
		var msg *palomatypes.MsgUpdateParams

		BeforeEach(func() {
			addr = make([]string, 3)
			for i := range addr {
				addr[i] = rand.AccAddress().String()
			}
			msg = &palomatypes.MsgUpdateParams{
				Authority: authority,
				Params: palomatypes.Params{
					GasExemptAddresses: []string{
						addr[0],
						addr[1],
					},
				},
				Metadata: valsettypes.MsgMetadata{
					Creator: authority,
					Signers: []string{authority},
				},
			}
		})

		It("should not return an error", func() {
			_, err := srv.UpdateParams(ctx, msg)
			Expect(err).To(BeNil())
		})
		It("should not return a nil response", func() {
			res, _ := srv.UpdateParams(ctx, msg)
			Expect(res).ToNot(BeNil())
		})

		It("should update the parameters", func() {
			_, err := srv.UpdateParams(ctx, msg)
			Expect(err).To(BeNil())

			p := k.GetParams(ctx)
			Expect(p.GasExemptAddresses).To(HaveLen(2))
			Expect(p.GasExemptAddresses).To(ConsistOf(addr[0], addr[1]))
		})
	})
})
