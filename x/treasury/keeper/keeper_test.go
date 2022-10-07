package keeper_test

// import (
// 	"fmt"
// 	"testing"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"
// 	"github.com/palomachain/paloma/app"
// 	xchain "github.com/palomachain/paloma/internal/x-chain"
// 	schedulertypes "github.com/palomachain/paloma/x/scheduler/types"
// 	"github.com/palomachain/paloma/x/treasury/keeper"
// 	"github.com/palomachain/paloma/x/treasury/types"
// 	"github.com/stretchr/testify/mock"

// 	xchainmocks "github.com/palomachain/paloma/internal/x-chain/mocks"
// 	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
// )

// func TestTreasuryKeeper(t *testing.T) {
// 	RegisterFailHandler(Fail)

// 	RunSpecs(t, "Treasury Keeper")
// }

// var _ = Describe("adding funds", func() {

// 	var k *keeper.Keeper
// 	var ctx sdk.Context
// 	var a app.TestApp

// 	BeforeEach(func() {
// 		t := GinkgoT()
// 		a = app.NewTestApp(t, false)
// 		ctx = a.NewContext(false, tmproto.Header{
// 			Height: 5,
// 		})
// 		k = &a.TreasuryKeeper
// 	})

// 	Context("sending coins to address adds it to the bank", func() {
// 		var bm *xchainmocks.Bridge
// 		var fc *xchainmocks.FundCollecter
// 		typ := xchain.Type("abc")
// 		refID := xchain.ReferenceID("def")
// 		BeforeEach(func() {
// 			fc = xchainmocks.NewFundCollecter(GinkgoT())
// 			k.Chains = append(k.Chains, fc)
// 		})
// 		BeforeEach(func() {
// 			bm = xchainmocks.NewBridge(GinkgoT())
// 			bm.On("VerifyJob", mock.Anything, mock.Anything, mock.Anything, refID).Return(nil)
// 			a.SchedulerKeeper.Chains[typ] = bm
// 		})
// 		var jobAddr sdk.AccAddress
// 		jobID := "wohoo"
// 		BeforeEach(func() {
// 			var err error
// 			jobAddr, err = a.SchedulerKeeper.AddNewJob(ctx, &schedulertypes.Job{
// 				ID:    jobID,
// 				Owner: sdk.AccAddress("me"),
// 				Routing: schedulertypes.Routing{
// 					ChainType:        typ,
// 					ChainReferenceID: refID,
// 				},
// 				Definition: []byte("definition"),
// 				Payload:    []byte("payload"),
// 			})
// 			Expect(err).To(BeNil())
// 		})

// 		It("changes the balance", func() {
// 			Expect(a.BankKeeper.GetAllBalances(ctx, jobAddr)).To(HaveLen(0))
// 			err := k.AddFunds(ctx, typ, refID, jobID, sdk.NewInt(55))
// 			Expect(err).To(BeNil())
// 			Expect(a.BankKeeper.GetAllBalances(ctx, jobAddr)).To(
// 				And(
// 					HaveLen(1),
// 					Equal(
// 						sdk.NewCoins(
// 							sdk.NewCoin(fmt.Sprintf("%s/%s", typ, refID), sdk.NewInt(55)),
// 						),
// 					),
// 				),
// 			)
// 		})

// 		It("returns an error when trying to add zero funds", func() {
// 			err := k.AddFunds(ctx, typ, refID, jobID, sdk.NewInt(0))
// 			Expect(err).To(MatchError(types.ErrCannotAddZeroFunds))
// 		})

// 		It("returns an error when trying to add negative funds", func() {
// 			err := k.AddFunds(ctx, typ, refID, jobID, sdk.NewInt(-55))
// 			Expect(err).To(MatchError(types.ErrCannotAddNegativeFunds))
// 		})

// 		When("job's address does not exist", func() {
// 			It("returns an error", func() {
// 				err := k.AddFunds(ctx, typ, refID, "i don't exist", sdk.NewInt(55))
// 				Expect(err).NotTo(BeNil())
// 			})
// 		})
// 	})
// })
